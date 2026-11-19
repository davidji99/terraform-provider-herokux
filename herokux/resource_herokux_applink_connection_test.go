package herokux

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/applink"
	"github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	linkApp      = "6588f886-4e0a-463d-adbc-efeedd0e0001"
	linkOtherApp = "6588f886-4e0a-463d-adbc-efeedd0e0003"
	linkAddon    = "97d4b501-03da-4eec-955c-efeedd0e0002"
	linkName     = "production_org"
	linkID       = linkApp + ":" + linkAddon + ":" + linkName
)

type appLinkTransport func(*http.Request) (*http.Response, error)

func (f appLinkTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type appLinkFixture struct {
	sync.Mutex
	meta                      *Config
	key                       string
	apiURL                    string
	serviceStatus             int
	platformStatus            int
	redirect                  bool
	serviceType               string
	remoteID, status, orgType string
	afterCreate               string
	creates, deletes          int
}

func newAppLinkFixture(t *testing.T) *appLinkFixture {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	f := &appLinkFixture{key: string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})),
		apiURL: "https://applink-api.heroku.com/addons/" + linkAddon, serviceType: "heroku-applink", orgType: "SalesforceOrg", afterCreate: "connected"}
	platform := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.Lock()
		defer f.Unlock()
		assert.Equal(t, "Bearer platform-test", r.Header.Get("Authorization"))
		assert.Empty(t, r.Header.Get("X-Addon-SSO"))
		w.Header().Set("Content-Type", "application/json")
		if f.platformStatus != 0 {
			http.Error(w, `{"id":"record_not_found"}`, f.platformStatus)
			return
		}
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/apps/"), "/")
		if len(parts) < 2 || (parts[0] != linkApp && parts[0] != linkOtherApp) {
			http.NotFound(w, r)
			return
		}
		app := parts[0]
		switch strings.Join(parts[1:], "/") {
		case "addons/" + linkAddon:
			_, _ = fmt.Fprintf(w, `{"id":%q,"addon_service":{"name":%q}}`, linkAddon, f.serviceType)
		case "addon-attachments":
			if r.Header.Get("Range") == "" {
				w.Header().Set("Next-Range", "id ..next")
				w.WriteHeader(206)
				_, _ = w.Write([]byte("[]"))
				return
			}
			_, _ = fmt.Fprintf(w, `[{"name":"MY_ALIAS","addon":{"id":%q},"app":{"id":%q}}]`, linkAddon, app)
		case "config-vars":
			_ = json.NewEncoder(w).Encode(map[string]string{"MY_ALIAS_API_URL": f.apiURL, "MY_ALIAS_TOKEN": "attachment-test-token", "UNRELATED": "unrelated-config-secret"})
		case "addons/" + linkAddon + "/sso":
			assert.Equal(t, "application/vnd.heroku+json; version=3.sdk", r.Header.Get("Accept"))
			_, _ = w.Write([]byte(`{"token":"sso-test"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(platform.Close)
	service := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.Lock()
		defer f.Unlock()
		assert.Equal(t, "Bearer attachment-test-token", r.Header.Get("Authorization"))
		assert.Contains(t, []string{linkApp, linkOtherApp}, r.Header.Get("X-App-UUID"))
		assert.Equal(t, base64.StdEncoding.EncodeToString([]byte(`{"token":"sso-test"}`)), r.Header.Get("X-Addon-SSO"))
		assert.Empty(t, r.Header.Get("X-Platform-Only"))
		w.Header().Set("Content-Type", "application/json")
		if f.redirect {
			http.Redirect(w, r, platform.URL+"/credential-trap", http.StatusTemporaryRedirect)
			return
		}
		if f.serviceStatus != 0 {
			w.WriteHeader(f.serviceStatus)
			_ = json.NewEncoder(w).Encode(map[string]string{"id": "forbidden", "message": f.key})
			return
		}
		writeConn := func() {
			_ = json.NewEncoder(w).Encode(map[string]any{"id": f.remoteID, "status": f.status, "addon_id": linkAddon, "org": map[string]string{"type": f.orgType, "connection_name": linkName, "id": "00D000000000111", "instance_url": "https://mock.my.salesforce.com"}})
		}
		if r.Method == "POST" {
			assert.Equal(t, "/addons/"+linkAddon+"/connections/salesforce/jwt", r.URL.Path)
			var b map[string]string
			assert.NoError(t, json.NewDecoder(r.Body).Decode(&b))
			assert.True(t, b["jwt_private_key"] == f.key, "wrong JWT private key")
			assert.NotEmpty(t, b["alias"])
			assert.Equal(t, linkName, b["connection_name"])
			assert.NotEmpty(t, b["username"])
			assert.NotEmpty(t, b["client_id"])
			assert.Equal(t, appLinkLoginURL, b["login_url"])
			if f.remoteID != "" && f.status != "disconnected" {
				http.Error(w, "already connected", 409)
				return
			}
			f.creates++
			f.remoteID = fmt.Sprintf("remote-%d", f.creates)
			f.status = "connecting"
			w.WriteHeader(201)
			writeConn()
			f.status = f.afterCreate
			return
		}
		assert.Equal(t, "/addons/"+linkAddon+"/connections/"+linkName, r.URL.Path)
		if f.remoteID == "" {
			w.WriteHeader(404)
			_, _ = w.Write([]byte(`{"id":"record_not_found"}`))
			return
		}
		switch r.Method {
		case "GET":
			writeConn()
			if f.status == "disconnecting" {
				f.status = "disconnected"
			}
		case "DELETE":
			f.deletes++
			f.status = "disconnecting"
			w.WriteHeader(204)
		default:
			http.Error(w, "unexpected method", 400)
		}
	}))
	t.Cleanup(service.Close)
	serviceURL, err := url.Parse(service.URL)
	require.NoError(t, err)
	transport := appLinkTransport(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host == "applink-api.heroku.com" && r.URL.Scheme == "https" {
			r = r.Clone(r.Context())
			r.URL.Host = serviceURL.Host
			return service.Client().Transport.RoundTrip(r)
		}
		if r.URL.Scheme+"://"+r.URL.Host == platform.URL {
			return platform.Client().Transport.RoundTrip(r)
		}
		return nil, errors.New("unexpected outbound host")
	})
	client := applink.New(&config.Config{PlatformBaseURL: platform.URL, APIToken: "platform-test", AcceptHeader: api.DefaultAcceptHeader, UserAgent: "herokux-test", CustomHTTPHeaders: map[string]string{"X-Platform-Only": "test"}}, transport)
	f.meta = &Config{API: &api.Client{AppLink: client}}
	return f
}

func TestAppLinkIdentityAndErrors(t *testing.T) {
	f := newAppLinkFixture(t)
	f.remoteID, f.status = "original", "connected"
	res := resourceHerokuxAppLinkConnection()
	d := schema.TestResourceDataRaw(t, res.Schema, nil)
	d.SetId(linkID)
	_, err := res.Importer.StateContext(t.Context(), d, f.meta)
	require.NoError(t, err)
	require.Equal(t, "original", d.Get("connection_id"))
	checkError := func() {
		t.Helper()
		ds := res.ReadContext(t.Context(), d, f.meta)
		require.True(t, ds.HasError())
		require.False(t, strings.Contains(fmt.Sprint(ds), "PRIVATE KEY"))
		require.Equal(t, linkID, d.Id())
	}
	for _, status := range []int{401, 403, 404, 500} {
		f.serviceStatus = status
		checkError()
	}
	f.serviceStatus = 0
	f.platformStatus = 404
	checkError()
	f.platformStatus = 0
	f.remoteID = "new-org-same-name"
	checkError()
	require.True(t, res.DeleteContext(t.Context(), d, f.meta).HasError())
	require.Zero(t, f.deletes)
	f.remoteID = "original"
	f.orgType = "DataCloudOrg"
	checkError()
	_, err = res.Importer.StateContext(t.Context(), d, f.meta)
	require.Error(t, err)
	f.orgType = "SalesforceOrg"
	f.remoteID = ""
	require.Empty(t, res.ReadContext(t.Context(), d, f.meta))
	require.Empty(t, d.Id())
}

func TestAppLinkTrustedEndpoint(t *testing.T) {
	f := newAppLinkFixture(t)
	for _, invalid := range []string{"http://applink-api.heroku.com", "https://applink-api.heroku.com.evil.example", "https://user@applink-api.heroku.com", "https://applink-api.heroku.com:444", "https://untrusted.example"} {
		f.apiURL = invalid
		_, err := f.meta.API.AppLink.Session(t.Context(), linkApp, linkAddon)
		require.ErrorContains(t, err, "trusted AppLink origin")
	}
	f.apiURL = "https://applink-api.heroku.com"
	f.serviceType = "other"
	_, err := f.meta.API.AppLink.Session(t.Context(), linkApp, linkAddon)
	require.Error(t, err)
	f.serviceType = "heroku-applink"
	f.redirect = true
	s, err := f.meta.API.AppLink.Session(t.Context(), linkApp, linkAddon)
	require.NoError(t, err)
	_, err = s.Create(t.Context(), linkName, "consumer", "user", appLinkLoginURL, f.key)
	require.ErrorContains(t, err, "307")
	require.Zero(t, f.creates)
}
