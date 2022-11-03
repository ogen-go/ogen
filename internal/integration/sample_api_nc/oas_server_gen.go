// Code generated by ogen, DO NOT EDIT.

package api

import (
	"context"
)

// Handler handles operations described by OpenAPI v3 specification.
type Handler interface {
	// DataGetFormat implements dataGetFormat operation.
	//
	// Retrieve data.
	//
	// GET /name/{id}/{foo}1234{bar}-{baz}!{kek}
	DataGetFormat(ctx context.Context, params DataGetFormatParams) (string, error)
	// DefaultTest implements defaultTest operation.
	//
	// POST /defaultTest
	DefaultTest(ctx context.Context, req DefaultTest, params DefaultTestParams) (int32, error)
	// ErrorGet implements errorGet operation.
	//
	// Returns error.
	//
	// GET /error
	ErrorGet(ctx context.Context) (ErrorStatusCode, error)
	// FoobarGet implements foobarGet operation.
	//
	// Dumb endpoint for testing things.
	//
	// GET /foobar
	FoobarGet(ctx context.Context, params FoobarGetParams) (FoobarGetRes, error)
	// FoobarPost implements foobarPost operation.
	//
	// Dumb endpoint for testing things.
	//
	// POST /foobar
	FoobarPost(ctx context.Context, req OptPet) (FoobarPostRes, error)
	// FoobarPut implements PUT /foobar operation.
	//
	// PUT /foobar
	FoobarPut(ctx context.Context) (FoobarPutDef, error)
	// GetHeader implements getHeader operation.
	//
	// Test for header param.
	//
	// GET /test/header
	GetHeader(ctx context.Context, params GetHeaderParams) (Hash, error)
	// NoAdditionalPropertiesTest implements noAdditionalPropertiesTest operation.
	//
	// GET /noAdditionalPropertiesTest
	NoAdditionalPropertiesTest(ctx context.Context) (NoAdditionalPropertiesTest, error)
	// NullableDefaultResponse implements nullableDefaultResponse operation.
	//
	// GET /nullableDefaultResponse
	NullableDefaultResponse(ctx context.Context) (NilIntStatusCode, error)
	// OneofBug implements oneofBug operation.
	//
	// POST /oneofBug
	OneofBug(ctx context.Context, req OneOfBugs) (OneofBugOK, error)
	// PatternRecursiveMapGet implements GET /patternRecursiveMap operation.
	//
	// GET /patternRecursiveMap
	PatternRecursiveMapGet(ctx context.Context) (PatternRecursiveMap, error)
	// PetCreate implements petCreate operation.
	//
	// Creates pet.
	//
	// POST /pet
	PetCreate(ctx context.Context, req OptPet) (Pet, error)
	// PetFriendsNamesByID implements petFriendsNamesByID operation.
	//
	// Returns names of all friends of pet.
	//
	// GET /pet/friendNames/{id}
	PetFriendsNamesByID(ctx context.Context, params PetFriendsNamesByIDParams) ([]string, error)
	// PetGet implements petGet operation.
	//
	// Returns pet from the system that the user has access to.
	//
	// GET /pet
	PetGet(ctx context.Context, params PetGetParams) (PetGetRes, error)
	// PetGetAvatarByID implements petGetAvatarByID operation.
	//
	// Returns pet avatar by id.
	//
	// GET /pet/avatar
	PetGetAvatarByID(ctx context.Context, params PetGetAvatarByIDParams) (PetGetAvatarByIDRes, error)
	// PetGetAvatarByName implements petGetAvatarByName operation.
	//
	// Returns pet's avatar by name.
	//
	// GET /pet/{name}/avatar
	PetGetAvatarByName(ctx context.Context, params PetGetAvatarByNameParams) (PetGetAvatarByNameRes, error)
	// PetGetByName implements petGetByName operation.
	//
	// Returns pet by name from the system that the user has access to.
	//
	// GET /pet/{name}
	PetGetByName(ctx context.Context, params PetGetByNameParams) (Pet, error)
	// PetNameByID implements petNameByID operation.
	//
	// Returns pet name by pet id.
	//
	// GET /pet/name/{id}
	PetNameByID(ctx context.Context, params PetNameByIDParams) (string, error)
	// PetUpdateNameAliasPost implements POST /pet/updateNameAlias operation.
	//
	// POST /pet/updateNameAlias
	PetUpdateNameAliasPost(ctx context.Context, req OptPetName) (PetUpdateNameAliasPostDef, error)
	// PetUpdateNamePost implements POST /pet/updateName operation.
	//
	// POST /pet/updateName
	PetUpdateNamePost(ctx context.Context, req OptString) (PetUpdateNamePostDef, error)
	// PetUploadAvatarByID implements petUploadAvatarByID operation.
	//
	// Uploads pet avatar by id.
	//
	// POST /pet/avatar
	PetUploadAvatarByID(ctx context.Context, req PetUploadAvatarByIDReq, params PetUploadAvatarByIDParams) (PetUploadAvatarByIDRes, error)
	// RecursiveArrayGet implements GET /recursiveArray operation.
	//
	// GET /recursiveArray
	RecursiveArrayGet(ctx context.Context) (RecursiveArray, error)
	// RecursiveMapGet implements GET /recursiveMap operation.
	//
	// GET /recursiveMap
	RecursiveMapGet(ctx context.Context) (RecursiveMap, error)
	// SecurityTest implements securityTest operation.
	//
	// GET /securityTest
	SecurityTest(ctx context.Context) (string, error)
	// StringIntMapGet implements GET /stringIntMap operation.
	//
	// GET /stringIntMap
	StringIntMapGet(ctx context.Context) (StringIntMap, error)
	// TestContentParameter implements testContentParameter operation.
	//
	// GET /testContentParameter
	TestContentParameter(ctx context.Context, params TestContentParameterParams) (string, error)
	// TestFloatValidation implements testFloatValidation operation.
	//
	// POST /testFloatValidation
	TestFloatValidation(ctx context.Context, req TestFloatValidation) (TestFloatValidationOK, error)
	// TestNullableOneofs implements testNullableOneofs operation.
	//
	// GET /testNullableOneofs
	TestNullableOneofs(ctx context.Context) (TestNullableOneofsRes, error)
	// TestObjectQueryParameter implements testObjectQueryParameter operation.
	//
	// GET /testObjectQueryParameter
	TestObjectQueryParameter(ctx context.Context, params TestObjectQueryParameterParams) (TestObjectQueryParameterOK, error)
}

// Server implements http server based on OpenAPI v3 specification and
// calls Handler to handle requests.
type Server struct {
	h   Handler
	sec SecurityHandler
	baseServer
}

// NewServer creates new Server.
func NewServer(h Handler, sec SecurityHandler, opts ...Option) (*Server, error) {
	s, err := newConfig(opts...).baseServer()
	if err != nil {
		return nil, err
	}
	return &Server{
		h:          h,
		sec:        sec,
		baseServer: s,
	}, nil
}