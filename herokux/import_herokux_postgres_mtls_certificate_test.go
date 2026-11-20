package herokux

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccHerokuxPostgresMTLSCertificate_importBasic(t *testing.T) {
	dbName := testAccConfig.GetDBNameorSkip(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPostgresMTLSCertificate_basic(dbName),
			},
			{
				ResourceName:      "herokux_postgres_mtls_certificate.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestPostgresMTLSCertificateImportMixedCase(t *testing.T) {
	const id = "c9c92d8a-eeb1-4590-8286-945d09aaee2c"
	const other = "c9c92d8a-eeb1-4590-8286-945d09aaee2d"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"id":%q,"status":"ready"}`, id)
	}))
	defer server.Close()
	client, err := api.New(config.PostgresBaseURL(server.URL))
	require.NoError(t, err)
	r := resourceHerokuxPostgresMTLSCertificate()
	d := r.Data(nil)
	d.SetId("db:" + strings.ToUpper(id))
	meta := &Config{API: client}
	entries, err := r.Importer.StateContext(t.Context(), d, meta)
	require.NoError(t, err)
	require.Equal(t, "db:"+id, entries[0].Id())
	require.Equal(t, id, entries[0].Get("cert_id"))
	d.SetId("db:" + other)
	_, err = r.Importer.StateContext(t.Context(), d, meta)
	require.Error(t, err)
}
