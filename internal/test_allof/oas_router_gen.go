// Code generated by ogen, DO NOT EDIT.

package api

import (
	"net/http"
)

func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	s.cfg.NotFound(w, r)
}

// ServeHTTP serves http request as defined by OpenAPI v3 specification,
// calling handler that matches the path or returning not found error.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	elem := r.URL.Path
	if len(elem) == 0 {
		s.notFound(w, r)
		return
	}
	// Static code generated router with unwrapped path search.
	switch r.Method {
	case "POST":
		if len(elem) == 0 {
			break
		}
		switch elem[0] {
		case '/': // Prefix: "/"
			if l := len("/"); len(elem) >= l && elem[0:l] == "/" {
				elem = elem[l:]
			} else {
				break
			}

			if len(elem) == 0 {
				break
			}
			switch elem[0] {
			case 'n': // Prefix: "nullableStrings"
				if l := len("nullableStrings"); len(elem) >= l && elem[0:l] == "nullableStrings" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					// Leaf: NullableStrings
					s.handleNullableStringsRequest([0]string{}, w, r)

					return
				}
			case 'o': // Prefix: "objectsWithConflicting"
				if l := len("objectsWithConflicting"); len(elem) >= l && elem[0:l] == "objectsWithConflicting" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					break
				}
				switch elem[0] {
				case 'A': // Prefix: "ArrayProperty"
					if l := len("ArrayProperty"); len(elem) >= l && elem[0:l] == "ArrayProperty" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						// Leaf: ObjectsWithConflictingArrayProperty
						s.handleObjectsWithConflictingArrayPropertyRequest([0]string{}, w, r)

						return
					}
				case 'P': // Prefix: "Properties"
					if l := len("Properties"); len(elem) >= l && elem[0:l] == "Properties" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						// Leaf: ObjectsWithConflictingProperties
						s.handleObjectsWithConflictingPropertiesRequest([0]string{}, w, r)

						return
					}
				}
			case 'r': // Prefix: "referencedAllof"
				if l := len("referencedAllof"); len(elem) >= l && elem[0:l] == "referencedAllof" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					s.handleReferencedAllofRequest([0]string{}, w, r)

					return
				}
				switch elem[0] {
				case 'O': // Prefix: "Optional"
					if l := len("Optional"); len(elem) >= l && elem[0:l] == "Optional" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						// Leaf: ReferencedAllofOptional
						s.handleReferencedAllofOptionalRequest([0]string{}, w, r)

						return
					}
				}
			case 's': // Prefix: "simple"
				if l := len("simple"); len(elem) >= l && elem[0:l] == "simple" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					break
				}
				switch elem[0] {
				case 'I': // Prefix: "Integer"
					if l := len("Integer"); len(elem) >= l && elem[0:l] == "Integer" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						// Leaf: SimpleInteger
						s.handleSimpleIntegerRequest([0]string{}, w, r)

						return
					}
				case 'O': // Prefix: "Objects"
					if l := len("Objects"); len(elem) >= l && elem[0:l] == "Objects" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						// Leaf: SimpleObjects
						s.handleSimpleObjectsRequest([0]string{}, w, r)

						return
					}
				}
			}
		}
	}
	s.notFound(w, r)
}

// Route is route object.
type Route struct {
	name  string
	count int
	args  [0]string
}

// OperationID returns OpenAPI operationId.
func (r Route) OperationID() string {
	return r.name
}

// Args returns parsed arguments.
func (r Route) Args() []string {
	return r.args[:r.count]
}

// FindRoute finds Route for given method and path.
func (s *Server) FindRoute(method, path string) (r Route, _ bool) {
	var (
		args = [0]string{}
		elem = path
	)
	r.args = args
	if elem == "" {
		return r, false
	}

	// Static code generated router with unwrapped path search.
	switch method {
	case "POST":
		if len(elem) == 0 {
			break
		}
		switch elem[0] {
		case '/': // Prefix: "/"
			if l := len("/"); len(elem) >= l && elem[0:l] == "/" {
				elem = elem[l:]
			} else {
				break
			}

			if len(elem) == 0 {
				break
			}
			switch elem[0] {
			case 'n': // Prefix: "nullableStrings"
				if l := len("nullableStrings"); len(elem) >= l && elem[0:l] == "nullableStrings" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					// Leaf: NullableStrings
					r.name = "NullableStrings"
					r.args = args
					r.count = 0
					return r, true
				}
			case 'o': // Prefix: "objectsWithConflicting"
				if l := len("objectsWithConflicting"); len(elem) >= l && elem[0:l] == "objectsWithConflicting" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					break
				}
				switch elem[0] {
				case 'A': // Prefix: "ArrayProperty"
					if l := len("ArrayProperty"); len(elem) >= l && elem[0:l] == "ArrayProperty" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						// Leaf: ObjectsWithConflictingArrayProperty
						r.name = "ObjectsWithConflictingArrayProperty"
						r.args = args
						r.count = 0
						return r, true
					}
				case 'P': // Prefix: "Properties"
					if l := len("Properties"); len(elem) >= l && elem[0:l] == "Properties" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						// Leaf: ObjectsWithConflictingProperties
						r.name = "ObjectsWithConflictingProperties"
						r.args = args
						r.count = 0
						return r, true
					}
				}
			case 'r': // Prefix: "referencedAllof"
				if l := len("referencedAllof"); len(elem) >= l && elem[0:l] == "referencedAllof" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					r.name = "ReferencedAllof"
					r.args = args
					r.count = 0
					return r, true
				}
				switch elem[0] {
				case 'O': // Prefix: "Optional"
					if l := len("Optional"); len(elem) >= l && elem[0:l] == "Optional" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						// Leaf: ReferencedAllofOptional
						r.name = "ReferencedAllofOptional"
						r.args = args
						r.count = 0
						return r, true
					}
				}
			case 's': // Prefix: "simple"
				if l := len("simple"); len(elem) >= l && elem[0:l] == "simple" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					break
				}
				switch elem[0] {
				case 'I': // Prefix: "Integer"
					if l := len("Integer"); len(elem) >= l && elem[0:l] == "Integer" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						// Leaf: SimpleInteger
						r.name = "SimpleInteger"
						r.args = args
						r.count = 0
						return r, true
					}
				case 'O': // Prefix: "Objects"
					if l := len("Objects"); len(elem) >= l && elem[0:l] == "Objects" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						// Leaf: SimpleObjects
						r.name = "SimpleObjects"
						r.args = args
						r.count = 0
						return r, true
					}
				}
			}
		}
	}
	return r, false
}
