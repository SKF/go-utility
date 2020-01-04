package tags

import (
	"fmt"
	"strings"
)

type Tags struct {
	data []string
}

// AddTagsAsString parses a comma separated string and adds to the list of tags
func (t *Tags) AddTagsAsString(input string) {
	tagList := strings.Split(strings.TrimSpace(input), ",")
	t.data = append(t.data, tagList...)
}

// AddTag adds a key, value pair to the list of tags
func (t *Tags) AddTag(key, value string) {
	if key != "" && value != "" {
		t.data = append(t.data, fmt.Sprintf("%s:%s", key, value))
	}
}

// String joins all tags to a comma separated string
func (t *Tags) String() string {
	return strings.Join(t.data, ",")
}
