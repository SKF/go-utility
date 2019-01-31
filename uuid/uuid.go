// Package uuid is a wrapper to github.com/google/uuid
package uuid

import (
	googleUUID "github.com/google/uuid"
)

const EmptyUUID UUID = UUID("00000000-0000-0000-0000-000000000000")

type UUID string

// New returns a random UUID
func New() UUID {
	return UUID(googleUUID.New().String())
}

func IsValid(value string) bool {
	return UUID(value).IsValid()
}

func (uuid UUID) IsValid() bool {
	return uuid.Validate() == nil
}

func (uuid UUID) Validate() (err error) {
	_, err = googleUUID.Parse(uuid.String())
	return
}

func (uuid UUID) String() string {
	return string(uuid)
}

func (uuid UUID) StringPtr() *string {
	val := uuid.String()
	return &val
}
