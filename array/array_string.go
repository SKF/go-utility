package array

func DistinctString(arr []string) []string {
	ret := make([]string, len(arr))
	copy(ret, arr)

	ss := StringSlice(ret)

	Distinct(&ss)
	return ss
}

func IntersectString(arr []string, arrs ...[]string) []string {
	ret := make([]string, len(arr))
	copy(ret, arr)

	ss := StringSlice(ret)
	Distinct(&ss)
	sarrs := []Arrayer{}

	for _, a := range arrs {
		sa := StringSlice(a)
		sarrs = append(sarrs, &sa)
	}
	Intersect(&ss, sarrs...)
	return ss
}

func DifferenceString(arr []string, arrs ...[]string) []string {
	ret := make([]string, len(arr))
	copy(ret, arr)

	ss := StringSlice(ret)
	Distinct(&ss)
	sarrs := []Arrayer{}

	for _, a := range arrs {
		sa := StringSlice(a)
		sarrs = append(sarrs, &sa)
	}
	Difference(&ss, sarrs...)
	return ss
}
