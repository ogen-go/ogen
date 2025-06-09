package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/integration/test_type_extension_name"
)

type testTypeNameHandler struct {
	optionalParams api.OptionalParams
	requiredParams api.RequiredParams
}

func (h *testTypeNameHandler) Optional(ctx context.Context, params api.OptionalParams) (*api.OptionalOK, error) {
	h.optionalParams = params
	return nil, assert.AnError
}

func (h *testTypeNameHandler) Required(ctx context.Context, params api.RequiredParams) (*api.RequiredOK, error) {
	h.requiredParams = params
	return nil, assert.AnError
}

type testTypeNameClient struct {
	request *http.Request
}

func (h *testTypeNameClient) Do(r *http.Request) (*http.Response, error) {
	h.request = r
	return nil, assert.AnError
}

func TestTypeExtensionName_Params(t *testing.T) {
	th := &testTypeNameHandler{}
	s, _ := api.NewServer(th)

	tc := &testTypeNameClient{}
	c, err := api.NewClient("", api.WithClient(tc))
	require.NoError(t, err)

	params := "?bar=0.2&foo=0.1"

	t.Run("Required", func(t *testing.T) {
		expected := api.RequiredParams{
			Foo: "0.1",
			Bar: 0.2,
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/required"+params, nil)
		s.ServeHTTP(w, r)

		require.Equal(t, expected, th.requiredParams)

		c.Required(context.Background(), expected)
		require.Equal(t, "/required"+params, tc.request.URL.String())
	})

	t.Run("Optional", func(t *testing.T) {
		expected := api.OptionalParams{
			Foo: api.NewOptDecimal("0.1"),
			Bar: api.NewOptDecimal2(0.2),
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/optional"+params, nil)
		s.ServeHTTP(w, r)

		require.Equal(t, expected, th.optionalParams)

		c.Optional(context.Background(), expected)
		require.Equal(t, "/optional"+params, tc.request.URL.String())
	})

	t.Run("Defaults", func(t *testing.T) {
		expected := api.OptionalParams{
			Foo: api.NewOptDecimal("1.23"),
			Bar: api.NewOptDecimal2(1.23),
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/optional", nil)
		s.ServeHTTP(w, r)

		require.Equal(t, expected, th.optionalParams)
	})
}

func TestTypeExtensionName_JSON(t *testing.T) {
	input := `{ "foo": "0.1", "bar": 0.2 }`

	t.Run("Required", func(t *testing.T) {
		expected := api.RequiredOK{
			Foo: "0.1",
			Bar: 0.2,
		}

		a := require.New(t)
		var p api.RequiredOK
		a.NoError(p.Decode(jx.DecodeStr(input)))
		a.Equal(p, expected)

		out, err := p.MarshalJSON()
		a.NoError(err)
		a.JSONEq(input, string(out))
	})

	t.Run("Optional", func(t *testing.T) {
		expected := api.OptionalOK{
			Foo: api.NewOptDecimal("0.1"),
			Bar: api.NewOptDecimal2(0.2),
		}

		a := require.New(t)
		var p api.OptionalOK
		a.NoError(p.Decode(jx.DecodeStr(input)))
		a.Equal(p, expected)

		out, err := p.MarshalJSON()
		a.NoError(err)
		a.JSONEq(input, string(out))
	})

	t.Run("Defaults", func(t *testing.T) {
		expected := api.OptionalOK{
			Foo: api.NewOptDecimal("1.23"),
			Bar: api.NewOptDecimal2(1.23),
		}

		a := require.New(t)
		var p api.OptionalOK
		a.NoError(p.Decode(jx.DecodeStr(`{}`)))
		a.Equal(expected, p)
	})
}
