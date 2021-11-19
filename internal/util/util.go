package util

// HasString check to see if a slice of strings contains a specified string.
func HasString(set []string, elem string) bool {
	for _, value := range set {
		if elem == value {
			return true
		}
	}
	return false
}
