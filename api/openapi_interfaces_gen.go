package api

type FoobarGetResponder interface {
	foobarGetResponder()
}

type FoobarPostResponder interface {
	foobarPostResponder()
}

type PetGetResponder interface {
	petGetResponder()
}

type PetPostRequester interface {
	petPostRequester()
}
