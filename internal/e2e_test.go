package internal

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"testing"
	"time"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	api.UnimplementedHandler
	pet api.Pet
}

func (s sampleAPIServer) DataGetFormat(ctx context.Context, params api.DataGetFormatParams) (string, error) {
	return fmt.Sprintf(
		"%d %s %s %s %s",
		params.ID,
		params.Foo,
		params.Bar,
		params.Baz,
		params.Kek,
	), nil
}

func (s sampleAPIServer) GetHeader(ctx context.Context, params api.GetHeaderParams) (api.Hash, error) {
	h := sha256.Sum256([]byte(params.XAuthToken))
	return api.Hash{
		Raw: h[:],
		Hex: hex.EncodeToString(h[:]),
	}, nil
}

func (s sampleAPIServer) PetUpdateNameAliasPost(ctx context.Context, req api.OptPetName) (api.PetUpdateNameAliasPostDef, error) {
	panic("implement me")
}

func (s sampleAPIServer) PetUpdateNamePost(ctx context.Context, req api.OptString) (api.PetUpdateNamePostDef, error) {
	code := http.StatusAccepted
	if _, ok := req.Get(); ok {
		code = http.StatusOK
	}
	return api.PetUpdateNamePostDef{
		StatusCode: code,
	}, nil
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

func (s sampleAPIServer) FoobarPut(ctx context.Context) (api.FoobarPutDef, error) {
	panic("implement me")
}

func (s sampleAPIServer) FoobarPost(ctx context.Context, req api.OptPet) (api.FoobarPostRes, error) {
	panic("implement me")
}

func (s sampleAPIServer) PetGet(ctx context.Context, params api.PetGetParams) (api.PetGetRes, error) {
	panic("implement me")
}

func (s *sampleAPIServer) PetCreate(ctx context.Context, req api.OptPet) (pet api.Pet, err error) {
	if val, ok := req.Get(); ok {
		s.pet = val
	}
	return req.Value, nil
}

func (s *sampleAPIServer) PetGetByName(ctx context.Context, params api.PetGetByNameParams) (api.Pet, error) {
	return s.pet, nil
}

func (s *sampleAPIServer) PetGetAvatarByName(ctx context.Context, params api.PetGetAvatarByNameParams) (api.PetGetAvatarByNameRes, error) {
	return &api.PetGetAvatarByNameOK{
		Data: bytes.NewReader(petAvatar),
	}, nil
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
		return &api.PetGetAvatarByIDOK{
			Data: bytes.NewReader(petAvatar),
		}, nil
	}
}

func (s *sampleAPIServer) PetUploadAvatarByID(ctx context.Context, req api.PetUploadAvatarByIDReq, params api.PetUploadAvatarByIDParams) (api.PetUploadAvatarByIDRes, error) {
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

		if !bytes.Equal(avatar, petAvatar) {
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

func (s sampleAPIServer) TestFloatValidation(ctx context.Context, req api.TestFloatValidation) error {
	panic("implement me")
}

func (s *sampleAPIServer) TestObjectQueryParameter(ctx context.Context, params api.TestObjectQueryParameterParams) (api.TestObjectQueryParameterOK, error) {
	if param, ok := params.FormObject.Get(); ok {
		return api.TestObjectQueryParameterOK{
			Style:  "form",
			Min:    param.Min,
			Max:    param.Max,
			Filter: param.Filter,
		}, nil
	}
	if param, ok := params.DeepObject.Get(); ok {
		return api.TestObjectQueryParameterOK{
			Style:  "deepObject",
			Min:    param.Min,
			Max:    param.Max,
			Filter: param.Filter,
		}, nil
	}
	return api.TestObjectQueryParameterOK{}, errors.New("invalid input")
}

func (s sampleAPIServer) OneofBug(ctx context.Context, req api.OneOfBugs) error {
	panic("implement me")
}

func (s sampleAPIServer) RecursiveMapGet(ctx context.Context) (api.RecursiveMap, error) {
	panic("implement me")
}

func (s sampleAPIServer) RecursiveArrayGet(ctx context.Context) (api.RecursiveArray, error) {
	panic("implement me")
}

func (s sampleAPIServer) DefaultTest(ctx context.Context, req api.DefaultTest, params api.DefaultTestParams) (int32, error) {
	return params.Default.Value, nil
}

func (s sampleAPIServer) NullableDefaultResponse(ctx context.Context) (api.NilIntStatusCode, error) {
	return api.NilIntStatusCode{
		StatusCode: 200,
		Response:   api.NewNilInt(1337),
	}, nil
}

func (s sampleAPIServer) TestContentParameter(ctx context.Context, params api.TestContentParameterParams) (string, error) {
	val, _ := params.Param.Get()
	return val.Style, nil
}

var _ api.Handler = (*sampleAPIServer)(nil)

//go:embed _testdata/payloads/pet.json
var petTestData string

func TestIntegration(t *testing.T) {
	t.Parallel()

	t.Run("Sample", func(t *testing.T) {
		t.Parallel()

		t.Run("Validate", func(t *testing.T) {
			badPet := api.Pet{
				Name: "k",
				ID:   -1,
				Kind: api.PetKindSmol,
			}
			err := badPet.Validate()
			require.Error(t, err)
			var validateErr *validate.Error
			require.ErrorAs(t, err, &validateErr)
			require.Len(t, validateErr.Fields, 2)
			require.Equal(t, "invalid: id (int: value -1 less than 0), name (string: len 1 less than minimum 4)", validateErr.Error())
		})

		handler := &sampleAPIServer{}
		h, err := api.NewServer(handler, handler)
		require.NoError(t, err)
		s := httptest.NewServer(h)
		defer s.Close()

		httpClient := s.Client()
		client, err := api.NewClient(s.URL, handler, api.WithClient(httpClient))
		require.NoError(t, err)
		ctx := context.Background()

		date := time.Date(2011, 10, 10, 7, 12, 34, 4125, time.UTC)

		friend := api.Pet{
			Birthday: conv.Date(date),
			ID:       43,
			Name:     "BestFriend",
			Rate:     time.Second * 5,
			URI:      url.URL{Scheme: "s3", Host: "foo", Path: "/baz"},
			IP:       netip.AddrFrom4([4]byte{127, 0, 0, 2}),
			IPV4:     netip.AddrFrom4([4]byte{127, 0, 0, 2}),
			Kind:     api.PetKindBig,
			IPV6:     netip.MustParseAddr("2001:0db8:85a3:0000:0000:8a2e:0370:7335"),
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
			TestDate:     api.NewOptDate(conv.Date(date)),
			TestDateTime: api.NewOptDateTime(conv.DateTime(date)),
			TestDuration: api.NewOptDuration(time.Minute),
			TestFloat1:   api.NewOptFloat64(20.0),
			TestInteger1: api.NewOptInt(10),
			TestTime:     api.NewOptTime(conv.Time(date)),
			UniqueID:     uuid.MustParse("f76e18ae-e5ed-4342-922d-762ed1dfe593"),
			URI:          url.URL{Scheme: "s3", Host: "foo", Path: "/bar"},
			IP:           netip.AddrFrom4([4]byte{127, 0, 0, 1}),
			IPV4:         netip.AddrFrom4([4]byte{127, 0, 0, 1}),
			IPV6:         netip.MustParseAddr("2001:0db8:85a3:0000:0000:8a2e:0370:7334"),
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
				Base64:   []byte("hello, world!"),
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

			a.True(pet.IP.Compare(got.IP) == 0, "IP")
			a.True(pet.IPV4.Compare(got.IPV4) == 0, "IPV4")
			a.True(pet.IPV6.Compare(got.IPV6) == 0, "IPV6")

			a.Equal(pet.URI.String(), got.URI.String(), "URI")

			a.Equal(pet.Next, got.Next, "Next")

			a.Equal(pet.Type, got.Type, "Type")

			a.Equal(pet.Friends, got.Friends, "Friends")
			a.Equal(pet.TestArray1, got.TestArray1, "TestArray1")
			a.Equal(pet.Primary, got.Primary, "Primary")
		}

		t.Run("PetCreate", func(t *testing.T) {
			got, err := client.PetCreate(ctx, api.NewOptPet(pet))
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
				stream := api.PetUploadAvatarByIDReq{
					Data: io.NopCloser(bytes.NewReader(petAvatar)),
				}
				got, err := client.PetUploadAvatarByID(ctx, stream, api.PetUploadAvatarByIDParams{
					PetID: petExistingID,
				})
				require.NoError(t, err)
				assert.IsType(t, &api.PetUploadAvatarByIDOK{}, got, fmt.Sprintf("receive response %v", got))
			})
			t.Run("NotFound", func(t *testing.T) {
				stream := api.PetUploadAvatarByIDReq{
					Data: io.NopCloser(bytes.NewReader(petAvatar)),
				}
				got, err := client.PetUploadAvatarByID(ctx, stream, api.PetUploadAvatarByIDParams{
					PetID: petNotFoundID,
				})
				require.NoError(t, err)
				assert.IsType(t, &api.NotFound{}, got, fmt.Sprintf("receive response %v", got))
			})
			t.Run("Error", func(t *testing.T) {
				stream := api.PetUploadAvatarByIDReq{
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
				assert.IsType(t, &api.PetGetAvatarByIDOK{}, got, fmt.Sprintf("receive response %v", got))

				raw := got.(*api.PetGetAvatarByIDOK)
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
		t.Run("PetUpdateNamePost", func(t *testing.T) {
			// Ensure optional body handled correctly.
			h, err := client.PetUpdateNamePost(ctx, api.OptString{})
			require.NoError(t, err)
			require.Equal(t, http.StatusAccepted, h.StatusCode)

			h, err = client.PetUpdateNamePost(ctx, api.NewOptString("amongus"))
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, h.StatusCode)
		})
		t.Run("GetHeader", func(t *testing.T) {
			h, err := client.GetHeader(ctx, api.GetHeaderParams{XAuthToken: "hello, world"})
			require.NoError(t, err)
			assert.NotEmpty(t, h.Raw)
			assert.Equal(t, hex.EncodeToString(h.Raw), h.Hex)
			assert.Equal(t, "09ca7e4eaa6e8ae9c7d261167129184883644d07dfba7cbfbc4c8a2e08360d5b", h.Hex)
		})
		t.Run("DataGetFormat", func(t *testing.T) {
			a := require.New(t)
			// Path: /name/{id}/{foo}1234{bar}-{baz}!{kek}
			req, err := http.NewRequestWithContext(ctx,
				http.MethodGet, s.URL+"/name/1/foo-1234bar+-baz/!kek*", http.NoBody)
			a.NoError(err)

			resp, err := httpClient.Do(req)
			a.NoError(err)

			data, err := io.ReadAll(resp.Body)
			a.NoError(err)
			a.Equal(`"1 foo- bar+ baz/ kek*"`, string(data))

			h, err := client.DataGetFormat(ctx, api.DataGetFormatParams{
				ID:  1,
				Foo: "foo-",
				Bar: "bar+",
				Baz: "baz/",
				Kek: "kek*",
			})
			a.NoError(err)
			assert.Equal(t, "1 foo- bar+ baz/ kek*", h)
		})
		t.Run("TestObjectQueryParameter", func(t *testing.T) {
			const (
				min    = 1
				max    = 5
				filter = "abc"
			)

			t.Run("formStyle", func(t *testing.T) {
				resp, err := client.TestObjectQueryParameter(ctx, api.TestObjectQueryParameterParams{
					FormObject: api.NewOptTestObjectQueryParameterFormObject(api.TestObjectQueryParameterFormObject{
						Min:    min,
						Max:    max,
						Filter: filter,
					}),
				})
				require.NoError(t, err)
				require.Equal(t, resp.Style, "form")
				require.Equal(t, resp.Min, min)
				require.Equal(t, resp.Max, max)
				require.Equal(t, resp.Filter, filter)
			})
			t.Run("deepObjectStyle", func(t *testing.T) {
				resp, err := client.TestObjectQueryParameter(ctx, api.TestObjectQueryParameterParams{
					DeepObject: api.NewOptTestObjectQueryParameterDeepObject(api.TestObjectQueryParameterDeepObject{
						Min:    min,
						Max:    max,
						Filter: filter,
					}),
				})
				require.NoError(t, err)
				require.Equal(t, resp.Style, "deepObject")
				require.Equal(t, resp.Min, min)
				require.Equal(t, resp.Max, max)
				require.Equal(t, resp.Filter, filter)
			})
		})
		t.Run("DefaultParameters", func(t *testing.T) {
			a := require.New(t)

			resp, err := client.DefaultTest(ctx, api.DefaultTest{}, api.DefaultTestParams{})
			a.NoError(err)
			a.Equal(int32(10), resp)

			resp, err = client.DefaultTest(ctx, api.DefaultTest{}, api.DefaultTestParams{
				Default: api.NewOptInt32(42),
			})
			a.NoError(err)
			a.Equal(int32(42), resp)
		})
		t.Run("HeaderSecurity", func(t *testing.T) {
			a := require.New(t)

			resp, err := client.SecurityTest(ctx)
			a.NoError(err)
			a.Equal("десять", resp)
		})
		t.Run("TestContentParameter", func(t *testing.T) {
			a := require.New(t)
			a.HTTPBodyContains(h.ServeHTTP, http.MethodGet, s.URL+"/testContentParameter", url.Values{
				"param": {`{"filter":"bar","style":"foo","min":10,"max":10}`},
			}, "foo")
		})
	})

	t.Run("TechEmpower", func(t *testing.T) {
		// Using TechEmpower as most popular general purpose framework benchmark.
		// https://github.com/TechEmpower/FrameworkBenchmarks/wiki/Project-Information-Framework-Tests-Overview#test-types
		t.Parallel()

		h, err := techempower.NewServer(techEmpowerServer{})
		require.NoError(t, err)

		s := httptest.NewServer(h)
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
