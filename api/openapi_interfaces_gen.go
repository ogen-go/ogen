package api

type FoobarGetResponse interface {
	implFoobarGetResponse()
}

type FoobarPostResponse interface {
	implFoobarPostResponse()
}

type PetGetResponse interface {
	implPetGetResponse()
}

type PetPostRequest interface {
	implPetPostRequest()
}
