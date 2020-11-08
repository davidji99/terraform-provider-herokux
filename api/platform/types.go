package platform

type WebhookLevel string

var WebhookLevels = struct {
	Sync   WebhookLevel
	Notify WebhookLevel
}{
	Sync:   "sync",
	Notify: "notify",
}

func (w WebhookLevel) ToString() string {
	return string(w)
}

type WebhookEventType string

var WebhookEventTypes = struct {
	APIAddonAttachment WebhookEventType
	APIAddon           WebhookEventType
	APIApp             WebhookEventType
	APIBuild           WebhookEventType
	APICollaborator    WebhookEventType
	APIDomain          WebhookEventType
	APIDyno            WebhookEventType
	APIFormation       WebhookEventType
	APIRelease         WebhookEventType
	APISniEndpoint     WebhookEventType
	APISslEndpoint     WebhookEventType
}{
	APIAddonAttachment: "api:addon-attachment",
	APIAddon:           "api:addon",
	APIApp:             "api:app",
	APIBuild:           "api:build",
	APICollaborator:    "api:collaborator",
	APIDomain:          "api:domain",
	APIDyno:            "api:dyno",
	APIFormation:       "api:formation",
	APIRelease:         "api:release",
	APISniEndpoint:     "api:sni-endpoint",
	APISslEndpoint:     "api:ssl-endpoint",
}

func (w WebhookEventType) ToString() string {
	return string(w)
}
