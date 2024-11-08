package utils

// ToDo @ender dedicated test for this function
func RemoveDuplicatesFromList[T comparable](items []T) []T {
	uniqueItems := make(map[T]bool)
	var result []T

	for _, item := range items {
		if _, exists := uniqueItems[item]; !exists {
			uniqueItems[item] = true
			result = append(result, item)
		}
	}

	return result
}
