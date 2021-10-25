package ogen

import (
	"bytes"
	"context"
	_ "embed"
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
	"github.com/ogen-go/ogen/validate"
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

func (s sampleAPIServer) PetFriendsNamesByID(ctx context.Context, params api.PetFriendsNamesByIDParams) ([]string, error) {
	if int64(params.ID) != s.pet.ID {
		return []string{}, nil
	}
	var names []string
	for _, f := range s.pet.Friends {
		names = append(names, f.Name)
	}
	return names, nil
}

func (s sampleAPIServer) PetNameByID(ctx context.Context, params api.PetNameByIDParams) (string, error) {
	panic("implement me")
}

func (s sampleAPIServer) FoobarGet(ctx context.Context, params api.FoobarGetParams) (api.FoobarGetRes, error) {
	panic("implement me")
}

func (s sampleAPIServer) FoobarPut(ctx context.Context) (api.FoobarPutDefault, error) {
	panic("implement me")
}

func (s sampleAPIServer) FoobarPost(ctx context.Context, req *api.Pet) (api.FoobarPostRes, error) {
	panic("implement me")
}

func (s sampleAPIServer) PetGet(ctx context.Context, params api.PetGetParams) (api.PetGetRes, error) {
	panic("implement me")
}

func (s *sampleAPIServer) PetCreate(ctx context.Context, req api.PetCreateReq) (pet api.Pet, err error) {
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

//go:embed _testdata/pet.json
var petTestData string

func TestIntegration(t *testing.T) {
	t.Parallel()

	t.Run("Sample", func(t *testing.T) {
		t.Parallel()

		t.Run("Validate", func(t *testing.T) {
			badPet := api.Pet{
				Name: "k",
				ID:   -1,
			}
			err := badPet.Validate()
			require.Error(t, err)
			var validateErr *validate.Error
			require.ErrorAs(t, err, &validateErr)
			require.Len(t, validateErr.Fields, 2)
			require.Equal(t, "invalid: id (value -1 less than 0), name (len 1 less than minimum 4)", validateErr.Error())
		})

		mux := chi.NewRouter()
		api.Register(mux, &sampleAPIServer{})
		s := httptest.NewServer(mux)
		defer s.Close()

		client := api.NewClient(s.URL)
		ctx := context.Background()

		date := time.Date(2011, 10, 10, 7, 12, 34, 4125, time.UTC)

		friend := api.Pet{
			Birthday: conv.Date(date),
			ID:       43,
			Name:     "BestFriend",
			Rate:     time.Second * 5,
			URI:      url.URL{Scheme: "s3", Host: "foo", Path: "/baz"},
			IP:       net.IPv4(127, 0, 0, 2),
			IPV4:     net.IPv4(127, 0, 0, 2),
			Kind:     api.PetKindBig,
			IPV6:     net.ParseIP("2001:0db8:85a3:0000:0000:8a2e:0370:7335"),
			Nickname: api.NewNilString("friend"),
		}
		primary := friend // Explicitly allocate new value.

		pet := api.Pet{
			Birthday:     conv.Date(date),
			ID:           42,
			Type:         api.NewOptPetType(api.PetTypeFofa),
			Name:         "SomePet",
			Nickname:     api.NewNilString("Nick"),
			NullStr:      api.NewOptNilString("Bar"),
			Rate:         time.Second,
			Tag:          api.NewOptUUID(uuid.MustParse("fc9d49c6-1f3d-4ecb-92c7-be6d5049b3c8")),
			TestDate:     api.NewOptTime(conv.Date(date)),
			TestDateTime: api.NewOptTime(conv.DateTime(date)),
			TestDuration: api.NewOptDuration(time.Minute),
			TestFloat1:   api.NewOptFloat64(1.0),
			TestInteger1: api.NewOptInt(10),
			TestTime:     api.NewOptTime(conv.Time(date)),
			UniqueID:     uuid.MustParse("f76e18ae-e5ed-4342-922d-762ed1dfe593"),
			URI:          url.URL{Scheme: "s3", Host: "foo", Path: "/bar"},
			IP:           net.IPv4(127, 0, 0, 1),
			IPV4:         net.IPv4(127, 0, 0, 1),
			IPV6:         net.ParseIP("2001:0db8:85a3:0000:0000:8a2e:0370:7334"),
			Next:         api.NewOptData(api.Data{Description: api.NewOptString("Foo")}),
			Kind:         api.PetKindSmol,
			Primary:      &primary,
			Friends:      []api.Pet{friend},
			TestArray1: [][]string{
				{"Foo", "Bar"},
				{"Baz"},
			},
		}

		t.Run("Valid", func(t *testing.T) {
			buf := new(bytes.Buffer)
			require.NoError(t, pet.WriteJSONTo(buf))
			require.True(t, jsoniter.Valid(buf.Bytes()), "json should be valid")
			require.JSONEq(t, petTestData, buf.String(), "should be equal to golden json")
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
			a.Equal(exp.Kind, got.Kind, "Kind")

			a.Equal(exp.TestDate.Set, got.TestDate.Set, "TestDate")
			a.True(exp.TestDate.Value.Equal(got.TestDate.Value), "TestDate %s (exp) != %s (got)", exp.TestDate.Value, got.TestDate.Value)

			a.Equal(exp.TestDateTime.Set, got.TestDateTime.Set, "TestDateTime")
			a.True(exp.TestDateTime.Value.Equal(got.TestDateTime.Value), "TestDateTime %s (exp) != %s (got)", exp.TestDateTime.Value, got.TestDateTime.Value)

			a.Equal(exp.TestDuration, got.TestDuration, "TestDuration")
			a.Equal(exp.TestFloat1, got.TestFloat1, "TestFloat1")
			a.Equal(exp.TestInteger1, got.TestInteger1, "TestInteger1")

			// Probably we need separate type for Time.
			a.Equal(exp.TestTime.Set, got.TestTime.Set, "TestTime")
			a.Equal(exp.TestTime.Value.Hour(), got.TestTime.Value.Hour(), "TestTime hour")
			a.Equal(exp.TestTime.Value.Minute(), got.TestTime.Value.Minute(), "TestTime hour")
			a.Equal(exp.TestTime.Value.Second(), got.TestTime.Value.Second(), "TestTime hour")

			a.Equal(pet.UniqueID, got.UniqueID, "UniqueID")

			a.True(pet.IP.Equal(got.IP), "IP")
			a.True(pet.IPV4.Equal(got.IPV4), "IPV4")
			a.True(pet.IPV6.Equal(got.IPV6), "IPV6")

			a.Equal(pet.URI.String(), got.URI.String(), "URI")

			a.Equal(pet.Next, got.Next, "Next")

			a.Equal(pet.Type, got.Type, "Type")

			a.Equal(pet.Friends, got.Friends, "Friends")
			a.Equal(pet.TestArray1, got.TestArray1, "TestArray1")
			a.Equal(pet.Primary, got.Primary, "Primary")
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
			t.Run("PetGet", func(t *testing.T) {
				got, err := client.PetFriendsNamesByID(ctx, api.PetFriendsNamesByIDParams{ID: int(pet.ID)})
				require.NoError(t, err)
				assert.Equal(t, []string{friend.Name}, got)
			})
		})
	})
	t.Run("TechEmpower", func(t *testing.T) {
		// Using TechEmpower as most popular general purpose framework benchmark.
		// https://github.com/TechEmpower/FrameworkBenchmarks/wiki/Project-Information-Framework-Tests-Overview#test-types
		t.Parallel()

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
