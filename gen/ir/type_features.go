package ir

import "slices"

func (t *Type) HasFeature(feature string) bool {
	return slices.Contains(t.Features, feature)
}

func (t *Type) AddFeature(feature string) {
	if t.HasFeature(feature) {
		return
	}

	t.Features = append(t.Features, feature)
	switch t.Kind {
	case KindAlias:
		t.AliasTo.AddFeature(feature)
	case KindArray:
		t.Item.AddFeature(feature)
	case KindGeneric:
		t.GenericOf.AddFeature(feature)
	case KindMap:
		t.Item.AddFeature(feature)
		for _, f := range t.Fields {
			f.Type.AddFeature(feature)
		}
	case KindPointer:
		t.PointerTo.AddFeature(feature)
	case KindStruct:
		for _, f := range t.Fields {
			f.Type.AddFeature(feature)
		}
	case KindSum:
		for _, s := range t.SumOf {
			s.AddFeature(feature)
		}
	}
}

func (t *Type) CloneFeatures() []string {
	if t == nil {
		return nil
	}
	return slices.Clone(t.Features)
}
