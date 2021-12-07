package ogen

// specBuilder provides a fluent API to edit the wrapped Spec.
type specBuilder struct {
	spec Spec
}

// SpecBuilder returns a new specBuilder.
func SpecBuilder() *specBuilder {
	return &specBuilder{}
}

// From copies the given spec into the builder.
func (b *specBuilder) From(spec Spec) *specBuilder {
	b.spec = spec
	return b
}

// Spec returns a copy of the internal Spec.
func (b *specBuilder) Spec() Spec {
	return b.spec
}

// SetOpenAPI sets the OpenAPI Specification version of the document.
func (b *specBuilder) SetOpenAPI(v string) *specBuilder {
	b.spec.OpenAPI = v
	return b
}

// SetInfo sets the Info of the Spec.
func (b *specBuilder) SetInfo(i Info) *specBuilder {
	b.spec.Info = i
	return b
}

// InfoBuilder returns an infoBuilder to edit the Info block of the Spec.
func (b *specBuilder) InfoBuilder() *infoBuilder {
	return InfoBuilder().From(b.spec.Info)
}

// infoBuilder provides a fluent API to edit the wrapped Info.
type infoBuilder struct {
	info Info
}

// InfoBuilder returns a new infoBuilder.
func InfoBuilder() *infoBuilder {
	return &infoBuilder{}
}

// From copies the given Info into the builder.
func (b *infoBuilder) From(i Info) *infoBuilder {
	b.info = i
	return b
}

// Info returns a copy of the internal Spec.
func (b *infoBuilder) Info() Info {
	return b.info
}

// SetTitle sets the title of the Info.
func (b *infoBuilder) SetTitle(t string) *infoBuilder {
	b.info.Title = t
	return b
}

// SetDescription sets the description of the Info.
func (b *infoBuilder) SetDescription(d string) *infoBuilder {
	b.info.Description = d
	return b
}

// SetTermsOfService sets the terms of service of the Info.
func (b *infoBuilder) SetTermsOfService(t string) *infoBuilder {
	b.info.TermsOfService = t
	return b
}

// SetVersion sets the version of the Info.
func (b *infoBuilder) SetVersion(v string) *infoBuilder {
	b.info.Version = v
	return b
}

// SetContact sets the Contact of the Info.
func (b *infoBuilder) SetContact(c *Contact) *infoBuilder {
	b.info.Contact = c
	return b
}

// SetLicense sets the License of the Info.
func (b *infoBuilder) SetLicense(l *License) *infoBuilder {
	b.info.License = l
	return b
}

// ContactBuilder returns a contactBuilder to edit the Contact block of the Info.
func (b *infoBuilder) ContactBuilder() *contactBuilder {
	return ContactBuilder().From(b.info.Contact)
}

// LicenseBuilder returns a contactBuilder to edit the Contact block of the Info.
func (b *infoBuilder) LicenseBuilder() *licenseBuilder {
	return LicenseBuilder().From(b.info.License)
}

// contactBuilder provides a fluent API to edit the wrapped Contact.
type contactBuilder struct {
	contact *Contact
}

// ContactBuilder returns a new contactBuilder.
func ContactBuilder() *contactBuilder {
	return &contactBuilder{new(Contact)}
}

// From copies the given Info into the builder.
func (b *contactBuilder) From(c *Contact) *contactBuilder {
	if c != nil {
		b.contact = c
	}
	return b
}

// Contact returns a copy of the internal Contact.
func (b *contactBuilder) Contact() *Contact {
	c := *b.contact
	return &c
}

// SetName sets the Name of the Contact.
func (b *contactBuilder) SetName(n string) *contactBuilder {
	b.contact.Name = n
	return b
}

// SetURL sets the URL of the Contact.
func (b *contactBuilder) SetURL(url string) *contactBuilder {
	b.contact.URL = url
	return b
}

// SetEmail sets the Email of the Contact.
func (b *contactBuilder) SetEmail(e string) *contactBuilder {
	b.contact.Email = e
	return b
}

// licenseBuilder provides a fluent API to edit the wrapped License.
type licenseBuilder struct {
	license *License
}

// LicenseBuilder returns a new licenseBuilder.
func LicenseBuilder() *licenseBuilder {
	return &licenseBuilder{new(License)}
}

// From copies the given Info into the builder.
func (b *licenseBuilder) From(l *License) *licenseBuilder {
	if l != nil {
		b.license = l
	}
	return b
}

// License returns a copy of the internal License.
func (b *licenseBuilder) License() *License {
	l := *b.license
	return &l
}

// SetName sets the Name of the License.
func (b *licenseBuilder) SetName(n string) *licenseBuilder {
	b.license.Name = n
	return b
}

// SetURL sets the URL of the License.
func (b *licenseBuilder) SetURL(url string) *licenseBuilder {
	b.license.URL = url
	return b
}
