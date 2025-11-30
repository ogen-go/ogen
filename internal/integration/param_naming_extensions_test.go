package integration

import (
	"testing"

	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/integration/test_param_naming_extensions"
)

func TestParamNamingExtensions(t *testing.T) {
	// This test verifies that x-ogen-name extension works for parameters.
	// The spec defines parameters with custom Go field names:
	// - itemid -> ItemID
	// - pickuptype -> PickupType
	// - x-custom-header -> CustomHeader
	a := require.New(t)

	// Create params using the custom field names
	params := api.GetItemParams{
		ItemID: "test-item-id",
	}

	// Verify the field types exist with the custom names
	a.Equal("test-item-id", params.ItemID)

	// Test optional fields
	params.PickupType.SetTo("delivery")
	a.True(params.PickupType.IsSet())
	a.Equal("delivery", params.PickupType.Value)

	params.CustomHeader.SetTo("custom-value")
	a.True(params.CustomHeader.IsSet())
	a.Equal("custom-value", params.CustomHeader.Value)
}
