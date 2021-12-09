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

// AddPathItem adds the given PathItem under the given Name to the Paths of the Spec.
func (s *Spec) AddPathItem(n string, p *PathItem) *Spec {
	if p != nil {
		s.initPaths()
		s.Paths[n] = *p
	}
	return s
}

// AddNamedPaths adds the given namedPaths to the Paths of the Spec.
func (s *Spec) AddNamedPaths(ps ...*NamedPathItem) *Spec {
	for _, p := range ps {
		s.AddPathItem(p.Name, p.PathItem)
	}
	return s
}

// SetComponents sets the Components of the Spec.
func (s *Spec) SetComponents(c *Components) *Spec {
	s.Components = c
	return s
}

// initPaths ensures the Paths map is allocated.
func (s *Spec) initPaths() {
	if s.Paths == nil {
		s.Paths = make(Paths)
	}
}

// TODO: AddSchemas
// // AddSchema adds the given Schema under the given Name to the Components of the Spec.
// func (s *Spec) AddSchema(n string, sc *Schema) *Spec {
// 	if sc != nil {
// 		s.Init()
// 		s.Components.Schemas[n] = *sc
// 	}
// 	return s
// }
//
// // AddSchemas adds the given namedSchemas to the Components of the Spec.
// func (s *Spec) AddSchemas(scs ...*namedSchema) *Spec {
// 	for _, sc := range scs {
// 		s.AddSchema(sc.Name, sc.Schema)
// 	}
// }
// TODO: AddResponses

// AddParameter adds the given Parameter under the given Name to the Components of the Spec.
func (s *Spec) AddParameter(n string, p *Parameter) *Spec {
	if p != nil {
		s.initParameters()
		s.Components.Parameters[n] = *p
	}
	return s
}

// AddNamedParameters adds the given namedParameters to the Components of the Spec.
func (s *Spec) AddNamedParameters(ps ...*NamedParameter) *Spec {
	for _, p := range ps {
		s.AddParameter(p.Name, p.Parameter)
	}
	return s
}

// initParameters ensures the Parameters map is allocated.
func (s *Spec) initParameters() {
	s.initComponents()
	if s.Components.Parameters == nil {
		s.Components.Parameters = make(map[string]Parameter)
	}
}

// initComponents ensures the Components property is non-nil.
func (s *Spec) initComponents() {
	if s.Components == nil {
		s.Components = new(Components)
	}
}

// TODO: AddRequestBodies

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

// NewPathItem returns a new PathItem.
func NewPathItem() *PathItem {
	return new(PathItem)
}

// SetRef sets the Ref of the PathItem.
func (p *PathItem) SetRef(r string) *PathItem {
	p.Ref = r
	return p
}

// SetDescription sets the Description of the PathItem.
func (p *PathItem) SetDescription(d string) *PathItem {
	p.Description = d
	return p
}

// SetGet sets the Get of the PathItem.
func (p *PathItem) SetGet(o *Operation) *PathItem {
	p.Get = o
	return p
}

// SetPut sets the Put of the PathItem.
func (p *PathItem) SetPut(o *Operation) *PathItem {
	p.Put = o
	return p
}

// SetPost sets the Post of the PathItem.
func (p *PathItem) SetPost(o *Operation) *PathItem {
	p.Post = o
	return p
}

// SetDelete sets the Delete of the PathItem.
func (p *PathItem) SetDelete(o *Operation) *PathItem {
	p.Delete = o
	return p
}

// SetOptions sets the Options of the PathItem.
func (p *PathItem) SetOptions(o *Operation) *PathItem {
	p.Options = o
	return p
}

// SetHead sets the Head of the PathItem.
func (p *PathItem) SetHead(o *Operation) *PathItem {
	p.Head = o
	return p
}

// SetPatch sets the Patch of the PathItem.
func (p *PathItem) SetPatch(o *Operation) *PathItem {
	p.Patch = o
	return p
}

// SetTrace sets the Trace of the PathItem.
func (p *PathItem) SetTrace(o *Operation) *PathItem {
	p.Trace = o
	return p
}

// SetServers sets the Servers of the PathItem.
func (p *PathItem) SetServers(srvs []Server) *PathItem {
	p.Servers = srvs
	return p
}

// AddServers adds Servers to the Servers of the PathItem.
func (p *PathItem) AddServers(srvs ...*Server) *PathItem {
	for _, srv := range srvs {
		if srv != nil {
			p.Servers = append(p.Servers, *srv)
		}
	}
	return p
}

// SetParameters sets the Parameters of the PathItem.
func (p *PathItem) SetParameters(ps []Parameter) *PathItem {
	p.Parameters = ps
	return p
}

// AddParameters adds Parameters to the Parameters of the PathItem.
func (p *PathItem) AddParameters(ps ...*Parameter) *PathItem {
	for _, i := range ps {
		if i != nil {
			p.Parameters = append(p.Parameters, *i)
		}
	}
	return p
}

// ToNamed returns a NamedPathItem wrapping the receiver.
func (p *PathItem) ToNamed(n string) *NamedPathItem {
	return NewNamedPath(n, p)
}

// NamedPathItem can be used to construct a reference to the wrapped PathItem.
type NamedPathItem struct {
	PathItem *PathItem
	Name     string
}

// NewNamedPath returns a new NamedPathItem.
func NewNamedPath(n string, p *PathItem) *NamedPathItem {
	return &NamedPathItem{p, n}
}

// AsLocalRef returns a new PathItem referencing the wrapped PathItem in the local document.
func (p *NamedPathItem) AsLocalRef() *PathItem {
	return NewPathItem().SetRef("#/components/parameters/" + escapeRef(p.Name))
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

// InPath sets the In of the Parameter to "PathItem".
func (p *Parameter) InPath() *Parameter {
	return p.SetIn("PathItem")
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

// ToNamed returns a NamedParameter wrapping the receiver.
func (p *Parameter) ToNamed(n string) *NamedParameter {
	return NewNamedParameter(n, p)
}

// NamedParameter can be used to construct a reference to the wrapped Parameter.
type NamedParameter struct {
	Parameter *Parameter
	Name      string
}

// NewNamedParameter returns a new NamedParameter.
func NewNamedParameter(n string, p *Parameter) *NamedParameter {
	return &NamedParameter{p, n}
}

// AsLocalRef returns a new Parameter referencing the wrapped Parameter in the local document.
func (p *NamedParameter) AsLocalRef() *Parameter {
	return NewParameter().SetRef("#/components/parameters/" + escapeRef(p.Name))
}

// TODO: RequestBody
// TODO: Response
// TODO: Media
// TODO: Discriminator

// // namedSchema can be used to construct a reference to the wrapped Schema.
// type namedSchema struct {
// 	*Schema
// 	Name string
// }

// TODO: Property

func escapeRef(ref string) string {
	return strings.NewReplacer("~", "~0", "/", "~1").Replace(ref)
}
