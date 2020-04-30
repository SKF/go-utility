package array

func countMap(arrs ...[]string) (result map[string]int) {
	result = make(map[string]int)

	for _, arr := range arrs {
		distinctArr := DistinctString(arr)
		for _, value := range distinctArr {
			if _, ok := result[value]; ok {
				result[value]++
			} else {
				result[value] = 1
			}
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

const oneOccurence = 1

func DifferenceString(arrs ...[]string) (result []string) {
	result = []string{}

	for key, count := range countMap(arrs...) {
		if count == oneOccurence {
			result = append(result, key)
		}
	}

	return
}

func ContainsString(arr []string, stringToCheck string) bool {
	for _, s := range arr {
		if s == stringToCheck {
			return true
		}
	}

	return false
}
