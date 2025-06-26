package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/_testdata/testtypes"
	api "github.com/ogen-go/ogen/internal/integration/test_type_extension"
)

type testTypeHandler struct {
	optionalParams api.OptionalParams
	requiredParams api.RequiredParams
}

func (h *testTypeHandler) Optional(ctx context.Context, params api.OptionalParams) (*api.OptionalOK, error) {
	h.optionalParams = params
	return nil, assert.AnError
}

func (h *testTypeHandler) Required(ctx context.Context, params api.RequiredParams) (*api.RequiredOK, error) {
	h.requiredParams = params
	return nil, assert.AnError
}

type testTypeClient struct {
	request *http.Request
}

func (h *testTypeClient) Do(r *http.Request) (*http.Response, error) {
	h.request = r
	return nil, assert.AnError
}

func TestTypeExtension_Params(t *testing.T) {
	th := &testTypeHandler{}
	s, _ := api.NewServer(th)

	tc := &testTypeClient{}
	c, err := api.NewClient("", api.WithClient(tc))
	require.NoError(t, err)

	params := url.Values{
		"ogenString":   []string{"1"},
		"ogenNumber":   []string{"2"},
		"jsonString":   []string{"3"},
		"jsonNumber":   []string{"4"},
		"textString":   []string{"5"},
		"textNumber":   []string{"6"},
		"string":       []string{"7"},
		"number":       []string{"8"},
		"alias":        []string{"9"},
		"pointer":      []string{"10"},
		"aliasPointer": []string{"11"},
		"array":        []string{"1", "2", "3"},
	}

	t.Run("Required", func(t *testing.T) {
		expected := api.RequiredParams{
			OgenString:   testtypes.StringOgen{Value: "1"},
			OgenNumber:   testtypes.NumberOgen{Value: 2},
			JsonString:   testtypes.StringJSON{Value: "3"},
			JsonNumber:   testtypes.NumberJSON{Value: 4},
			TextString:   testtypes.Text{Value: "5"},
			TextNumber:   testtypes.Text{Value: "6"},
			String:       testtypes.String("7"),
			Number:       testtypes.Number(8),
			Alias:        api.Alias{Value: "9"},
			Pointer:      testtypes.NumberOgen{Value: 10},
			AliasPointer: api.AliasPointer{Value: "11"},
			Array: []testtypes.StringJSON{
				{Value: "1"},
				{Value: "2"},
				{Value: "3"},
			},
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/required?"+params.Encode(), nil)
		s.ServeHTTP(w, r)

		require.Equal(t, expected, th.requiredParams)

		c.Required(context.Background(), expected)
		require.Equal(t, params, tc.request.URL.Query())
	})

	t.Run("Optional", func(t *testing.T) {
		expected := api.OptionalParams{
			OgenString:   api.NewOptStringOgen(testtypes.StringOgen{Value: "1"}),
			OgenNumber:   api.NewOptNumberOgen(testtypes.NumberOgen{Value: 2}),
			JsonString:   api.NewOptStringJSON(testtypes.StringJSON{Value: "3"}),
			JsonNumber:   api.NewOptNumberJSON(testtypes.NumberJSON{Value: 4}),
			TextString:   api.NewOptText(testtypes.Text{Value: "5"}),
			TextNumber:   api.NewOptText(testtypes.Text{Value: "6"}),
			String:       api.NewOptString(testtypes.String("7")),
			Number:       api.NewOptNumber(testtypes.Number(8)),
			Alias:        api.NewOptAlias(api.Alias{Value: "9"}),
			Pointer:      api.NewOptPointer(testtypes.NumberOgen{Value: 10}),
			AliasPointer: api.NewOptAliasPointer(api.AliasPointer{Value: "11"}),
			Array: []testtypes.StringJSON{
				{Value: "1"},
				{Value: "2"},
				{Value: "3"},
			},
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/optional?"+params.Encode(), nil)
		s.ServeHTTP(w, r)

		require.Equal(t, expected, th.optionalParams)

		c.Optional(context.Background(), expected)
		require.Equal(t, params, tc.request.URL.Query())
	})

	t.Run("Defaults", func(t *testing.T) {
		expected := api.OptionalParams{
			OgenString:   api.NewOptStringOgen(testtypes.StringOgen{Value: "10"}),
			OgenNumber:   api.NewOptNumberOgen(testtypes.NumberOgen{Value: 20}),
			JsonString:   api.NewOptStringJSON(testtypes.StringJSON{Value: "30"}),
			JsonNumber:   api.NewOptNumberJSON(testtypes.NumberJSON{Value: 40}),
			TextString:   api.NewOptText(testtypes.Text{Value: "50"}),
			TextNumber:   api.NewOptText(testtypes.Text{Value: "60"}),
			String:       api.NewOptString(testtypes.String("70")),
			Number:       api.NewOptNumber(testtypes.Number(80)),
			Alias:        api.NewOptAlias(api.Alias{Value: "90"}),
			Pointer:      api.NewOptPointer(testtypes.NumberOgen{Value: 100}),
			AliasPointer: api.NewOptAliasPointer(api.AliasPointer{Value: "110"}),
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/optional", nil)
		s.ServeHTTP(w, r)

		require.Equal(t, expected, th.optionalParams)
	})
}

func TestTypeExtension_JSON(t *testing.T) {
	input := `{
		"ogenString": "1",
		"ogenNumber": 2,
		"jsonString": "3",
		"jsonNumber": 4,
		"textString": "5",
		"textNumber": 6,
		"string": "7",
		"number": 8,
		"alias": "9",
		"pointer": 10,
		"aliasPointer": "11",
		"builtin": { "key1": "foo", "key2": "bar" },
		"array": [ "1", "2", "3" ],
		"map": { "key1": "1", "key2": "2", "key3": "3" } 
	}`

	t.Run("Required", func(t *testing.T) {
		expected := api.RequiredOK{
			OgenString:   testtypes.StringOgen{Value: "1"},
			OgenNumber:   testtypes.NumberOgen{Value: 2},
			JsonString:   testtypes.StringJSON{Value: "3"},
			JsonNumber:   testtypes.NumberJSON{Value: 4},
			TextString:   testtypes.Text{Value: "5"},
			TextNumber:   testtypes.Text{Value: "6"},
			String:       testtypes.String("7"),
			Number:       testtypes.Number(8),
			Alias:        api.Alias{Value: "9"},
			Pointer:      testtypes.NumberOgen{Value: 10},
			AliasPointer: api.AliasPointer{Value: "11"},
			Builtin:      map[string]any{"key1": "foo", "key2": "bar"},
			Array: []testtypes.StringJSON{
				{Value: "1"},
				{Value: "2"},
				{Value: "3"},
			},
			Map: map[string]testtypes.StringJSON{
				"key1": {Value: "1"},
				"key2": {Value: "2"},
				"key3": {Value: "3"},
			},
		}

		a := require.New(t)
		var p api.RequiredOK
		a.NoError(p.Decode(jx.DecodeStr(input)))
		a.Equal(expected, p)

		out, err := p.MarshalJSON()
		a.NoError(err)
		a.JSONEq(input, string(out))
	})

	t.Run("Optional", func(t *testing.T) {
		expected := api.OptionalOK{
			OgenString:   api.NewOptStringOgen(testtypes.StringOgen{Value: "1"}),
			OgenNumber:   api.NewOptNumberOgen(testtypes.NumberOgen{Value: 2}),
			JsonString:   api.NewOptStringJSON(testtypes.StringJSON{Value: "3"}),
			JsonNumber:   api.NewOptNumberJSON(testtypes.NumberJSON{Value: 4}),
			TextString:   api.NewOptText(testtypes.Text{Value: "5"}),
			TextNumber:   api.NewOptText(testtypes.Text{Value: "6"}),
			String:       api.NewOptString(testtypes.String("7")),
			Number:       api.NewOptNumber(testtypes.Number(8)),
			Alias:        api.NewOptAlias(api.Alias{Value: "9"}),
			Pointer:      api.NewOptPointer(testtypes.NumberOgen{Value: 10}),
			AliasPointer: api.NewOptAliasPointer(api.AliasPointer{Value: "11"}),
			Builtin:      api.NewOptAny(map[string]any{"key1": "foo", "key2": "bar"}),
			Array: []testtypes.StringJSON{
				{Value: "1"},
				{Value: "2"},
				{Value: "3"},
			},
			Map: api.NewOptOptionalOKMap(map[string]testtypes.StringJSON{
				"key1": {Value: "1"},
				"key2": {Value: "2"},
				"key3": {Value: "3"},
			}),
		}

		a := require.New(t)
		var p api.OptionalOK
		a.NoError(p.Decode(jx.DecodeStr(input)))
		a.Equal(expected, p)

		out, err := p.MarshalJSON()
		a.NoError(err)
		a.JSONEq(input, string(out))
	})

	t.Run("Defaults", func(t *testing.T) {
		expected := api.OptionalOK{
			OgenString:   api.NewOptStringOgen(testtypes.StringOgen{Value: "10"}),
			OgenNumber:   api.NewOptNumberOgen(testtypes.NumberOgen{Value: 20}),
			JsonString:   api.NewOptStringJSON(testtypes.StringJSON{Value: "30"}),
			JsonNumber:   api.NewOptNumberJSON(testtypes.NumberJSON{Value: 40}),
			TextString:   api.NewOptText(testtypes.Text{Value: "50"}),
			TextNumber:   api.NewOptText(testtypes.Text{Value: "60"}),
			String:       api.NewOptString(testtypes.String("70")),
			Number:       api.NewOptNumber(testtypes.Number(80)),
			Alias:        api.NewOptAlias(api.Alias{Value: "90"}),
			Pointer:      api.NewOptPointer(testtypes.NumberOgen{Value: 100}),
			AliasPointer: api.NewOptAliasPointer(api.AliasPointer{Value: "110"}),
		}

		a := require.New(t)
		var p api.OptionalOK
		a.NoError(p.Decode(jx.DecodeStr(`{}`)))
		a.Equal(expected, p)
	})
}
