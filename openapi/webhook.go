package openapi

import "github.com/ogen-go/ogen/location"

// Webhook is an OpenAPI Webhook.
type Webhook struct {
	// Name of the webhook.
	Name string
	// Operations of the webhook's Path Item.
	Operations []*Operation

	location.Pointer `json:"-" yaml:"-"`
}
