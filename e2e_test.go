package ogen

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-faster/jx"

	"github.com/ogen-go/ogen/conv"
	api "github.com/ogen-go/ogen/internal/sample_api"
	"github.com/ogen-go/ogen/internal/techempower"
	"github.com/ogen-go/ogen/json"
	"github.com/ogen-go/ogen/validate"
)

var (
	petExistingID int64 = 1337
	petNotFoundID int64 = 404
	petErrorID    int64 = 500

	petAvatar = []byte("pet avatar")
)

type techEmpowerServer struct{}

func (t techEmpowerServer) Caching(ctx context.Context, params techempower.CachingParams) (techempower.WorldObjects, error) {
	panic("implement me")
}

func (t techEmpowerServer) Updates(ctx context.Context, params techempower.UpdatesParams) (techempower.WorldObjects, error) {
	panic("implement me")
}

func (t techEmpowerServer) Queries(ctx context.Context, params techempower.QueriesParams) (techempower.WorldObjects, error) {
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

func (s sampleAPIServer) PetUpdateNameAliasPost(ctx context.Context, req api.PetName) (api.PetUpdateNameAliasPostDefStatusCode, error) {
	panic("implement me")
}

func (s sampleAPIServer) PetUpdateNamePost(ctx context.Context, req string) (api.PetUpdateNamePostDefStatusCode, error) {
	panic("implement me")
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

func (s sampleAPIServer) FoobarPut(ctx context.Context) (api.FoobarPutDefStatusCode, error) {
	panic("implement me")
}

func (s sampleAPIServer) FoobarPost(ctx context.Context, req api.Pet) (api.FoobarPostRes, error) {
	panic("implement me")
}

func (s sampleAPIServer) PetGet(ctx context.Context, params api.PetGetParams) (api.PetGetRes, error) {
	panic("implement me")
}

func (s *sampleAPIServer) PetCreate(ctx context.Context, req api.Pet) (pet api.Pet, err error) {
	s.pet = req
	return req, nil
}

func (s *sampleAPIServer) PetGetByName(ctx context.Context, params api.PetGetByNameParams) (api.Pet, error) {
	return s.pet, nil
}

func (s *sampleAPIServer) PetGetAvatarByID(ctx context.Context, params api.PetGetAvatarByIDParams) (api.PetGetAvatarByIDRes, error) {
	switch params.PetID {
	case petNotFoundID:
		return &api.NotFound{}, nil
	case petErrorID:
		return &api.ErrorStatusCode{
			StatusCode: http.StatusInternalServerError,
			Response:   api.Error{Message: "error"},
		}, nil
	default:
		return &api.PetGetAvatarByIDOKApplicationOctetStream{
			Data: bytes.NewReader(petAvatar),
		}, nil
	}
}

func (s *sampleAPIServer) PetUploadAvatarByID(ctx context.Context, req api.Stream, params api.PetUploadAvatarByIDParams) (api.PetUploadAvatarByIDRes, error) {
	switch params.PetID {
	case petNotFoundID:
		return &api.NotFound{}, nil
	case petErrorID:
		return &api.ErrorStatusCode{
			StatusCode: http.StatusInternalServerError,
			Response:   api.Error{Message: "error"},
		}, nil
	default:
		avatar, err := io.ReadAll(req)
		if err != nil {
			return &api.ErrorStatusCode{
				StatusCode: http.StatusInternalServerError,
				Response:   api.Error{Message: err.Error()},
			}, nil
		}

		if string(avatar) != string(petAvatar) {
			return &api.ErrorStatusCode{
				StatusCode: http.StatusBadRequest,
				Response:   api.Error{Message: "unexpected avatar"},
			}, nil
		}

		return &api.PetUploadAvatarByIDOK{}, nil
	}
}

func (s *sampleAPIServer) ErrorGet(ctx context.Context) (api.ErrorStatusCode, error) {
	return api.ErrorStatusCode{
		StatusCode: http.StatusInternalServerError,
		Response: api.Error{
			Message: "test_error",
		},
	}, nil
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
			require.Len(t, validateErr.Fields, 3)
			require.Equal(t, "invalid: id (value -1 less than 0), name (len 1 less than minimum 4), kind (invalid enum value: )", validateErr.Error())
		})

		s := httptest.NewServer(api.NewServer(&sampleAPIServer{}))
		defer s.Close()

		client, err := api.NewClient(s.URL)
		require.NoError(t, err)
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
			Kind:         api.PetKindSmol,
			Primary:      &primary,
			Friends:      []api.Pet{friend},
			TestArray1: [][]string{
				{"Foo", "Bar"},
				{"Baz"},
			},
			Next: api.NewOptData(api.Data{
				Description: api.NewDescriptionSimpleDataDescription(api.DescriptionSimple{
					Description: "foo",
				}),
				ID:       api.NewIntID(10),
				Email:    "foo@example.com",
				Format:   "1-2",
				Hostname: "example.org",
			}),
		}

		t.Run("Valid", func(t *testing.T) {
			data := json.Encode(pet)
			t.Logf("%s", data)
			require.True(t, jx.Valid(data), "json should be valid")
			require.JSONEq(t, petTestData, string(data), "should be equal to golden json")
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
			got, err := client.PetCreate(ctx, pet)
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

		t.Run("PetUploadAvatar", func(t *testing.T) {
			t.Run("OK", func(t *testing.T) {
				stream := api.Stream{
					Data: io.NopCloser(bytes.NewReader(petAvatar)),
				}
				got, err := client.PetUploadAvatarByID(ctx, stream, api.PetUploadAvatarByIDParams{
					PetID: petExistingID,
				})
				require.NoError(t, err)
				assert.IsType(t, &api.PetUploadAvatarByIDOK{}, got, fmt.Sprintf("receive response %v", got))
			})
			t.Run("NotFound", func(t *testing.T) {
				stream := api.Stream{
					Data: io.NopCloser(bytes.NewReader(petAvatar)),
				}
				got, err := client.PetUploadAvatarByID(ctx, stream, api.PetUploadAvatarByIDParams{
					PetID: petNotFoundID,
				})
				require.NoError(t, err)
				assert.IsType(t, &api.NotFound{}, got, fmt.Sprintf("receive response %v", got))
			})
			t.Run("Error", func(t *testing.T) {
				stream := api.Stream{
					Data: bytes.NewReader(petAvatar),
				}
				got, err := client.PetUploadAvatarByID(ctx, stream, api.PetUploadAvatarByIDParams{
					PetID: petErrorID,
				})
				require.NoError(t, err)
				assert.IsType(t, &api.ErrorStatusCode{}, got, fmt.Sprintf("receive response %v", got))
			})
		})
		t.Run("PetGetAvatar", func(t *testing.T) {
			t.Run("OK", func(t *testing.T) {
				got, err := client.PetGetAvatarByID(ctx, api.PetGetAvatarByIDParams{
					PetID: petExistingID,
				})
				require.NoError(t, err)
				assert.IsType(t, &api.PetGetAvatarByIDOKApplicationOctetStream{}, got, fmt.Sprintf("receive response %v", got))

				raw := got.(*api.PetGetAvatarByIDOKApplicationOctetStream)
				avatar, err := io.ReadAll(raw)
				require.NoError(t, err)

				require.Equal(t, petAvatar, avatar)
			})
			t.Run("NotFound", func(t *testing.T) {
				got, err := client.PetGetAvatarByID(ctx, api.PetGetAvatarByIDParams{
					PetID: petNotFoundID,
				})
				require.NoError(t, err)
				assert.IsType(t, &api.NotFound{}, got, fmt.Sprintf("receive response %v", got))
			})
			t.Run("Error", func(t *testing.T) {
				got, err := client.PetGetAvatarByID(ctx, api.PetGetAvatarByIDParams{
					PetID: petErrorID,
				})
				require.NoError(t, err)
				assert.IsType(t, &api.ErrorStatusCode{}, got, fmt.Sprintf("receive response %v", got))
			})
			t.Run("ErrorGet", func(t *testing.T) {
				got, err := client.ErrorGet(ctx)
				require.NoError(t, err)

				errStatusCode := api.ErrorStatusCode{
					StatusCode: http.StatusInternalServerError,
					Response: api.Error{
						Message: "test_error",
					},
				}
				assert.Equal(t, errStatusCode, got)
			})
		})
	})

	t.Run("TechEmpower", func(t *testing.T) {
		// Using TechEmpower as most popular general purpose framework benchmark.
		// https://github.com/TechEmpower/FrameworkBenchmarks/wiki/Project-Information-Framework-Tests-Overview#test-types
		t.Parallel()

		s := httptest.NewServer(techempower.NewServer(techEmpowerServer{}))
		defer s.Close()

		client, err := techempower.NewClient(s.URL)
		require.NoError(t, err)
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
