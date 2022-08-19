package gen

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/ogen-go/ogen/openapi"
)

func Test_filterMostSpecific(t *testing.T) {
	sch := new(openapi.MediaType)
	list := func(input ...string) (r map[string]*openapi.MediaType) {
		r = map[string]*openapi.MediaType{}
		for _, value := range input {
			r[value] = sch
		}
		return r
	}

	tests := []struct {
		contents map[string]*openapi.MediaType
		expected map[string]*openapi.MediaType
		wantErr  bool
	}{
		{list("text/html", "text/*"), list("text/html"), false},
		{list("text/*"), list("text/*"), false},
		{list("*/*"), list("*/*"), false},
		{list("text/html", "*/*"), list("text/html"), false},
		{list("text/*", "*/*"), list("text/*"), false},
		{list("text/html", "text/*", "*/*", "*"), list("text/html"), false},
		{list("*/json", "application/json"), list("application/json"), false},
		{list("*/json", "application/json", "text/json"), list("application/json", "text/json"), false},
		{list("*/json", "*application/json"), list("*application/json"), false},
		{list("application/*", "text/html"), list("application/*", "text/html"), false},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)
			core, logs := observer.New(zapcore.DebugLevel)

			// Make a copy of the testdata to avoid modifying it.
			contents := map[string]*openapi.MediaType{}
			for k, v := range tt.contents {
				contents[k] = v
			}

			err := filterMostSpecific(contents, zap.New(core))
			if tt.wantErr {
				a.Error(err)
			} else {
				a.NoError(err)
				a.Equal(tt.expected, contents)
				// Ensure that there is a log message for every filtered media type.
				if before, after := len(tt.contents), len(tt.expected); after < before {
					entries := logs.FilterMessage("Filter common content type").All()
					a.NotEmpty(entries)
					a.Len(entries, before-after)
				}
			}
		})
	}
}
