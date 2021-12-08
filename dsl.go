package ogen

import "strings"

// NewSpec returns a new Spec.
func NewSpec() *Spec {
	return new(Spec)
}

// SetOpenAPI sets the OpenAPI Specification version of the document.
func (s *Spec) SetOpenAPI(v string) *Spec {
	s.OpenAPI = v
	return s
}

// SetInfo sets the Info of the Spec.
func (s *Spec) SetInfo(i *Info) *Spec {
	if i != nil {
		s.Info = *i
	}
	return s
}

// SetServers sets the Servers of the Spec.
func (s *Spec) SetServers(srvs []Server) *Spec {
	s.Servers = srvs
	return s
}

// AddServers adds Servers to the Servers of the Spec.
func (s *Spec) AddServers(srvs ...*Server) *Spec {
	for _, srv := range srvs {
		if srv != nil {
			s.Servers = append(s.Servers, *srv)
		}
	}
	return s
}

// SetPaths sets the Paths of the Spec.
func (s *Spec) SetPaths(p Paths) *Spec {
	s.Paths = p
	return s
}

// AddPaths adds paths to the Paths of the Spec.
func (s *Spec) AddPaths(ps ...*path) *Spec {
	for _, p := range ps {
		if s.Paths == nil {
			s.Paths = make(Paths)
		}
		s.Paths[p.path] = *p.item
	}
	return s
}

// NewInfo returns a new Info.
func NewInfo() *Info {
	return new(Info)
}

// SetTitle sets the title of the Info.
func (i *Info) SetTitle(t string) *Info {
	i.Title = t
	return i
}

// SetDescription sets the description of the Info.
func (i *Info) SetDescription(d string) *Info {
	i.Description = d
	return i
}

// SetTermsOfService sets the terms of service of the Info.
func (i *Info) SetTermsOfService(t string) *Info {
	i.TermsOfService = t
	return i
}

// SetContact sets the Contact of the Info.
func (i *Info) SetContact(c *Contact) *Info {
	i.Contact = c
	return i
}

// SetLicense sets the License of the Info.
func (i *Info) SetLicense(l *License) *Info {
	i.License = l
	return i
}

// SetVersion sets the version of the Info.
func (i *Info) SetVersion(v string) *Info {
	i.Version = v
	return i
}

// NewContact returns a new Contact.
func NewContact() *Contact {
	return new(Contact)
}

// SetName sets the Name of the Contact.
func (c *Contact) SetName(n string) *Contact {
	c.Name = n
	return c
}

// SetURL sets the URL of the Contact.
func (c *Contact) SetURL(url string) *Contact {
	c.URL = url
	return c
}

// SetEmail sets the Email of the Contact.
func (c *Contact) SetEmail(e string) *Contact {
	c.Email = e
	return c
}

// NewLicense returns a new License.
func NewLicense() *License {
	return new(License)
}

// SetName sets the Name of the License.
func (l *License) SetName(n string) *License {
	l.Name = n
	return l
}

// SetURL sets the URL of the License.
func (l *License) SetURL(url string) *License {
	l.URL = url
	return l
}

// NewServer returns a new Server.
func NewServer() *Server {
	return new(Server)
}

// SetDescription sets the Description of the Server.
func (s *Server) SetDescription(d string) *Server {
	s.Description = d
	return s
}

// SetURL sets the URL of the Server.
func (s *Server) SetURL(url string) *Server {
	s.URL = url
	return s
}

// TODO: Components

// path holds the PathItem for a single path.
type path struct {
	path string
	item *PathItem
}

// NewPath returns a new path.
func NewPath(p string) *path {
	return &path{p, new(PathItem)}
}

// SetRef sets the Ref of the PathItem.
func (p *path) SetRef(r string) *path {
	p.item.Ref = r
	return p
}

// SetDescription sets the Description of the PathItem.
func (p *path) SetDescription(d string) *path {
	p.item.Description = d
	return p
}

// SetGet sets the Get of the PathItem.
func (p *path) SetGet(o *Operation) *path {
	p.item.Get = o
	return p
}

// SetPut sets the Put of the PathItem.
func (p *path) SetPut(o *Operation) *path {
	p.item.Put = o
	return p
}

// SetPost sets the Post of the PathItem.
func (p *path) SetPost(o *Operation) *path {
	p.item.Post = o
	return p
}

// SetDelete sets the Delete of the PathItem.
func (p *path) SetDelete(o *Operation) *path {
	p.item.Delete = o
	return p
}

// SetOptions sets the Options of the PathItem.
func (p *path) SetOptions(o *Operation) *path {
	p.item.Options = o
	return p
}

// SetHead sets the Head of the PathItem.
func (p *path) SetHead(o *Operation) *path {
	p.item.Head = o
	return p
}

// SetPatch sets the Patch of the PathItem.
func (p *path) SetPatch(o *Operation) *path {
	p.item.Patch = o
	return p
}

// SetTrace sets the Trace of the PathItem.
func (p *path) SetTrace(o *Operation) *path {
	p.item.Trace = o
	return p
}

// SetServers sets the Servers of the PathItem.
func (p *path) SetServers(srvs []Server) *path {
	p.item.Servers = srvs
	return p
}

// AddServers adds Servers to the Servers of the PathItem.
func (p *path) AddServers(srvs ...*Server) *path {
	for _, srv := range srvs {
		if srv != nil {
			p.item.Servers = append(p.item.Servers, *srv)
		}
	}
	return p
}

// SetParameters sets the Parameters of the PathItem.
func (p *path) SetParameters(ps []Parameter) *path {
	p.item.Parameters = ps
	return p
}

// AddParameters adds Parameters to the Parameters of the PathItem.
func (p *path) AddParameters(ps ...*Parameter) *path {
	for _, i := range ps {
		if i != nil {
			p.item.Parameters = append(p.item.Parameters, *i)
		}
	}
	return p
}

// LocalRef returns the ref for the path in the local document.
func (p *path) LocalRef() string {
	return "#/paths/" + strings.NewReplacer("~", "~0", "/", "~1").Replace(p.path)
}

// NewOperation returns a new Operation.
func NewOperation() *Operation {
	return new(Operation)
}

// SetTags sets the Tags of the Operation.
func (o *Operation) SetTags(ts []string) *Operation {
	o.Tags = ts
	return o
}

// AddTags adds Tags to the Tags of the Operation.
func (o *Operation) AddTags(ts ...string) *Operation {
	o.Tags = append(o.Tags, ts...)
	return o
}

// SetDescription sets the Description of the Operation.
func (o *Operation) SetDescription(d string) *Operation {
	o.Description = d
	return o
}

// SetOperationID sets the OperationID of the Operation.
func (o *Operation) SetOperationID(id string) *Operation {
	o.OperationID = id
	return o
}

// SetParameters sets the Parameters of the Operation.
func (o *Operation) SetParameters(ps []Parameter) *Operation {
	o.Parameters = ps
	return o
}

// AddParameters adds Parameters to the Parameters of the Operation.
func (o *Operation) AddParameters(ps ...*Parameter) *Operation {
	for _, p := range ps {
		if p != nil {
			o.Parameters = append(o.Parameters, *p)
		}
	}
	return o
}

// NewParameter returns a new Parameter.
func NewParameter() *Parameter {
	return new(Parameter)
}

// SetRef sets the Ref of the Parameter.
func (p *Parameter) SetRef(r string) *Parameter {
	p.Ref = r
	return p
}

// SetName sets the Name of the Parameter.
func (p *Parameter) SetName(n string) *Parameter {
	p.Name = n
	return p
}

// SetIn sets the In of the Parameter.
func (p *Parameter) SetIn(i string) *Parameter {
	p.In = i
	return p
}

// InPath sets the In of the Parameter to "path".
func (p *Parameter) InPath() *Parameter {
	return p.SetIn("path")
}

// InQuery sets the In of the Parameter to "query".
func (p *Parameter) InQuery() *Parameter {
	return p.SetIn("query")
}

// InHeader sets the In of the Parameter to "header".
func (p *Parameter) InHeader() *Parameter {
	return p.SetIn("header")
}

// InCookie sets the In of the Parameter to "cookie".
func (p *Parameter) InCookie() *Parameter {
	return p.SetIn("cookie")
}

// SetDescription sets the Description of the Parameter.
func (p *Parameter) SetDescription(d string) *Parameter {
	p.Description = d
	return p
}

// SetSchema sets the Schema of the Parameter.
func (p *Parameter) SetSchema(s Schema) *Parameter {
	p.Schema = s
	return p
}

// SetRequired sets the Required of the Parameter.
func (p *Parameter) SetRequired(r bool) *Parameter {
	p.Required = r
	return p
}

// SetDeprecated sets the Deprecated of the Parameter.
func (p *Parameter) SetDeprecated(d bool) *Parameter {
	p.Deprecated = d
	return p
}

// SetContent sets the Content of the Parameter.
func (p *Parameter) SetContent(c map[string]Media) *Parameter {
	p.Content = c
	return p
}

// SetStyle sets the Style of the Parameter.
func (p *Parameter) SetStyle(s string) *Parameter {
	p.Style = s
	return p
}

// SetExplode sets the Explode of the Parameter.
func (p *Parameter) SetExplode(e bool) *Parameter {
	p.Explode = &e
	return p
}

// ToNamedParameter returns a namedParameter wrapping the receiver.
func (p *Parameter) ToNamedParameter(n string) *namedParameter {
	return NewNamedParameter(n, p)
}

// namedParameter can be used to construct a Reference to the wrapped Parameter.
type namedParameter struct {
	*Parameter
	name string
}

// NewNamedParameter returns a new namedParameter.
func NewNamedParameter(n string, p *Parameter) *namedParameter {
	return &namedParameter{p, n}
}

// LocalRef returns the ref for the Parameter in the local document.
func (p *namedParameter) LocalRef() string {
	return "#/components/parameters/" + escapeRef(p.name)
}

// TODO: Parameter
// TODO: RequestBody
// TODO: Response
// TODO: Media
// TODO: Discriminator
// TODO: Schema
// TODO: Property

func escapeRef(ref string) string {
	return strings.NewReplacer("~", "~0", "/", "~1").Replace(ref)
}
