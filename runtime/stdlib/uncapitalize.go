package stdlib

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/dop251/goja"
)

var allCapsMatcher = regexp.MustCompile(`^[:upper:]+$`)

func smartUncapitalize(s string) string {

	if allCapsMatcher.MatchString(s) {
		return strings.ToLower(s)
	}

	s = strings.ToLower(s[0:1]) + s[1:]

	return s
}

type smartCapFieldNameMapper struct{}

func (u smartCapFieldNameMapper) FieldName(_ reflect.Type, f reflect.StructField) string {
	return smartUncapitalize(f.Name)
}

func (u smartCapFieldNameMapper) MethodName(_ reflect.Type, m reflect.Method) string {
	return smartUncapitalize(m.Name)
}

// UncapFieldNameMapper returns a FieldNameMapper that uncapitalises struct field and method names
// making the first letter lower case.
func newSmartCapFieldNameMapper() goja.FieldNameMapper {
	return smartCapFieldNameMapper{}
}
