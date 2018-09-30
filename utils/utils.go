package utils

// StrIndex is the helper function to find the position of a string in a []string
func StrIndex(vs []string, t string) int {
	for i, v := range vs {
		if v == t {
			return i
		}
	}
	return -1
}

// StrSlice2DCopy is a helper function to make a deep copy of a 2-dimensional string slice.
func StrSlice2DCopy(src [][]string) (cpy [][]string) {
	cpy = make([][]string, len(src))
	for i := range src {
		cpy[i] = make([]string, len(src[i]))
		copy(cpy[i], src[i])
	}
	return
}
