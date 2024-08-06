package ogen

import (
	"encoding/json"
	"slices"
	"strings"

	"github.com/go-faster/jx"

	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/openapi"
)

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
	s.Servers = slices.Clone(srvs)
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
	s.initPaths()
	s.Paths[n] = p
	return s
}

// AddNamedPathItems adds the given namedPaths to the Paths of the Spec.
func (s *Spec) AddNamedPathItems(ps ...*NamedPathItem) *Spec {
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

// AddSchema adds the given Schema under the given Name to the Components of the Spec.
func (s *Spec) AddSchema(n string, sc *Schema) *Spec {
	s.initSchemas()
	s.Components.Schemas[n] = sc
	return s
}

// AddNamedSchemas adds the given namedSchemas to the Components of the Spec.
func (s *Spec) AddNamedSchemas(scs ...*NamedSchema) *Spec {
	for _, sc := range scs {
		s.AddSchema(sc.Name, sc.Schema)
	}
	return s
}

// AddResponse adds the given Response under the given Name to the Components of the Spec.
func (s *Spec) AddResponse(n string, sc *Response) *Spec {
	s.initResponses()
	s.Components.Responses[n] = sc
	return s
}

// AddNamedResponses adds the given namedResponses to the Components of the Spec.
func (s *Spec) AddNamedResponses(scs ...*NamedResponse) *Spec {
	for _, sc := range scs {
		s.AddResponse(sc.Name, sc.Response)
	}
	return s
}

// AddParameter adds the given Parameter under the given Name to the Components of the Spec.
func (s *Spec) AddParameter(n string, p *Parameter) *Spec {
	s.initParameters()
	s.Components.Parameters[n] = p
	return s
}

// AddNamedParameters adds the given namedParameters to the Components of the Spec.
func (s *Spec) AddNamedParameters(ps ...*NamedParameter) *Spec {
	for _, p := range ps {
		s.AddParameter(p.Name, p.Parameter)
	}
	return s
}

// AddRequestBody adds the given RequestBody under the given Name to the Components of the Spec.
func (s *Spec) AddRequestBody(n string, sc *RequestBody) *Spec {
	s.initRequestBodies()
	s.Components.RequestBodies[n] = sc
	return s
}

// AddNamedRequestBodies adds the given namedRequestBodies to the Components of the Spec.
func (s *Spec) AddNamedRequestBodies(scs ...*NamedRequestBody) *Spec {
	for _, sc := range scs {
		s.AddRequestBody(sc.Name, sc.RequestBody)
	}
	return s
}

// RefSchema returns a new Schema referencing the given name.
func (s *Spec) RefSchema(n string) *NamedSchema {
	if s.Components != nil && s.Components.Schemas != nil {
		if r, ok := s.Components.Schemas[n]; ok {
			return NewNamedSchema(n, r).AsLocalRef().ToNamed(n)
		}
	}
	return nil
}

// RefResponse returns a new Response referencing the given name.
func (s *Spec) RefResponse(n string) *NamedResponse {
	if s.Components != nil && s.Components.Responses != nil {
		if r, ok := s.Components.Responses[n]; ok {
			return NewNamedResponse(n, r).AsLocalRef().ToNamed(n)
		}
	}
	return nil
}

// RefRequestBody returns a new RequestBody referencing the given name.
func (s *Spec) RefRequestBody(n string) *NamedRequestBody {
	if s.Components != nil && s.Components.RequestBodies != nil {
		if r, ok := s.Components.RequestBodies[n]; ok {
			return NewNamedRequestBody(n, r).AsLocalRef().ToNamed(n)
		}
	}
	return nil
}

// initComponents ensures the Components property is non-nil.
func (s *Spec) initComponents() {
	if s.Components == nil {
		s.Components = new(Components)
	}
}

// initParameters ensures the Parameters map is allocated.
func (s *Spec) initParameters() {
	s.initComponents()
	if s.Components.Parameters == nil {
		s.Components.Parameters = make(map[string]*Parameter)
	}
}

// initSchemas ensures the Schemas map is allocated.
func (s *Spec) initSchemas() {
	s.initComponents()
	if s.Components.Schemas == nil {
		s.Components.Schemas = make(map[string]*Schema)
	}
}

// initResponses ensures the Responses map is allocated.
func (s *Spec) initResponses() {
	s.initComponents()
	if s.Components.Responses == nil {
		s.Components.Responses = make(map[string]*Response)
	}
}

// initRequestBodies ensures the RequestBodies map is allocated.
func (s *Spec) initRequestBodies() {
	s.initComponents()
	if s.Components.RequestBodies == nil {
		s.Components.RequestBodies = make(map[string]*RequestBody)
	}
}

// NewRequestBody returns a new RequestBody.
func NewRequestBody() *RequestBody {
	return new(RequestBody)
}

// SetRef sets the Ref of the RequestBody.
func (r *RequestBody) SetRef(ref string) *RequestBody {
	r.Ref = ref
	return r
}

// SetDescription sets the Description of the RequestBody.
func (r *RequestBody) SetDescription(d string) *RequestBody {
	r.Description = d
	return r
}

// SetContent sets the Content of the RequestBody.
func (r *RequestBody) SetContent(c map[string]Media) *RequestBody {
	r.Content = c
	return r
}

// AddContent adds the given Schema under the MediaType to the Content of the Response.
func (r *RequestBody) AddContent(mt string, s *Schema) *RequestBody {
	if s != nil {
		r.initContent()
		r.Content[mt] = Media{Schema: s}
	}
	return r
}

// SetJSONContent sets the given Schema under the JSON MediaType to the Content of the Response.
func (r *RequestBody) SetJSONContent(s *Schema) *RequestBody {
	return r.AddContent("application/json", s)
}

// initContent ensures the Content map is allocated.
func (r *RequestBody) initContent() {
	if r.Content == nil {
		r.Content = make(map[string]Media)
	}
}

// SetRequired sets the Required of the RequestBody.
func (r *RequestBody) SetRequired(req bool) *RequestBody {
	r.Required = req
	return r
}

// ToNamed returns a NamedRequestBody wrapping the receiver.
func (r *RequestBody) ToNamed(n string) *NamedRequestBody {
	return NewNamedRequestBody(n, r)
}

// NamedRequestBody can be used to construct a reference to the wrapped RequestBody.
type NamedRequestBody struct {
	RequestBody *RequestBody
	Name        string
}

// NewNamedRequestBody returns a new NamedRequestBody.
func NewNamedRequestBody(n string, p *RequestBody) *NamedRequestBody {
	return &NamedRequestBody{p, n}
}

// AsLocalRef returns a new RequestBody referencing the wrapped RequestBody in the local document.
func (p *NamedRequestBody) AsLocalRef() *RequestBody {
	return NewRequestBody().SetRef("#/components/requestBodies/" + escapeRef(p.Name))
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
	p.Servers = slices.Clone(srvs)
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
func (p *PathItem) SetParameters(ps []*Parameter) *PathItem {
	p.Parameters = slices.Clone(ps)
	return p
}

// AddParameters adds Parameters to the Parameters of the PathItem.
func (p *PathItem) AddParameters(ps ...*Parameter) *PathItem {
	p.Parameters = append(p.Parameters, ps...)
	return p
}

// ToNamed returns a NamedPathItem wrapping the receiver.
func (p *PathItem) ToNamed(n string) *NamedPathItem {
	return NewNamedPathItem(n, p)
}

// NamedPathItem can be used to construct a reference to the wrapped PathItem.
type NamedPathItem struct {
	PathItem *PathItem
	Name     string
}

// NewNamedPathItem returns a new NamedPathItem.
func NewNamedPathItem(n string, p *PathItem) *NamedPathItem {
	return &NamedPathItem{p, n}
}

// AsLocalRef returns a new PathItem referencing the wrapped PathItem in the local document.
func (p *NamedPathItem) AsLocalRef() *PathItem {
	return NewPathItem().SetRef("#/paths/" + escapeRef(p.Name))
}

// NewOperation returns a new Operation.
func NewOperation() *Operation {
	return new(Operation)
}

// SetTags sets the Tags of the Operation.
func (o *Operation) SetTags(ts []string) *Operation {
	o.Tags = slices.Clone(ts)
	return o
}

// AddTags adds Tags to the Tags of the Operation.
func (o *Operation) AddTags(ts ...string) *Operation {
	o.Tags = append(o.Tags, ts...)
	return o
}

// SetSummary sets the Summary of the Operation.
func (o *Operation) SetSummary(s string) *Operation {
	o.Summary = s
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
func (o *Operation) SetParameters(ps []*Parameter) *Operation {
	o.Parameters = slices.Clone(ps)
	return o
}

// AddParameters adds Parameters to the Parameters of the Operation.
func (o *Operation) AddParameters(ps ...*Parameter) *Operation {
	o.Parameters = append(o.Parameters, ps...)
	return o
}

// SetRequestBody sets the RequestBody of the Operation.
func (o *Operation) SetRequestBody(r *RequestBody) *Operation {
	o.RequestBody = r
	return o
}

// SetResponses sets the Responses of the Operation.
func (o *Operation) SetResponses(r Responses) *Operation {
	o.Responses = r
	return o
}

// AddResponse adds the given Response under the given Name to the Responses of the Operation.
func (o *Operation) AddResponse(n string, p *Response) *Operation {
	o.initResponses()
	o.Responses[n] = p
	return o
}

// AddNamedResponses adds the given namedResponses to the Responses of the Operation.
func (o *Operation) AddNamedResponses(ps ...*NamedResponse) *Operation {
	for _, p := range ps {
		o.AddResponse(p.Name, p.Response)
	}
	return o
}

// initResponses ensures the Responses map is allocated.
func (o *Operation) initResponses() {
	if o.Responses == nil {
		o.Responses = make(Responses)
	}
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
	return p.SetIn(openapi.LocationPath.String())
}

// InQuery sets the In of the Parameter to "query".
func (p *Parameter) InQuery() *Parameter {
	return p.SetIn(openapi.LocationQuery.String())
}

// InHeader sets the In of the Parameter to "header".
func (p *Parameter) InHeader() *Parameter {
	return p.SetIn(openapi.LocationHeader.String())
}

// InCookie sets the In of the Parameter to "cookie".
func (p *Parameter) InCookie() *Parameter {
	return p.SetIn(openapi.LocationCookie.String())
}

// SetDescription sets the Description of the Parameter.
func (p *Parameter) SetDescription(d string) *Parameter {
	p.Description = d
	return p
}

// SetSchema sets the Schema of the Parameter.
func (p *Parameter) SetSchema(s *Schema) *Parameter {
	if s != nil {
		p.Schema = s
	}
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

// TODO(masseelch): Add Content helpers for Parameter

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

// NewResponse returns a new Response.
func NewResponse() *Response {
	return new(Response)
}

// SetRef sets the Ref of the Response.
func (r *Response) SetRef(ref string) *Response {
	r.Ref = ref
	return r
}

// SetDescription sets the Description of the Response.
func (r *Response) SetDescription(d string) *Response {
	r.Description = d
	return r
}

// SetHeaders sets the Headers of the Response.
func (r *Response) SetHeaders(h map[string]*Header) *Response {
	r.Headers = h
	return r
}

// SetContent sets the Content of the Response.
func (r *Response) SetContent(c map[string]Media) *Response {
	r.Content = c
	return r
}

// AddContent adds the given Schema under the MediaType to the Content of the Response.
func (r *Response) AddContent(mt string, s *Schema) *Response {
	if s != nil {
		r.initContent()
		r.Content[mt] = Media{Schema: s}
	}
	return r
}

// SetJSONContent sets the given Schema under the JSON MediaType to the Content of the Response.
func (r *Response) SetJSONContent(s *Schema) *Response {
	return r.AddContent("application/json", s)
}

// initContent ensures the Content map is allocated.
func (r *Response) initContent() {
	if r.Content == nil {
		r.Content = make(map[string]Media)
	}
}

// SetLinks sets the Links of the Response.
func (r *Response) SetLinks(l map[string]*Link) *Response {
	r.Links = l
	return r
}

// ToNamed returns a NamedResponse wrapping the receiver.
func (r *Response) ToNamed(n string) *NamedResponse {
	return NewNamedResponse(n, r)
}

// NamedResponse can be used to construct a reference to the wrapped Response.
type NamedResponse struct {
	Response *Response
	Name     string
}

// NewNamedResponse returns a new NamedResponse.
func NewNamedResponse(n string, p *Response) *NamedResponse {
	return &NamedResponse{p, n}
}

// AsLocalRef returns a new Response referencing the wrapped Response in the local document.
func (p *NamedResponse) AsLocalRef() *Response {
	return NewResponse().SetRef("#/components/responses/" + escapeRef(p.Name))
}

// TODO(masseelch): Discriminator

// NewSchema returns a new Schema.
func NewSchema() *Schema {
	return new(Schema)
}

// SetRef sets the Ref of the Schema.
func (s *Schema) SetRef(r string) *Schema {
	s.Ref = r
	return s
}

// SetSummary sets the Summary of the Schema.
func (s *Schema) SetSummary(smry string) *Schema {
	s.Summary = smry
	return s
}

// SetDescription sets the Description of the Schema.
func (s *Schema) SetDescription(d string) *Schema {
	s.Description = d
	return s
}

// SetType sets the Type of the Schema.
func (s *Schema) SetType(t string) *Schema {
	s.Type = t
	return s
}

// SetFormat sets the Format of the Schema.
func (s *Schema) SetFormat(f string) *Schema {
	s.Format = f
	return s
}

// SetProperties sets the Properties of the Schema.
func (s *Schema) SetProperties(p *Properties) *Schema {
	s.SetType("object")
	if p != nil {
		s.Properties = *p
	}
	return s
}

// AddOptionalProperties adds the Properties to the Properties of the Schema.
func (s *Schema) AddOptionalProperties(ps ...*Property) *Schema {
	s.SetType("object")
	for _, p := range ps {
		if p != nil {
			s.Properties = append(s.Properties, *p)
		}
	}
	return s
}

// AddRequiredProperties adds the Properties to the Properties of the Schema and marks them as required.
func (s *Schema) AddRequiredProperties(ps ...*Property) *Schema {
	s.AddOptionalProperties(ps...)
	for _, p := range ps {
		if p != nil {
			s.Required = append(s.Required, p.Name)
		}
	}
	return s
}

// SetRequired sets the Required of the Schema.
func (s *Schema) SetRequired(r []string) *Schema {
	s.Required = slices.Clone(r)
	return s
}

// SetItems sets the Items of the Schema.
func (s *Schema) SetItems(i *Schema) *Schema {
	s.Items = &Items{
		Item: i,
	}
	return s
}

// SetNullable sets the Nullable of the Schema.
func (s *Schema) SetNullable(n bool) *Schema {
	s.Nullable = n
	return s
}

// SetAllOf sets the AllOf of the Schema.
func (s *Schema) SetAllOf(a []*Schema) *Schema {
	s.AllOf = slices.Clone(a)
	return s
}

// SetOneOf sets the OneOf of the Schema.
func (s *Schema) SetOneOf(o []*Schema) *Schema {
	s.OneOf = slices.Clone(o)
	return s
}

// SetAnyOf sets the AnyOf of the Schema.
func (s *Schema) SetAnyOf(a []*Schema) *Schema {
	s.AnyOf = slices.Clone(a)
	return s
}

// SetDiscriminator sets the Discriminator of the Schema.
func (s *Schema) SetDiscriminator(d *Discriminator) *Schema {
	s.Discriminator = d
	return s
}

// SetEnum sets the Enum of the Schema.
func (s *Schema) SetEnum(e []json.RawMessage) *Schema {
	for _, val := range e {
		s.Enum = append(s.Enum, val)
	}
	return s
}

// SetMultipleOf sets the MultipleOf of the Schema.
func (s *Schema) SetMultipleOf(m *uint64) *Schema {
	if m != nil {
		val := *m
		e := &jx.Encoder{}
		e.UInt64(val)
		s.MultipleOf = e.Bytes()
	}
	return s
}

// SetMaximum sets the Maximum of the Schema.
func (s *Schema) SetMaximum(m *int64) *Schema {
	if m != nil {
		val := *m
		e := &jx.Encoder{}
		e.Int64(val)
		s.Maximum = e.Bytes()
	}
	return s
}

// SetExclusiveMaximum sets the ExclusiveMaximum of the Schema.
func (s *Schema) SetExclusiveMaximum(e bool) *Schema {
	s.ExclusiveMaximum = e
	return s
}

// SetMinimum sets the Minimum of the Schema.
func (s *Schema) SetMinimum(m *int64) *Schema {
	if m != nil {
		val := *m
		e := &jx.Encoder{}
		e.Int64(val)
		s.Minimum = e.Bytes()
	}
	return s
}

// SetExclusiveMinimum sets the ExclusiveMinimum of the Schema.
func (s *Schema) SetExclusiveMinimum(e bool) *Schema {
	s.ExclusiveMinimum = e
	return s
}

// SetMaxLength sets the MaxLength of the Schema.
func (s *Schema) SetMaxLength(m *uint64) *Schema {
	s.MaxLength = m
	return s
}

// SetMinLength sets the MinLength of the Schema.
func (s *Schema) SetMinLength(m *uint64) *Schema {
	s.MinLength = m
	return s
}

// SetPattern sets the Pattern of the Schema.
func (s *Schema) SetPattern(p string) *Schema {
	s.Pattern = p
	return s
}

// SetMaxItems sets the MaxItems of the Schema.
func (s *Schema) SetMaxItems(m *uint64) *Schema {
	s.MaxItems = m
	return s
}

// SetMinItems sets the MinItems of the Schema.
func (s *Schema) SetMinItems(m *uint64) *Schema {
	s.MinItems = m
	return s
}

// SetUniqueItems sets the UniqueItems of the Schema.
func (s *Schema) SetUniqueItems(u bool) *Schema {
	s.UniqueItems = u
	return s
}

// SetMaxProperties sets the MaxProperties of the Schema.
func (s *Schema) SetMaxProperties(m *uint64) *Schema {
	s.MaxProperties = m
	return s
}

// SetMinProperties sets the MinProperties of the Schema.
func (s *Schema) SetMinProperties(m *uint64) *Schema {
	s.MinProperties = m
	return s
}

// SetDefault sets the Default of the Schema.
func (s *Schema) SetDefault(d json.RawMessage) *Schema {
	s.Default = Default(d)
	return s
}

// SetDeprecated sets the Deprecated of the Schema.
func (s *Schema) SetDeprecated(d bool) *Schema {
	s.Deprecated = d
	return s
}

// ToNamed returns a NamedSchema wrapping the receiver.
func (s *Schema) ToNamed(n string) *NamedSchema {
	return NewNamedSchema(n, s)
}

// Int returns an integer OAS data type (Schema).
func Int() *Schema { return schema("integer", "") }

// Int32 returns an 32-bit integer OAS data type (Schema).
func Int32() *Schema { return schema("integer", "int32") }

// Int64 returns an 64-bit integer OAS data type (Schema).
func Int64() *Schema { return schema("integer", "int64") }

// Float returns a float OAS data type (Schema).
func Float() *Schema { return schema("number", "float") }

// Double returns a double OAS data type (Schema).
func Double() *Schema { return schema("number", "double") }

// String returns a string OAS data type (Schema).
func String() *Schema { return schema("string", "") }

// UUID returns a UUID OAS data type (Schema).
func UUID() *Schema { return schema("string", "uuid") }

// Bytes returns a base64 encoded OAS data type (Schema).
func Bytes() *Schema { return schema("string", "byte") }

// Binary returns a sequence of octets OAS data type (Schema).
func Binary() *Schema { return schema("string", "binary") }

// Bool returns a boolean OAS data type (Schema).
func Bool() *Schema { return schema("boolean", "") }

// Date returns a date as defined by full-date - RFC3339 OAS data type (Schema).
func Date() *Schema { return schema("string", "date") }

// DateTime returns a date as defined by date-time - RFC3339 OAS data type (Schema).
func DateTime() *Schema { return schema("string", "date-time") }

// Password returns an obscured OAS data type (Schema).
func Password() *Schema { return schema("string", "password") }

// schema returns a Schema for a primitive type.
func schema(t, f string) *Schema {
	return NewSchema().SetType(t).SetFormat(f)
}

// AsArray returns a new "array" Schema wrapping the receiver.
func (s *Schema) AsArray() *Schema {
	return &Schema{
		Type: jsonschema.Array.String(),
		Items: &Items{
			Item: s,
		},
	}
}

// AsEnum returns a new "enum" Schema wrapping the receiver.
func (s *Schema) AsEnum(def json.RawMessage, values ...json.RawMessage) *Schema {
	return &Schema{
		Type:    s.Type,
		Default: Default(def),
		Enum:    append([]json.RawMessage{}, values...),
	}
}

// ToProperty returns a Property with the given name and with this Schema.
func (s *Schema) ToProperty(n string) *Property {
	return NewProperty().SetName(n).SetSchema(s)
}

// NamedSchema can be used to construct a reference to the wrapped Schema.
type NamedSchema struct {
	Schema *Schema
	Name   string
}

// NewNamedSchema returns a new NamedSchema.
func NewNamedSchema(n string, p *Schema) *NamedSchema {
	return &NamedSchema{p, n}
}

// AsLocalRef returns a new Schema referencing the wrapped Schema in the local document.
func (p *NamedSchema) AsLocalRef() *Schema {
	return NewSchema().SetRef("#/components/schemas/" + escapeRef(p.Name))
}

// NewProperty returns a new Property.
func NewProperty() *Property {
	return new(Property)
}

// SetName sets the Name of the Property.
func (p *Property) SetName(n string) *Property {
	p.Name = n
	return p
}

// SetSchema sets the Schema of the Property.
func (p *Property) SetSchema(s *Schema) *Property {
	p.Schema = s
	return p
}

func escapeRef(ref string) string {
	return strings.NewReplacer("~", "~0", "/", "~1").Replace(ref)
}
