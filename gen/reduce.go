package gen

import (
	"bytes"
	"maps"
	"reflect"
	"slices"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/ogen-go/ogen/gen/ir"
	ogenjson "github.com/ogen-go/ogen/json"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/location"
	"github.com/ogen-go/ogen/openapi"
)

// reduceDefault implements convenient errors, representing common default
// response as error instead of variant of each response.
func (g *Generator) reduceDefault(ops []*openapi.Operation) error {
	log := g.log.Named("convenient")
	if g.opt.ConvenientErrors.IsDisabled() {
		log.Info("Convenient errors are disabled, skip reduce")
		return nil
	}
	reduceFailed := func(msg string, p position) error {
		if g.opt.ConvenientErrors.IsForced() {
			err := errors.Wrap(errors.New(msg), "can't reduce to convenient error")

			pos, ok := p.Position()
			if !ok {
				return err
			}

			return &location.Error{
				File: p.File(),
				Pos:  pos,
				Err:  err,
			}
		}
		log.Info("Convenient errors are not available",
			zap.String("reason", msg),
			zapPosition(p),
		)
		return nil
	}

	if len(ops) < 1 {
		return nil
	}

	// Compare first default response to others.
	//
	// TODO(tdakkota): reduce by 4XX/5XX?
	first := ops[0]
	d := first.Responses.Default
	if d == nil {
		return reduceFailed(`operation has no "default" response`, first.Responses)
	}
	switch {
	case len(d.Content) < 1:
		// TODO(tdakkota): point to "content", not to the entire response
		return reduceFailed(`response is no-content`, d)
	case len(d.Content) > 1:
		// TODO(tdakkota): point to "content", not to the entire response
		return reduceFailed(`response is multi-content`, d)
	}
	{
		var ct ir.Encoding
		for key := range d.Content {
			ct = ir.Encoding(key)
			break
		}
		if override, ok := g.opt.ContentTypeAliases[string(ct)]; ok {
			ct = override
		}
		if !ct.JSON() {
			return reduceFailed(`response content must be JSON`, d)
		}
	}

	var c responseComparator
	for _, op := range ops[1:] {
		switch other := op.Responses.Default; {
		case other == nil:
			return reduceFailed(`operation has no "default" response`, op.Responses)
		case !c.compare(d, other):
			return reduceFailed(`response is different`, other)
		}
	}

	ctx := &genctx{
		global: g.tstorage,
		local:  g.tstorage,
	}

	log.Info("Generating convenient error response", zapPosition(d))
	resp, err := g.responseToIR(ctx, "ErrResp", "reduced default response", d, true)
	if err != nil {
		return errors.Wrap(err, "default")
	}

	hasJSON := false
	for _, media := range resp.Contents {
		if media.Encoding.JSON() {
			hasJSON = true
			break
		}
	}
	if resp.NoContent != nil || len(resp.Contents) > 1 || !hasJSON {
		return errors.Wrap(err, "too complicated to reduce default error")
	}

	g.errType = resp
	return nil
}

type responseComparator struct{}

func (c responseComparator) compare(a, b *openapi.Response) bool {
	// Compile time check to not forget to update compareResponses.
	type check struct {
		Ref         openapi.Ref
		Description string
		Headers     map[string]*openapi.Header
		Content     map[string]*openapi.MediaType

		location.Pointer `json:"-" yaml:"-"`
	}
	var (
		_ = (*check)(a)
		_ = (*check)(b)
	)

	switch {
	case a == b:
		return true
	case a == nil || b == nil:
		return false
	case !a.Ref.IsZero() && a.Ref == b.Ref:
		return true
	}

	return maps.EqualFunc(a.Headers, b.Headers, c.compareHeader) &&
		maps.EqualFunc(a.Content, b.Content, c.compareMediaType)
}

func (c responseComparator) compareHeader(a, b *openapi.Header) bool {
	switch {
	case a == b:
		return true
	case a == nil || b == nil:
		return false
	case !a.Ref.IsZero() && a.Ref == b.Ref:
		return true
	}

	return a.Name == b.Name &&
		c.compareSchema(a.Schema, b.Schema) &&
		c.compareParameterContent(a.Content, b.Content) &&
		a.Content == b.Content &&
		a.In == b.In &&
		a.Style == b.Style &&
		a.Explode == b.Explode &&
		a.Required == b.Required &&
		a.AllowReserved == b.AllowReserved
}

func (c responseComparator) compareParameterContent(a, b *openapi.ParameterContent) bool {
	switch {
	case a == b:
		return true
	case a == nil || b == nil:
		return false
	}

	return a.Name == b.Name && c.compareMediaType(a.Media, b.Media)
}

func (c responseComparator) compareMediaType(a, b *openapi.MediaType) bool {
	switch {
	case a == b:
		return true
	case a == nil || b == nil:
		return false
	}

	return c.compareSchema(a.Schema, b.Schema) &&
		maps.EqualFunc(a.Encoding, b.Encoding, c.compareEncoding) &&
		a.XOgenJSONStreaming == b.XOgenJSONStreaming
}

func (c responseComparator) compareEncoding(a, b *openapi.Encoding) bool {
	switch {
	case a == b:
		return true
	case a == nil || b == nil:
		return false
	}

	return a.ContentType == b.ContentType &&
		maps.EqualFunc(a.Headers, b.Headers, c.compareHeader) &&
		a.Style == b.Style &&
		a.Explode == b.Explode &&
		a.AllowReserved == b.AllowReserved
}

func (c responseComparator) compareSchema(a, b *jsonschema.Schema) bool {
	switch {
	case a == b:
		return true
	case a == nil || b == nil:
		return false
	case !a.Ref.IsZero() && a.Ref == b.Ref:
		return true
	}

	compareRequired := func(a, b []string) bool {
		if len(a) != len(b) {
			return false
		}
		k := make(map[string]struct{}, len(a))
		for _, v := range a {
			k[v] = struct{}{}
		}
		for _, v := range b {
			if _, ok := k[v]; !ok {
				return false
			}
		}
		return true
	}

	return a.XOgenName == b.XOgenName &&
		a.Type == b.Type &&
		a.Format == b.Format &&
		a.ContentEncoding == b.ContentEncoding &&
		a.ContentMediaType == b.ContentMediaType &&
		c.compareSchema(a.Item, b.Item) &&
		slices.EqualFunc(a.Items, b.Items, c.compareSchema) &&
		reflect.DeepEqual(a.AdditionalProperties, b.AdditionalProperties) &&
		slices.EqualFunc(a.PatternProperties, b.PatternProperties, c.comparePatternProperty) &&
		slices.EqualFunc(a.Enum, b.Enum, reflect.DeepEqual) &&
		slices.EqualFunc(a.Properties, b.Properties, c.compareProperty) &&
		compareRequired(a.Required, b.Required) &&
		a.Nullable == b.Nullable &&
		slices.EqualFunc(a.OneOf, b.OneOf, c.compareSchema) &&
		slices.EqualFunc(a.AnyOf, b.AnyOf, c.compareSchema) &&
		slices.EqualFunc(a.AllOf, b.AllOf, c.compareSchema) &&
		c.compareDiscriminator(a.Discriminator, b.Discriminator) &&
		c.compareXML(a.XML, b.XML) &&
		c.compareNum(a.Maximum, b.Maximum) &&
		a.ExclusiveMaximum == b.ExclusiveMaximum &&
		c.compareNum(a.Minimum, b.Minimum) &&
		a.ExclusiveMinimum == b.ExclusiveMinimum &&
		c.compareNum(a.MultipleOf, b.MultipleOf) &&
		reflect.DeepEqual(a.MaxLength, b.MaxLength) &&
		reflect.DeepEqual(a.MinLength, b.MinLength) &&
		a.Pattern == b.Pattern &&
		reflect.DeepEqual(a.MaxItems, b.MaxItems) &&
		reflect.DeepEqual(a.MinItems, b.MinItems) &&
		a.UniqueItems == b.UniqueItems &&
		reflect.DeepEqual(a.MaxProperties, b.MaxProperties) &&
		reflect.DeepEqual(a.MinProperties, b.MinProperties) &&
		reflect.DeepEqual(a.Default, b.Default) && a.DefaultSet == b.DefaultSet &&
		maps.Equal(a.ExtraTags, b.ExtraTags) &&
		a.XOgenTimeFormat == b.XOgenTimeFormat
}

func (c responseComparator) comparePatternProperty(a, b jsonschema.PatternProperty) bool {
	return a.Pattern == b.Pattern && c.compareSchema(a.Schema, b.Schema)
}

func (c responseComparator) compareProperty(a, b jsonschema.Property) bool {
	return a.Name == b.Name && c.compareSchema(a.Schema, b.Schema)
}

func (c responseComparator) compareDiscriminator(a, b *jsonschema.Discriminator) bool {
	switch {
	case a == b:
		return true
	case a == nil || b == nil:
		return false
	}

	return a.PropertyName == b.PropertyName &&
		maps.EqualFunc(a.Mapping, b.Mapping, c.compareSchema)
}

func (c responseComparator) compareXML(a, b *jsonschema.XML) bool {
	switch {
	case a == b:
		return true
	case a == nil || b == nil:
		return false
	}

	return *a == *b
}

func (c responseComparator) compareNum(a, b jsonschema.Num) bool {
	if bytes.Equal(a, b) {
		return true
	}
	r, _ := ogenjson.Equal(a, b)
	return r
}
