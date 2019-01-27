package array

type Arrayer interface {
	Equal(i, j int) bool
	Drop(i int)
	Get(i int) interface{}
	Len() int
	Contains(elem interface{}) bool
}

type StringSlice []string

func (ss StringSlice) Equal(i, j int) bool   { return ss[i] == ss[j] }
func (ss *StringSlice) Drop(i int)           { *ss = append((*ss)[:i], (*ss)[i+1:]...) }
func (ss StringSlice) Get(i int) interface{} { return ss[i] }
func (ss StringSlice) Len() int              { return len(ss) }
func (ss StringSlice) Contains(elem interface{}) bool {
	elemString, ok := elem.(string)
	if !ok {
		return false
	}

	for _, s := range ss {
		if s == elemString {
			return true
		}
	}
	return false
}

// Distinct drops all duplicate elements in arr
func Distinct(arr Arrayer) {
	for i := arr.Len() - 1; i > 0; i-- {
		for j := i - 1; j >= 0; j-- {
			if arr.Equal(j, i) {
				arr.Drop(i)
				break
			}
		}
	}
}

// Intersect drops all elements in arr that is not present in all of arrs
func Intersect(arr Arrayer, arrs ...Arrayer) {
	for _, a := range arrs {
		for i := arr.Len() - 1; i >= 0; i-- {
			if !a.Contains(arr.Get(i)) {
				arr.Drop(i)
			}
		}
	}
}

// Difference drops all elements in arr that is also present in one of arrs
func Difference(arr Arrayer, arrs ...Arrayer) {
	for _, a := range arrs {
		for i := arr.Len() - 1; i >= 0; i-- {
			if a.Contains(arr.Get(i)) {
				arr.Drop(i)
			}
		}
	}
}
