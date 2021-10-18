package ogen

import (
	"bytes"
	"context"
	"net"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/conv"
	api "github.com/ogen-go/ogen/internal/sample_api"
	"github.com/ogen-go/ogen/internal/techempower"
)

type techEmpowerServer struct{}

func (t techEmpowerServer) Caching(ctx context.Context, params techempower.CachingParams) ([]techempower.WorldObject, error) {
	panic("implement me")
}

func (t techEmpowerServer) Updates(ctx context.Context, params techempower.UpdatesParams) ([]techempower.WorldObject, error) {
	panic("implement me")
}

func (t techEmpowerServer) Queries(ctx context.Context, params techempower.QueriesParams) ([]techempower.WorldObject, error) {
	return nil, nil
}

func (t techEmpowerServer) DB(ctx context.Context) (techempower.WorldObject, error) {
	return techempower.WorldObject{
		ID:           1,
		RandomNumber: 10,
	}, nil
}

func (t techEmpowerServer) JSON(ctx context.Context) (techempower.HelloWorld, error) {
	return techempower.HelloWorld{
		Message: "Hello, world!",
	}, nil
}

type sampleAPIServer struct {
	pet api.Pet
}

func (s sampleAPIServer) FoobarGet(ctx context.Context, params api.FoobarGetParams) (api.FoobarGetResponse, error) {
	panic("implement me")
}

func (s sampleAPIServer) FoobarPut(ctx context.Context) (api.FoobarPutDefault, error) {
	panic("implement me")
}

func (s sampleAPIServer) FoobarPost(ctx context.Context, req *api.Pet) (api.FoobarPostResponse, error) {
	panic("implement me")
}

func (s sampleAPIServer) PetGet(ctx context.Context, params api.PetGetParams) (api.PetGetResponse, error) {
	panic("implement me")
}

func (s *sampleAPIServer) PetCreate(ctx context.Context, req api.PetCreateRequest) (pet api.Pet, err error) {
	switch p := req.(type) {
	case *api.Pet:
		s.pet = *p
		return s.pet, nil
	default:
		panic("not implemented")
	}
}

func (s *sampleAPIServer) PetGetByName(ctx context.Context, params api.PetGetByNameParams) (api.Pet, error) {
	return s.pet, nil
}

func TestIntegration(t *testing.T) {
	t.Run("Sample", func(t *testing.T) {
		mux := chi.NewRouter()
		api.Register(mux, &sampleAPIServer{})
		s := httptest.NewServer(mux)
		defer s.Close()

		client := api.NewClient(s.URL)
		ctx := context.Background()

		date := time.Date(2011, 10, 10, 7, 12, 34, 4125, time.UTC)
		pet := api.Pet{
			Birthday:     conv.Date(date),
			ID:           42,
			Type:         api.NewOptPetType(api.PetTypeFofa),
			Name:         "SomePet",
			Nickname:     api.NewNilString("Nick"),
			NullStr:      api.NewOptNilString("Bar"),
			Rate:         time.Second,
			Tag:          api.NewOptUUID(uuid.New()),
			TestDate:     api.NewOptTime(conv.Date(date)),
			TestDateTime: api.NewOptTime(conv.DateTime(date)),
			TestDuration: api.NewOptDuration(time.Minute),
			TestFloat1:   api.NewOptFloat64(1.0),
			TestInteger1: api.NewOptInt(10),
			TestTime:     api.NewOptTime(conv.Time(date)),
			UniqueID:     uuid.New(),
			URI:          url.URL{Scheme: "s3", Host: "foo", Path: "bar"},
			IP:           net.IPv4(127, 0, 0, 1),
			IPV4:         net.IPv4(127, 0, 0, 1),
			IPV6:         net.ParseIP("2001:0db8:85a3:0000:0000:8a2e:0370:7334"),
			Next:         api.NewOptData(api.Data{Description: api.NewOptString("Foo")}),

			// TODO(ernado): support decoding and check those
			Friends: nil,
			TestArray1: [][]string{
				{"Foo", "Bar"},
				{"Baz"},
				{},
			},
		}

		t.Run("Valid", func(t *testing.T) {
			buf := new(bytes.Buffer)
			require.NoError(t, pet.WriteJSONTo(buf))
			require.True(t, jsoniter.Valid(buf.Bytes()), "json should be valid")
		})

		// Can't use assert.Equal due to time.Time type equality checks.
		assertPet := func(t testing.TB, exp, got api.Pet) {
			a := assert.New(t)
			a.True(exp.Birthday.Equal(got.Birthday), "Birthday")
			a.Equal(exp.ID, got.ID, "ID")
			a.Equal(exp.Name, got.Name, "Name")
			a.Equal(exp.Nickname, got.Nickname, "Nickname")
			a.Equal(exp.NullStr, got.NullStr, "NullStr")
			a.Equal(exp.Rate, got.Rate, "Rate")
			a.Equal(exp.Tag, got.Tag, "Tag")

			a.Equal(exp.TestDate.Set, got.TestDate.Set, "TestDate")
			a.True(exp.TestDate.Value.Equal(got.TestDate.Value), "TestDate %s (exp) != %s (got)", exp.TestDate.Value, got.TestDate.Value)

			a.Equal(exp.TestDateTime.Set, got.TestDateTime.Set, "TestDateTime")
			a.True(exp.TestDateTime.Value.Equal(got.TestDateTime.Value), "TestDateTime %s (exp) != %s (got)", exp.TestDateTime.Value, got.TestDateTime.Value)

			a.Equal(exp.TestDuration, got.TestDuration, "TestDuration")
			a.Equal(exp.TestFloat1, got.TestFloat1, "TestFloat1")
			a.Equal(exp.TestInteger1, got.TestInteger1, "TestInteger1")

			// Probably we need separate type for Time.
			a.Equal(exp.TestTime.Set, got.TestTime.Set, "TestTime")
			a.Equal(exp.TestTime.Value.Hour(), got.TestTime.Value.Hour())
			a.Equal(exp.TestTime.Value.Minute(), got.TestTime.Value.Minute())
			a.Equal(exp.TestTime.Value.Second(), got.TestTime.Value.Second())

			a.Equal(pet.UniqueID, got.UniqueID)

			a.True(pet.IP.Equal(got.IP), "IP")
			a.True(pet.IPV4.Equal(got.IPV4), "IPV4")
			a.True(pet.IPV6.Equal(got.IPV6), "IPV6")

			a.Equal(pet.URI.String(), got.URI.String(), "URI")

			a.Equal(pet.Next, got.Next, "Next")

			a.Equal(pet.Type, got.Type, "Type")
		}

		t.Run("PetCreate", func(t *testing.T) {
			got, err := client.PetCreate(ctx, &pet)
			require.NoError(t, err)
			assertPet(t, pet, got)

			t.Run("PetGet", func(t *testing.T) {
				got, err := client.PetGetByName(ctx, api.PetGetByNameParams{Name: pet.Name})
				require.NoError(t, err)
				assertPet(t, pet, got)
			})
		})
	})
	t.Run("TechEmpower", func(t *testing.T) {
		// Using TechEmpower as most popular general purpose framework benchmark.
		// https://github.com/TechEmpower/FrameworkBenchmarks/wiki/Project-Information-Framework-Tests-Overview#test-types

		mux := chi.NewRouter()
		techempower.Register(mux, techEmpowerServer{})
		s := httptest.NewServer(mux)
		defer s.Close()

		client := techempower.NewClient(s.URL)
		ctx := context.Background()

		t.Run("JSON", func(t *testing.T) {
			res, err := client.JSON(ctx)
			require.NoError(t, err)
			require.Equal(t, "Hello, world!", res.Message)
		})
		t.Run("DB", func(t *testing.T) {
			res, err := client.DB(ctx)
			require.NoError(t, err)
			require.Equal(t, int64(1), res.ID)
			require.Equal(t, int64(10), res.RandomNumber)
		})
	})
}
