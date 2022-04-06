package env

import (
	"log"
	"os"
	"strconv"
)

const bitSize64 = 64

// MustGetAsString returns value for given environment variable
func MustGetAsString(variableName string) string {
	value := os.Getenv(variableName)
	if value == "" {
		log.Panicf("System variable %s not set", variableName)
	}

	return value
}

// GetAsString returns value for given environment variable, with default if not found
func GetAsString(variableName string, defaultValue string) string {
	value := os.Getenv(variableName)
	if value == "" {
		return defaultValue
	}

	return value
}

// GetAsFloat returns value for given environment variable, with default if not found
func GetAsFloat(variableName string, defaultValue float64) (floatValue float64) {
	stringValue := os.Getenv(variableName)
	if stringValue == "" {
		return defaultValue
	}

	var err error
	if floatValue, err = strconv.ParseFloat(stringValue, bitSize64); err != nil {
		log.Panicf("Failed to parse string %s to float - %+v", stringValue, err)
	}

	return
}

// GetAsInt returns value for given environment variable, with default if not found
func GetAsInt(variableName string, defaultValue int) (intValue int) {
	stringValue := os.Getenv(variableName)
	if stringValue == "" {
		return defaultValue
	}

	var err error
	if intValue, err = strconv.Atoi(stringValue); err != nil {
		log.Panicf("Failed to parse string %s to int - %+v", stringValue, err)
	}

	return
}

// GetAsBool returns value for given environment variable, with default if not found
func GetAsBool(variableName string, defaultValue bool) (boolValue bool) {
	stringValue := os.Getenv(variableName)
	if stringValue == "" {
		return defaultValue
	}

	var err error
	if boolValue, err = strconv.ParseBool(stringValue); err != nil {
		log.Panicf("Failed to parse string %s to bool - %+v", stringValue, err)
	}

	return
}
