package auth

import (
	"bytes"
	"reflect"
	"strings"
)

// Contains returns true if 'a' contains 'b' and true if the operation was OK.
// If 'a' is a string, a string comparison is performed. Otherwise 'b' may be a
// string or a slice/array of equal or less than the length of 'a'.
func Contains(a, b interface{}) (bool, bool) {
	aValue := reflect.ValueOf(a)
	aType := reflect.TypeOf(a)
	if aType == nil {
		return false, false
	}
	aKind := aType.Kind()

	if aKind == reflect.String {
		bType := reflect.TypeOf(b)
		if bType.Kind() != reflect.String {
			return false, false
		}
		bValue := reflect.ValueOf(b)
		return strings.Contains(aValue.String(), bValue.String()), true
	}

	if aKind == reflect.Slice || aKind == reflect.Array {
		aLength := aValue.Len()
		bType := reflect.TypeOf(b)
		if bType.Kind() == reflect.String {
			for i := 0; i < aLength; i++ {
				if AreEqual(aValue.Index(i).Interface(), b) {
					return true, true
				}
			}
		}
		if bType.Kind() == reflect.Slice ||
			bType.Kind() == reflect.Array {
			bValue := reflect.ValueOf(b)
			bLength := bValue.Len()
			if aLength < bLength {
				return false, true
			}
			for i := 0; i < aLength; i++ {
				aI := aValue.Index(i).Interface()
				for j := 0; j < bLength; j++ {
					bI := bValue.Index(j).Interface()
					if AreEqual(aI, bI) {
						return true, true
					}

				}
			}
		}
	}

	return false, true
}

// AreEqual returns true if 'a' is equal to 'b'. This function performs a byte
// comparison; E.g. normalize strings to lower or upper case.
func AreEqual(a, b interface{}) bool {
	ab, ok := a.([]byte)
	if !ok {
		return reflect.DeepEqual(a, b)
	}
	bb, ok := b.([]byte)
	if !ok {
		return false
	}
	return bytes.Equal(ab, bb)
}
