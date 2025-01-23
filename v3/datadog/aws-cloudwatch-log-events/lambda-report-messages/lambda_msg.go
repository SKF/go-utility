package lambdareportmessages

import (
	"regexp"
	"strings"
)

var (
	patternStartMsg  = regexp.MustCompile(`^START RequestId:\W+([\d\w-]+)`)
	patternEndMsg    = regexp.MustCompile(`^END RequestId:\W+([\d\w-]+)\W*$`)
	patternReportMsg = regexp.MustCompile(`^REPORT RequestId:\W+([\d\w-]+)\W+Duration:\W+(\d+\.?\d*)\W+(\w+)\W+Billed Duration:\W+(\d+)\W+(\w+)\W+Memory Size:\W+(\d+)\W+(\w+)\W+Max Memory Used:\W+(\d+)\W+(\w+)\W*$`)
)

// Parse will try to parse a Lamda Report Message
func Parse(msg string) interface{} {
	switch {
	case strings.HasPrefix(msg, "START"):
		return parseStartMsg(msg)
	case strings.HasPrefix(msg, "END"):
		return parseEndMsg(msg)
	case strings.HasPrefix(msg, "REPORT"):
		return parseReportMsg(msg)
	default:
		return nil
	}
}

type LambdaBaseMsg struct {
	Type      string `json:"type"`
	RequestID string `json:"requestId"`
}

type LambdaReportMsg struct {
	LambdaBaseMsg
	Duration           string `json:"duration"`
	DurationUnit       string `json:"durationUnit"`
	BilledDuration     string `json:"billedDuration"`
	BilledDurationUnit string `json:"billedDurationUnit"`
	MemorySize         string `json:"memorySize"`
	MemorySizeUnit     string `json:"memorySizeUnit"`
	MaxMemoryUsed      string `json:"maxMemoryUsed"`
	MaxMemoryUsedUnit  string `json:"maxMemoryUsedUnit"`
}

func parseStartMsg(msg string) *LambdaBaseMsg {
	matches := patternStartMsg.FindAllStringSubmatch(msg, -1)
	if len(matches) == 1 && len(matches[0]) == 2 {
		return &LambdaBaseMsg{
			Type:      "start",
			RequestID: matches[0][1],
		}
	}

	return nil
}

func parseEndMsg(msg string) *LambdaBaseMsg {
	matches := patternEndMsg.FindAllStringSubmatch(msg, -1)
	if len(matches) == 1 && len(matches[0]) == 2 {
		return &LambdaBaseMsg{
			Type:      "end",
			RequestID: matches[0][1],
		}
	}

	return nil
}

func parseReportMsg(msg string) *LambdaReportMsg {
	matches := patternReportMsg.FindAllStringSubmatch(msg, -1)
	if len(matches) == 1 && len(matches[0]) == 10 {
		base := LambdaBaseMsg{Type: "report", RequestID: matches[0][1]}

		return &LambdaReportMsg{
			LambdaBaseMsg:      base,
			Duration:           matches[0][2],
			DurationUnit:       matches[0][3],
			BilledDuration:     matches[0][4],
			BilledDurationUnit: matches[0][5],
			MemorySize:         matches[0][6],
			MemorySizeUnit:     matches[0][7],
			MaxMemoryUsed:      matches[0][8],
			MaxMemoryUsedUnit:  matches[0][9],
		}
	}

	return nil
}
