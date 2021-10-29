package util

// CreatePointerFromValue returns a pointer to the passed value v
func CreatePointerFromValue(v interface{}) *interface{} {
	return &v
}
