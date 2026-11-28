package integration

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/integration/test_embed_struct_aliases"
)

// Methods of the aliased struct are promoted to the alias.
type petGetter interface {
	GetName() string
	GetTag() api.OptString
}

var (
	_ petGetter = (*api.CreatePetApplicationJSON)(nil)
	_ petGetter = (*api.CreatePetApplicationXWwwFormUrlencoded)(nil)
	_ petGetter = (*api.CreatePetOK)(nil)
	_ petGetter = (*api.CreatePetCreated)(nil)
)

type embedStructAliasesHandler struct{}

func (embedStructAliasesHandler) CreatePet(_ context.Context, req api.CreatePetReq) (api.CreatePetRes, error) {
	switch req := req.(type) {
	case *api.CreatePetApplicationJSON:
		return &api.CreatePetOK{Pet: req.Pet}, nil
	case *api.CreatePetApplicationXWwwFormUrlencoded:
		return &api.CreatePetCreated{Pet: req.Pet}, nil
	default:
		panic("unreachable")
	}
}

func TestEmbedStructAliases(t *testing.T) {
	pet := api.Pet{Name: "Rex", Tag: api.NewOptString("dog")}

	t.Run("JSON", func(t *testing.T) {
		a := require.New(t)

		v := api.CreatePetOK{Pet: pet}
		e := jx.GetEncoder()
		v.Encode(e)
		a.JSONEq(`{"name":"Rex","tag":"dog"}`, e.String())

		var got api.CreatePetOK
		a.NoError(got.Decode(jx.DecodeBytes(e.Bytes())))
		a.Equal(v, got)
		a.Equal("Rex", got.GetName())
	})

	t.Run("Validate", func(t *testing.T) {
		a := require.New(t)

		a.NoError((&api.CreatePetApplicationJSON{Pet: pet}).Validate())
		a.Error((&api.CreatePetApplicationJSON{Pet: api.Pet{Name: "TooLongName"}}).Validate())
	})

	t.Run("Server", func(t *testing.T) {
		a := require.New(t)
		ctx := context.Background()

		srv, err := api.NewServer(embedStructAliasesHandler{})
		a.NoError(err)
		s := httptest.NewServer(srv)
		t.Cleanup(s.Close)

		client, err := api.NewClient(s.URL, api.WithClient(s.Client()))
		a.NoError(err)

		res, err := client.CreatePet(ctx, &api.CreatePetApplicationJSON{Pet: pet})
		a.NoError(err)
		a.Equal(&api.CreatePetOK{Pet: pet}, res)

		res, err = client.CreatePet(ctx, &api.CreatePetApplicationXWwwFormUrlencoded{Pet: pet})
		a.NoError(err)
		a.Equal(&api.CreatePetCreated{Pet: pet}, res)

		_, err = client.CreatePet(ctx, &api.CreatePetApplicationJSON{Pet: api.Pet{Name: "TooLongName"}})
		a.Error(err)
	})
}
