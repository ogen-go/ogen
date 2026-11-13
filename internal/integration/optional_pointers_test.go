package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/_testdata/testtypes"
	api "github.com/ogen-go/ogen/internal/integration/test_optional_pointers"
)

func TestOptionalPointersOpt(t *testing.T) {
	a := require.New(t)

	unset := api.NewOptStringFromPtr(nil)
	a.False(unset.IsSet())
	a.Nil(unset.Ptr())
	a.Equal(api.OptString{}, unset)

	v := "foo"
	set := api.NewOptStringFromPtr(&v)
	a.True(set.IsSet())
	a.Equal(api.NewOptString("foo"), set)

	got := set.Ptr()
	a.NotNil(got)
	a.Equal("foo", *got)
	// Ptr returns a copy: changing it does not change the wrapper, nor the source.
	*got = "bar"
	a.Equal("foo", set.Value)
	a.Equal("foo", v)

	n := int64(42)
	a.Equal(api.NewOptInt64(42), api.NewOptInt64FromPtr(&n))
	a.Equal(int64(42), *api.NewOptInt64FromPtr(&n).Ptr())

	name := api.Name("ogen")
	a.Equal(api.NewOptName("ogen"), api.NewOptNameFromPtr(&name))
	a.Nil(api.NewOptNameFromPtr(nil).Ptr())

	obj := api.Object{ID: 1}
	a.Equal(api.NewOptObject(obj), api.NewOptObjectFromPtr(&obj))
	a.Equal(&obj, api.NewOptObjectFromPtr(&obj).Ptr())
}

func TestOptionalPointersNil(t *testing.T) {
	a := require.New(t)

	null := api.NewNilStringFromPtr(nil)
	a.True(null.IsNull())
	a.Nil(null.Ptr())

	v := "foo"
	set := api.NewNilStringFromPtr(&v)
	a.False(set.IsNull())
	a.Equal(api.NewNilString("foo"), set)
	a.Equal("foo", *set.Ptr())
}

func TestOptionalPointersOptNil(t *testing.T) {
	a := require.New(t)

	// A nil pointer is an explicit null, not an omitted value.
	null := api.NewOptNilStringFromPtr(nil)
	a.True(null.IsSet())
	a.True(null.IsNull())
	a.False(null.IsEmpty())
	a.Nil(null.Ptr())

	// An omitted value has no pointer either.
	a.Nil(api.OptNilString{}.Ptr())

	v := "foo"
	set := api.NewOptNilStringFromPtr(&v)
	a.True(set.IsSet())
	a.False(set.IsNull())
	a.Equal(api.NewOptNilString("foo"), set)
	a.Equal("foo", *set.Ptr())

	list := []string{"a", "b"}
	a.Equal(api.NewOptNilStringArray(list), api.NewOptNilStringArrayFromPtr(&list))
	a.Equal(list, *api.NewOptNilStringArrayFromPtr(&list).Ptr())
	a.True(api.NewOptNilStringArrayFromPtr(nil).IsNull())
}

// The wrappers of other Go types get Ptr too; their constructors are covered by the
// cases above, as every wrapper is rendered by the same template.
func TestOptionalPointersOtherTypes(t *testing.T) {
	a := require.New(t)

	var w api.Wrappers
	a.Nil(w.OptTime.Ptr())
	// The zero Nil wrapper holds the zero value, not null.
	a.Equal(&api.NullableObject{}, w.NilObject.Ptr())
	w.NilObject.SetToNull()
	a.Nil(w.NilObject.Ptr())

	now := time.Now()
	w.OptTime.SetTo(now)
	a.Equal(now, *w.OptTime.Ptr())

	w.NilObject.SetTo(api.NullableObject{ID: 2})
	a.Equal(api.NullableObject{ID: 2}, *w.NilObject.Ptr())
}

// A wrapper of an x-ogen-type external type converts from and to a pointer of that type.
func TestOptionalPointersExternal(t *testing.T) {
	a := require.New(t)

	a.Nil(api.NewOptStringOgenFromPtr(nil).Ptr())

	ext := testtypes.StringOgen{Value: "ogen"}
	opt := api.NewOptStringOgenFromPtr(&ext)
	a.True(opt.IsSet())
	a.Equal(ext, *opt.Ptr())
}
