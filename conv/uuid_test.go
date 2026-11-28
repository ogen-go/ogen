package conv

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUUIDOf(t *testing.T) {
	// Any [16]byte UUID type, like the standard library uuid.UUID.
	type myUUID [16]byte

	a := require.New(t)
	u1, u2 := uuid.New(), uuid.New()

	a.Equal(u1.String(), UUIDOfToString(myUUID(u1)))
	got, err := ToUUIDOf[myUUID](u1.String())
	a.NoError(err)
	a.Equal(myUUID(u1), got)
	_, err = ToUUIDOf[myUUID]("not-uuid")
	a.Error(err)

	strs := UUIDOfArrayToString([]myUUID{myUUID(u1), myUUID(u2)})
	a.Equal([]string{u1.String(), u2.String()}, strs)
	arr, err := ToUUIDOfArray[myUUID](strs)
	a.NoError(err)
	a.Equal([]myUUID{myUUID(u1), myUUID(u2)}, arr)
}
