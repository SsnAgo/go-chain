package types

type Address [20]uint8

func (a Address) String() string {
	return string(a[:])
}

func (a Address) IsZero() bool {
	for _, v := range a {
		if v != 0 {
			return false
		}
	}
	return true
}
