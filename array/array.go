package array

func countMap(arrs ...[]string) (result map[string]int) {
	result = make(map[string]int)
	for _, arr := range arrs {
		distinctArr := DistinctString(arr)
		for _, value := range distinctArr {
			result[value]++
		}
	}

	return
}

func DistinctString(arr []string) (result []string) {
	arrayMap := make(map[string]bool)
	for _, value := range arr {
		arrayMap[value] = true
	}

	for value := range arrayMap {
		result = append(result, value)
	}

	return
}

func IntersectString(arrs ...[]string) (result []string) {
	result = []string{}
	noArrs := len(arrs)
	for key, count := range countMap(arrs...) {
		if count == noArrs {
			result = append(result, key)
		}
	}

	return
}

func DifferenceString(arrs ...[]string) (result []string) {
	result = []string{}
	for key, count := range countMap(arrs...) {
		if count == 1 {
			result = append(result, key)
		}
	}

	return
}
