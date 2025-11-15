package validators

type Integer interface {
	int | int8 | int16 | int32 | int64
}

func ValidateGreaterOrEqualTo[T Integer](target T, nums ...T) bool {
	for _, n := range nums {
		if n < target {
			return false
		}
	}
	return true
}
