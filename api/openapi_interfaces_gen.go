package api

type FoobarGetResponse interface {
	implFoobarGetResponse()
}

type FoobarPostResponse interface {
	implFoobarPostResponse()
}

type PetPostRequest interface {
	implPetPostRequest()
}
