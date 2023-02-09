package ir

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTagGetTags(t *testing.T) {
	tests := []struct {
		name     string
		tag      Tag
		wantTags string
	}{
		{
			"Empty",
			Tag{},
			``,
		},
		{
			"Only JSON",
			Tag{
				JSON: "field",
			},
			`json:"field"`,
		},
		{
			"Only ExtraTags",
			Tag{
				ExtraTags: map[string]string{
					"gorm": "primaryKey",
				},
			},
			`gorm:"primaryKey"`,
		},
		{
			"JSON+ExtraTags",
			Tag{
				JSON: "field",
				ExtraTags: map[string]string{
					"gorm": "primaryKey",
				},
			},
			`json:"field" gorm:"primaryKey"`,
		},
		{
			"JSON+ExtraTags2",
			Tag{
				JSON: "field",
				ExtraTags: map[string]string{
					"gorm":  "primaryKey",
					"valid": "customIdValidator",
				},
			},
			`json:"field" gorm:"primaryKey" valid:"customIdValidator"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags := tt.tag.GetTags()
			require.Equal(t, tt.wantTags, tags)
		})
	}
}
