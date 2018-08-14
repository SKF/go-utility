// Package env is a wrapper to os.Getenv

package env

import (
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
)

// MustGetAsString returns value for given environment variable,
// panics if not found
func MustGetAsString(variableName string) string {
	value := os.Getenv(variableName)
	if value == "" {
		log.
			WithField("variable", variableName).
			Panic("System variable not set")
	}
	return value
}

// GetAsString returns value for given environment variable,
// returns default if not found
func GetAsString(variableName string, defaultValue string) string {
	value := os.Getenv(variableName)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetAsFloat returns float64 value for given environment variable,
// returns default if not found or couldn't be parsed.
func GetAsFloat(variableName string, defaultValue float64) (floatValue float64) {
	stringValue := os.Getenv(variableName)
	if stringValue == "" {
		return defaultValue
	}

	var err error
	if floatValue, err = strconv.ParseFloat(stringValue, 64); err != nil {
		log.
			WithField("variable", variableName).
			WithField("value", stringValue).
			WithField("error", err).
			Error("Failed to parse string to float")
		return defaultValue
	}
	return
}

// GetAsInt returns int value for given environment variable,
// returns default if not found or couldn't be parsed.
func GetAsInt(variableName string, defaultValue int) (intValue int) {
	stringValue := os.Getenv(variableName)
	if stringValue == "" {
		return defaultValue
	}

	var err error
	if intValue, err = strconv.Atoi(stringValue); err != nil {
		log.
			WithField("variable", variableName).
			WithField("value", stringValue).
			WithField("error", err).
			Error("Failed to parse string to int")
		return defaultValue
	}

	return
}
