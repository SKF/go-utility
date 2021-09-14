package pgxcompat

import (
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgtype"

	"github.com/SKF/go-utility/v2/uuid"
)

const uuidBinaryLength = 16
const uuidStrShortLength = 32
const uuidStrFullLength = 36

var errUndefined = errors.New("cannot encode status undefined")
var errBadStatus = errors.New("invalid status")

type UUID struct {
	UUID   uuid.UUID
	Status pgtype.Status
}

type Getter interface {
	Get() interface{}
}

func (u *UUID) Set(value interface{}) error { // nolint:gocyclo
	if getterValue, ok := value.(Getter); ok {
		value2 := getterValue.Get()
		if value2 != getterValue {
			return u.Set(value2)
		}
	}

	switch value := value.(type) {
	case nil:
		*u = UUID{Status: pgtype.Null}
	case uuid.UUID:
		if value == uuid.EmptyUUID {
			*u = UUID{Status: pgtype.Null}
		} else {
			*u = UUID{UUID: value, Status: pgtype.Present}
		}
	case [16]byte:
		uuidStr, _ := fromBinary(value[:]) // nolint:errcheck // we know input is 16 bytes
		*u = UUID{UUID: uuid.UUID(uuidStr), Status: pgtype.Present}
	case []byte:
		if value == nil {
			*u = UUID{Status: pgtype.Null}
		} else if uuidStr, err := fromBinary(value[:]); err != nil {
			return err
		} else {
			*u = UUID{UUID: uuid.UUID(uuidStr), Status: pgtype.Present}
		}
	case string:
		if len(value) == uuidStrShortLength {
			value = fmt.Sprintf("%s-%s-%s-%s-%s", value[0:8], value[8:12], value[12:16], value[16:20], value[20:32])
		}

		*u = UUID{UUID: uuid.UUID(value), Status: pgtype.Present}
	case *string:
		if value == nil {
			*u = UUID{Status: pgtype.Null}
		} else {
			return u.Set(*value)
		}

	default:
		return fmt.Errorf("cannot convert %v of type %T to UUID", value, value)
	}

	return nil
}

func (u UUID) Get() interface{} {
	switch u.Status {
	case pgtype.Present:
		return u.UUID
	case pgtype.Null:
		return nil
	default:
		return u.Status
	}
}

func (u *UUID) AssignTo(target interface{}) (err error) {
	switch u.Status {
	case pgtype.Present:
		switch dst := target.(type) {
		case *uuid.UUID:
			*dst = u.UUID
			return nil
		case *[16]byte:
			*dst, err = toBinary(u.UUID.String())
			return err
		case *[]byte:
			*dst = make([]byte, uuidBinaryLength)
			bin, err := toBinary(u.UUID.String())
			copy(*dst, bin[:])

			return err
		case *string:
			*dst = u.UUID.String()
			return nil
		default:
			if nextDst, retry := pgtype.GetAssignToDstType(dst); retry {
				return u.AssignTo(nextDst)
			}

			return fmt.Errorf("unable to assign to %T", dst)
		}
	case pgtype.Null:
		return pgtype.NullAssignTo(target)
	}

	return fmt.Errorf("cannot assign %v into %T", u, target)
}

func (u *UUID) DecodeText(ci *pgtype.ConnInfo, value []byte) error {
	if value == nil {
		*u = UUID{Status: pgtype.Null}
		return nil
	}

	uuidValue := uuid.UUID(string(value))
	if err := uuidValue.Validate(); err != nil {
		return err
	}

	*u = UUID{UUID: uuidValue, Status: pgtype.Present}

	return nil
}

func (u *UUID) DecodeBinary(ci *pgtype.ConnInfo, value []byte) error {
	if value == nil {
		*u = UUID{Status: pgtype.Null}
		return nil
	}

	uuidValue, err := fromBinary(value)
	if err != nil {
		return err
	}

	*u = UUID{UUID: uuid.UUID(uuidValue), Status: pgtype.Present}

	return nil
}

func (u UUID) EncodeText(ci *pgtype.ConnInfo, outBuf []byte) ([]byte, error) {
	switch u.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, errUndefined
	}

	return append(outBuf, u.UUID.String()...), nil
}

func (u UUID) EncodeBinary(ci *pgtype.ConnInfo, outBuf []byte) ([]byte, error) {
	switch u.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, errUndefined
	}

	tmpBuf, err := toBinary(u.UUID.String())
	if err != nil {
		return nil, err
	}

	return append(outBuf, tmpBuf[:]...), nil
}

// Scan implements the database/sql Scanner interface.
func (u *UUID) Scan(value interface{}) error {
	if value == nil {
		*u = UUID{Status: pgtype.Null}
		return nil
	}

	switch value := value.(type) {
	case string:
		return u.DecodeText(nil, []byte(value))
	case []byte:
		return u.DecodeText(nil, value)
	}

	return fmt.Errorf("cannot scan %T", value)
}

// Value implements the database/sql/driver Valuer interface.
func (u UUID) Value() (driver.Value, error) {
	return pgtype.EncodeValueText(u)
}

func (u UUID) MarshalJSON() ([]byte, error) {
	switch u.Status {
	case pgtype.Present:
		return []byte(`"` + u.UUID.String() + `"`), nil
	case pgtype.Null:
		return []byte("null"), nil
	case pgtype.Undefined:
		return nil, errUndefined
	}

	return nil, errBadStatus
}

func (u *UUID) UnmarshalJSON(b []byte) error {
	*u = UUID{UUID: uuid.EmptyUUID, Status: pgtype.Undefined}

	if err := json.Unmarshal(b, &u.UUID); err != nil {
		return err
	}

	u.Status = pgtype.Null
	if u.UUID.IsValid() && u.UUID != uuid.EmptyUUID {
		u.Status = pgtype.Present
	}

	return nil
}

func fromBinary(src []byte) (string, error) {
	if len(src) != uuidBinaryLength {
		return "", fmt.Errorf("invalid length for UUID: %v", len(src))
	}

	return fmt.Sprintf("%x-%x-%x-%x-%x", src[0:4], src[4:6], src[6:8], src[8:10], src[10:16]), nil
}

// parseUUID converts a string UUID in standard form to a byte array.
func toBinary(src string) (dst [16]byte, err error) {
	switch len(src) {
	case uuidStrFullLength:
		src = src[0:8] + src[9:13] + src[14:18] + src[19:23] + src[24:]
	case uuidStrShortLength:
		// dashes already stripped, assume valid
	default:
		// assume invalid.
		return dst, fmt.Errorf("cannot parse UUID %v", src)
	}

	buf, err := hex.DecodeString(src)
	if err != nil {
		return dst, err
	}

	copy(dst[:], buf)

	return dst, err
}
