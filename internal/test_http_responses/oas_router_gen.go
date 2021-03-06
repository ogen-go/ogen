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
	case "GET":
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
			case 'a': // Prefix: "anyContentTypeBinaryStringSchema"
				if l := len("anyContentTypeBinaryStringSchema"); len(elem) >= l && elem[0:l] == "anyContentTypeBinaryStringSchema" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					s.handleAnyContentTypeBinaryStringSchemaRequest([0]string{}, w, r)

					return
				}
				switch elem[0] {
				case 'D': // Prefix: "Default"
					if l := len("Default"); len(elem) >= l && elem[0:l] == "Default" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						// Leaf: AnyContentTypeBinaryStringSchemaDefault
						s.handleAnyContentTypeBinaryStringSchemaDefaultRequest([0]string{}, w, r)

						return
					}
				}
			case 'c': // Prefix: "combined"
				if l := len("combined"); len(elem) >= l && elem[0:l] == "combined" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					// Leaf: Combined
					s.handleCombinedRequest([0]string{}, w, r)

					return
				}
			case 'h': // Prefix: "headers"
				if l := len("headers"); len(elem) >= l && elem[0:l] == "headers" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					break
				}
				switch elem[0] {
				case '2': // Prefix: "200"
					if l := len("200"); len(elem) >= l && elem[0:l] == "200" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						// Leaf: Headers200
						s.handleHeaders200Request([0]string{}, w, r)

						return
					}
				case 'C': // Prefix: "Combined"
					if l := len("Combined"); len(elem) >= l && elem[0:l] == "Combined" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						// Leaf: HeadersCombined
						s.handleHeadersCombinedRequest([0]string{}, w, r)

						return
					}
				case 'D': // Prefix: "Default"
					if l := len("Default"); len(elem) >= l && elem[0:l] == "Default" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						// Leaf: HeadersDefault
						s.handleHeadersDefaultRequest([0]string{}, w, r)

						return
					}
				case 'P': // Prefix: "Pattern"
					if l := len("Pattern"); len(elem) >= l && elem[0:l] == "Pattern" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						// Leaf: HeadersPattern
						s.handleHeadersPatternRequest([0]string{}, w, r)

						return
					}
				}
			case 'i': // Prefix: "intersectPatternCode"
				if l := len("intersectPatternCode"); len(elem) >= l && elem[0:l] == "intersectPatternCode" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					// Leaf: IntersectPatternCode
					s.handleIntersectPatternCodeRequest([0]string{}, w, r)

					return
				}
			case 'm': // Prefix: "multipleGenericResponses"
				if l := len("multipleGenericResponses"); len(elem) >= l && elem[0:l] == "multipleGenericResponses" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					// Leaf: MultipleGenericResponses
					s.handleMultipleGenericResponsesRequest([0]string{}, w, r)

					return
				}
			case 'o': // Prefix: "octetStream"
				if l := len("octetStream"); len(elem) >= l && elem[0:l] == "octetStream" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					break
				}
				switch elem[0] {
				case 'B': // Prefix: "BinaryStringSchema"
					if l := len("BinaryStringSchema"); len(elem) >= l && elem[0:l] == "BinaryStringSchema" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						// Leaf: OctetStreamBinaryStringSchema
						s.handleOctetStreamBinaryStringSchemaRequest([0]string{}, w, r)

						return
					}
				case 'E': // Prefix: "EmptySchema"
					if l := len("EmptySchema"); len(elem) >= l && elem[0:l] == "EmptySchema" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						// Leaf: OctetStreamEmptySchema
						s.handleOctetStreamEmptySchemaRequest([0]string{}, w, r)

						return
					}
				}
			case 't': // Prefix: "textPlainBinaryStringSchema"
				if l := len("textPlainBinaryStringSchema"); len(elem) >= l && elem[0:l] == "textPlainBinaryStringSchema" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					// Leaf: TextPlainBinaryStringSchema
					s.handleTextPlainBinaryStringSchemaRequest([0]string{}, w, r)

					return
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
	case "GET":
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
			case 'a': // Prefix: "anyContentTypeBinaryStringSchema"
				if l := len("anyContentTypeBinaryStringSchema"); len(elem) >= l && elem[0:l] == "anyContentTypeBinaryStringSchema" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					r.name = "AnyContentTypeBinaryStringSchema"
					r.args = args
					r.count = 0
					return r, true
				}
				switch elem[0] {
				case 'D': // Prefix: "Default"
					if l := len("Default"); len(elem) >= l && elem[0:l] == "Default" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						// Leaf: AnyContentTypeBinaryStringSchemaDefault
						r.name = "AnyContentTypeBinaryStringSchemaDefault"
						r.args = args
						r.count = 0
						return r, true
					}
				}
			case 'c': // Prefix: "combined"
				if l := len("combined"); len(elem) >= l && elem[0:l] == "combined" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					// Leaf: Combined
					r.name = "Combined"
					r.args = args
					r.count = 0
					return r, true
				}
			case 'h': // Prefix: "headers"
				if l := len("headers"); len(elem) >= l && elem[0:l] == "headers" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					break
				}
				switch elem[0] {
				case '2': // Prefix: "200"
					if l := len("200"); len(elem) >= l && elem[0:l] == "200" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						// Leaf: Headers200
						r.name = "Headers200"
						r.args = args
						r.count = 0
						return r, true
					}
				case 'C': // Prefix: "Combined"
					if l := len("Combined"); len(elem) >= l && elem[0:l] == "Combined" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						// Leaf: HeadersCombined
						r.name = "HeadersCombined"
						r.args = args
						r.count = 0
						return r, true
					}
				case 'D': // Prefix: "Default"
					if l := len("Default"); len(elem) >= l && elem[0:l] == "Default" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						// Leaf: HeadersDefault
						r.name = "HeadersDefault"
						r.args = args
						r.count = 0
						return r, true
					}
				case 'P': // Prefix: "Pattern"
					if l := len("Pattern"); len(elem) >= l && elem[0:l] == "Pattern" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						// Leaf: HeadersPattern
						r.name = "HeadersPattern"
						r.args = args
						r.count = 0
						return r, true
					}
				}
			case 'i': // Prefix: "intersectPatternCode"
				if l := len("intersectPatternCode"); len(elem) >= l && elem[0:l] == "intersectPatternCode" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					// Leaf: IntersectPatternCode
					r.name = "IntersectPatternCode"
					r.args = args
					r.count = 0
					return r, true
				}
			case 'm': // Prefix: "multipleGenericResponses"
				if l := len("multipleGenericResponses"); len(elem) >= l && elem[0:l] == "multipleGenericResponses" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					// Leaf: MultipleGenericResponses
					r.name = "MultipleGenericResponses"
					r.args = args
					r.count = 0
					return r, true
				}
			case 'o': // Prefix: "octetStream"
				if l := len("octetStream"); len(elem) >= l && elem[0:l] == "octetStream" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					break
				}
				switch elem[0] {
				case 'B': // Prefix: "BinaryStringSchema"
					if l := len("BinaryStringSchema"); len(elem) >= l && elem[0:l] == "BinaryStringSchema" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						// Leaf: OctetStreamBinaryStringSchema
						r.name = "OctetStreamBinaryStringSchema"
						r.args = args
						r.count = 0
						return r, true
					}
				case 'E': // Prefix: "EmptySchema"
					if l := len("EmptySchema"); len(elem) >= l && elem[0:l] == "EmptySchema" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						// Leaf: OctetStreamEmptySchema
						r.name = "OctetStreamEmptySchema"
						r.args = args
						r.count = 0
						return r, true
					}
				}
			case 't': // Prefix: "textPlainBinaryStringSchema"
				if l := len("textPlainBinaryStringSchema"); len(elem) >= l && elem[0:l] == "textPlainBinaryStringSchema" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					// Leaf: TextPlainBinaryStringSchema
					r.name = "TextPlainBinaryStringSchema"
					r.args = args
					r.count = 0
					return r, true
				}
			}
		}
	}
	return r, false
}
