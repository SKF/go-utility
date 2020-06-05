package cache

import (
	"strings"
)

const separator = ";"

type ObjectKey string

// FuncName returns the function that created the key
func (o ObjectKey) FuncName() string {
	return strings.Split(string(o), separator)[0]
}

// Key creates a cache key from a function name (of who's is reading/writing
// from/to the cache) and a list of unique key elements.
func Key(funcName string, fields ...string) ObjectKey {
	keyFields := append([]string{funcName}, fields...)
	keyStr := strings.Join(keyFields, separator)
	return ObjectKey(keyStr)
}
