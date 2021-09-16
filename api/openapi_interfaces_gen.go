package api

type FoobarGetResponse interface {
	foobarGetResponse()
}

type FoobarPostResponse interface {
	foobarPostResponse()
}

type PetGetResponse interface {
	petGetResponse()
}

type PetPostRequest interface {
	petPostRequest()
}
