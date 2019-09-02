package str

import "strings"

// ReplaceSpecialCharForLabelKey replace special char to "-" for label key
func ReplaceSpecialCharForLabelKey(str string) string {
	str = strings.ReplaceAll(str, "@", "-")
	str = strings.ReplaceAll(str, "\"", "-")
	str = strings.ReplaceAll(str, "'", "-")
	str = strings.ReplaceAll(str, " ", "")
	str = strings.ReplaceAll(str, "{", "")
	str = strings.ReplaceAll(str, "}", "")
	return str
}

// ReplaceSpecialCharForLabelValue replace special char to "-" for label value
func ReplaceSpecialCharForLabelValue(str string) string {
	str = strings.ReplaceAll(str, "@", "-")
	str = strings.ReplaceAll(str, "/", "-")
	str = strings.ReplaceAll(str, "\"", "-")
	str = strings.ReplaceAll(str, "\\", "-")
	str = strings.ReplaceAll(str, "'", "-")
	str = strings.ReplaceAll(str, " ", "")
	str = strings.ReplaceAll(str, "{", "")
	str = strings.ReplaceAll(str, "}", "")
	return str
}

// ReplaceSpecialCharForLabel replace special char for label
func ReplaceSpecialCharForLabel(ss map[string]string) map[string]string {
	ret := make(map[string]string)
	for key, value := range ss {
		newKey := ReplaceSpecialCharForLabelKey(key)
		newValue := ReplaceSpecialCharForLabelValue(value)
		ret[newKey] = newValue
	}
	return ret
}
