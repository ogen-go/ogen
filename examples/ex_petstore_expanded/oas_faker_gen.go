// Code generated by ogen, DO NOT EDIT.

package api

// SetFake set fake values.
func (s *Error) SetFake() {
	{
		{
			s.Code = int32(0)
		}
	}
	{
		{
			s.Message = "string"
		}
	}
}

// SetFake set fake values.
func (s *FindPetsOKApplicationJSON) SetFake() {
	var unwrapped []Pet
	{
		unwrapped = nil
		for i := 0; i < 0; i++ {
			var elem Pet
			{
				elem.SetFake()
			}
			unwrapped = append(unwrapped, elem)
		}
	}
	*s = FindPetsOKApplicationJSON(unwrapped)
}

// SetFake set fake values.
func (s *NewPet) SetFake() {
	{
		{
			s.Name = "string"
		}
	}
	{
		{
			s.Tag.SetFake()
		}
	}
}

// SetFake set fake values.
func (s *OptString) SetFake() {
	var elem string
	{
		elem = "string"
	}
	s.SetTo(elem)
}

// SetFake set fake values.
func (s *Pet) SetFake() {
	{
		{
			s.Name = "string"
		}
	}
	{
		{
			s.Tag.SetFake()
		}
	}
	{
		{
			s.ID = int64(0)
		}
	}
}
