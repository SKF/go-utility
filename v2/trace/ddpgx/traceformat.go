package ddpgx

import (
	"strings"
)

type TagValueFormatter interface {
	format(interface{}) interface{}
}

func NewDefaultFormatter() TagValueFormatter {
	return &StripNewLinesFormatter{}
}

type StripNewLinesFormatter struct{}

func (s *StripNewLinesFormatter) format(input interface{}) interface{} {
	if value, ok := input.(string); ok {
		return s.stripNewlines(value)
	}

	return input
}

func (s *StripNewLinesFormatter) stripNewlines(input string) string {
	out := patternNewlines.ReplaceAllString(input, " ")
	return strings.TrimSpace(out)
}

type NoopFormatter struct{}

func NewNoopFormatter() TagValueFormatter {
	return &NoopFormatter{}
}

func (_ NoopFormatter) format(v interface{}) interface{} {
	return v
}
