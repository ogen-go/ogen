package api

type FoobarGetResponder interface {
	foobarGetResponder()
}

type FoobarPostResponder interface {
	foobarPostResponder()
}

type PetCreateRequester interface {
	petCreateRequester()
}

type PetGetResponder interface {
	petGetResponder()
}
