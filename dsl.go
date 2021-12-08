package ogen

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

// SetVersion sets the version of the Info.
func (i *Info) SetVersion(v string) *Info {
	i.Version = v
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
