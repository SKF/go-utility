package pgxcompat

import (
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/SKF/go-utility/v2/uuid"
	"github.com/jackc/pgtype"
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

func (dst *UUID) Set(src interface{}) error {
	if src == nil {
		*dst = UUID{Status: pgtype.Null}
		return nil
	}

	if value, ok := src.(interface{ Get() interface{} }); ok {
		value2 := value.Get()
		if value2 != value {
			return dst.Set(value2)
		}
	}

	switch value := src.(type) {
	case uuid.UUID:
		*dst = UUID{UUID: value, Status: pgtype.Present}
	case [16]byte:
		uuidStr, _ := fromBinary(value[:]) // nolint:errcheck // we know input is 16 bytes
		*dst = UUID{UUID: uuid.UUID(uuidStr), Status: pgtype.Present}
	case []byte:
		if value == nil {
			*dst = UUID{Status: pgtype.Null}
		} else if uuidStr, err := fromBinary(value[:]); err != nil {
			return err
		} else {
			*dst = UUID{UUID: uuid.UUID(uuidStr), Status: pgtype.Present}
		}
	case string:
		if len(value) == uuidStrShortLength {
			value = fmt.Sprintf("%s-%s-%s-%s-%s", value[0:8], value[8:12], value[12:16], value[16:20], value[20:32])
		}

		*dst = UUID{UUID: uuid.UUID(value), Status: pgtype.Present}
	case *string:
		if value == nil {
			*dst = UUID{Status: pgtype.Null}
		} else {
			return dst.Set(*value)
		}
	default:
		return fmt.Errorf("cannot convert %v of type %T to UUID", value, value)
	}

	return nil
}

func (dst UUID) Get() interface{} {
	switch dst.Status {
	case pgtype.Present:
		return dst.UUID
	case pgtype.Null:
		return nil
	default:
		return dst.Status
	}
}

func (src *UUID) AssignTo(dst interface{}) (err error) {
	switch src.Status {
	case pgtype.Present:
		switch v := dst.(type) {
		case *uuid.UUID:
			*v = src.UUID
			return nil
		case *[16]byte:
			*v, err = toBinary(src.UUID.String())
			return err
		case *[]byte:
			*v = make([]byte, uuidBinaryLength)
			bin, err := toBinary(src.UUID.String())
			copy(*v, bin[:])

			return err
		case *string:
			*v = src.UUID.String()
			return nil
		default:
			if nextDst, retry := pgtype.GetAssignToDstType(v); retry {
				return src.AssignTo(nextDst)
			}

			return fmt.Errorf("unable to assign to %T", dst)
		}
	case pgtype.Null:
		return pgtype.NullAssignTo(dst)
	}

	return fmt.Errorf("cannot assign %v into %T", src, dst)
}

func (dst *UUID) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = UUID{Status: pgtype.Null}
		return nil
	}

	u := uuid.UUID(string(src))
	if err := u.Validate(); err != nil {
		return err
	}

	*dst = UUID{UUID: u, Status: pgtype.Present}

	return nil
}

func (dst *UUID) DecodeBinary(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*dst = UUID{Status: pgtype.Null}
		return nil
	}

	if uuidStr, err := fromBinary(src); err != nil {
		return err
	} else {
		*dst = UUID{UUID: uuid.UUID(uuidStr), Status: pgtype.Present}
	}

	return nil
}

func (src UUID) EncodeText(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, errUndefined
	}

	return append(buf, src.UUID.String()...), nil
}

func (src UUID) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case pgtype.Null:
		return nil, nil
	case pgtype.Undefined:
		return nil, errUndefined
	}

	tmpBuf, err := toBinary(src.UUID.String())
	if err != nil {
		return nil, err
	}

	return append(buf, tmpBuf[:]...), nil
}

// Scan implements the database/sql Scanner interface.
func (dst *UUID) Scan(src interface{}) error {
	if src == nil {
		*dst = UUID{Status: pgtype.Null}
		return nil
	}

	switch src := src.(type) {
	case string:
		return dst.DecodeText(nil, []byte(src))
	case []byte:
		return dst.DecodeText(nil, src)
	}

	return fmt.Errorf("cannot scan %T", src)
}

// Value implements the database/sql/driver Valuer interface.
func (src UUID) Value() (driver.Value, error) {
	return pgtype.EncodeValueText(src)
}

func (src UUID) MarshalJSON() ([]byte, error) {
	switch src.Status {
	case pgtype.Present:
		return []byte(`"` + src.UUID.String() + `"`), nil
	case pgtype.Null:
		return []byte("null"), nil
	case pgtype.Undefined:
		return nil, errUndefined
	}

	return nil, errBadStatus
}

func (dst *UUID) UnmarshalJSON(b []byte) error {
	*dst = UUID{UUID: uuid.EmptyUUID, Status: pgtype.Undefined}

	if err := json.Unmarshal(b, &dst.UUID); err != nil {
		return err
	}

	dst.Status = pgtype.Null
	if dst.UUID.IsValid() && dst.UUID != uuid.EmptyUUID {
		dst.Status = pgtype.Present
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
