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
		case '/': // Prefix: "/test"
			if l := len("/test"); len(elem) >= l && elem[0:l] == "/test" {
				elem = elem[l:]
			} else {
				break
			}

			if len(elem) == 0 {
				break
			}
			switch elem[0] {
			case 'F': // Prefix: "FormURLEncoded"
				if l := len("FormURLEncoded"); len(elem) >= l && elem[0:l] == "FormURLEncoded" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					// Leaf: TestFormURLEncoded
					s.handleTestFormURLEncodedRequest([0]string{}, w, r)

					return
				}
			case 'M': // Prefix: "Multipart"
				if l := len("Multipart"); len(elem) >= l && elem[0:l] == "Multipart" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					s.handleTestMultipartRequest([0]string{}, w, r)

					return
				}
				switch elem[0] {
				case 'U': // Prefix: "Upload"
					if l := len("Upload"); len(elem) >= l && elem[0:l] == "Upload" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						// Leaf: TestMultipartUpload
						s.handleTestMultipartUploadRequest([0]string{}, w, r)

						return
					}
				}
			case 'S': // Prefix: "ShareFormSchema"
				if l := len("ShareFormSchema"); len(elem) >= l && elem[0:l] == "ShareFormSchema" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					// Leaf: TestShareFormSchema
					s.handleTestShareFormSchemaRequest([0]string{}, w, r)

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
	case "POST":
		if len(elem) == 0 {
			break
		}
		switch elem[0] {
		case '/': // Prefix: "/test"
			if l := len("/test"); len(elem) >= l && elem[0:l] == "/test" {
				elem = elem[l:]
			} else {
				break
			}

			if len(elem) == 0 {
				break
			}
			switch elem[0] {
			case 'F': // Prefix: "FormURLEncoded"
				if l := len("FormURLEncoded"); len(elem) >= l && elem[0:l] == "FormURLEncoded" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					// Leaf: TestFormURLEncoded
					r.name = "TestFormURLEncoded"
					r.args = args
					r.count = 0
					return r, true
				}
			case 'M': // Prefix: "Multipart"
				if l := len("Multipart"); len(elem) >= l && elem[0:l] == "Multipart" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					r.name = "TestMultipart"
					r.args = args
					r.count = 0
					return r, true
				}
				switch elem[0] {
				case 'U': // Prefix: "Upload"
					if l := len("Upload"); len(elem) >= l && elem[0:l] == "Upload" {
						elem = elem[l:]
					} else {
						break
					}

					if len(elem) == 0 {
						// Leaf: TestMultipartUpload
						r.name = "TestMultipartUpload"
						r.args = args
						r.count = 0
						return r, true
					}
				}
			case 'S': // Prefix: "ShareFormSchema"
				if l := len("ShareFormSchema"); len(elem) >= l && elem[0:l] == "ShareFormSchema" {
					elem = elem[l:]
				} else {
					break
				}

				if len(elem) == 0 {
					// Leaf: TestShareFormSchema
					r.name = "TestShareFormSchema"
					r.args = args
					r.count = 0
					return r, true
				}
			}
		}
	}
	return r, false
}
