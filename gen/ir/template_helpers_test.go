package ir

import "testing"

func TestTypeAcceptsJSONString(t *testing.T) {
	t.Run("PrimitiveString", func(t *testing.T) {
		if !Primitive(String, nil).AcceptsJSONString() {
			t.Fatal("string primitive should accept JSON string")
		}
	})

	t.Run("AliasString", func(t *testing.T) {
		alias := &Type{Kind: KindAlias, AliasTo: Primitive(String, nil)}
		if !alias.AcceptsJSONString() {
			t.Fatal("alias to string should accept JSON string")
		}
	})

	t.Run("SumWithString", func(t *testing.T) {
		sum := &Type{
			Kind: KindSum,
			SumOf: []*Type{
				{Kind: KindStruct},
				Primitive(String, nil),
			},
		}
		if !sum.AcceptsJSONString() {
			t.Fatal("sum with string variant should accept JSON string")
		}
	})

	t.Run("SumWithoutString", func(t *testing.T) {
		sum := &Type{
			Kind: KindSum,
			SumOf: []*Type{
				{Kind: KindStruct},
				{Kind: KindMap},
			},
		}
		if sum.AcceptsJSONString() {
			t.Fatal("sum without string variant should not accept JSON string")
		}
	})
}
