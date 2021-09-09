package pgxcompat_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgtype/testutil"

	"github.com/SKF/go-utility/v2/uuid"
	"github.com/SKF/go-utility/v2/uuid/pgxcompat"
)

func TestUUIDTranscode(t *testing.T) {
	testutil.TestSuccessfulTranscode(t, "uuid", []interface{}{
		&pgxcompat.UUID{UUID: uuid.UUID("5b9cb067-8180-4953-846c-be5764532dc0"), Status: pgtype.Present},
		&pgxcompat.UUID{Status: pgtype.Null},
	})
}

type SomeUUIDWrapper struct {
	SomeUUIDType
}

type SomeUUIDType [16]byte

func stringPtr(str string) *string {
	return &str
}

func TestUUIDSet(t *testing.T) {
	successfulTests := []struct {
		name   string
		source interface{}
		result pgxcompat.UUID
	}{
		{
			name:   "nil",
			source: nil,
			result: pgxcompat.UUID{Status: pgtype.Null},
		},
		{
			name:   "byte array",
			source: [16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			result: pgxcompat.UUID{UUID: uuid.UUID("00010203-0405-0607-0809-0a0b0c0d0e0f"), Status: pgtype.Present},
		},
		{
			name:   "byte slice",
			source: []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			result: pgxcompat.UUID{UUID: uuid.UUID("00010203-0405-0607-0809-0a0b0c0d0e0f"), Status: pgtype.Present},
		},
		{
			name:   "nil byte slice",
			source: ([]byte)(nil),
			result: pgxcompat.UUID{Status: pgtype.Null},
		},
		{
			name:   "string",
			source: "00010203-0405-0607-0809-0a0b0c0d0e0f",
			result: pgxcompat.UUID{UUID: uuid.UUID("00010203-0405-0607-0809-0a0b0c0d0e0f"), Status: pgtype.Present},
		},
		{
			name:   "string without dashes",
			source: "000102030405060708090a0b0c0d0e0f",
			result: pgxcompat.UUID{UUID: uuid.UUID("00010203-0405-0607-0809-0a0b0c0d0e0f"), Status: pgtype.Present},
		},
		{
			name:   "string pointer",
			source: stringPtr("00010203-0405-0607-0809-0a0b0c0d0e0f"),
			result: pgxcompat.UUID{UUID: uuid.UUID("00010203-0405-0607-0809-0a0b0c0d0e0f"), Status: pgtype.Present},
		},
		{
			name:   "nil string pointer",
			source: nil,
			result: pgxcompat.UUID{Status: pgtype.Null},
		},
	}

	for _, tt := range successfulTests {
		var r pgxcompat.UUID

		err := r.Set(tt.source)
		if err != nil {
			t.Errorf("%s: %v", tt.name, err)
		}

		if r != tt.result {
			t.Errorf("%s: expected %v to convert to %v, but it was %v", tt.name, tt.source, tt.result, r)
		}
	}
}

func TestUUIDAssignToByteArray(t *testing.T) { // nolint:gocyclo
	var (
		src      = pgxcompat.UUID{UUID: uuid.UUID("00010203-0405-0607-0809-0a0b0c0d0e0f"), Status: pgtype.Present}
		dst      [16]byte
		expected = [16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	)

	err := src.AssignTo(&dst)
	if err != nil {
		t.Error(err)
	}

	if dst != expected {
		t.Errorf("expected %v to assign %v, but result was %v", src, expected, dst)
	}
}

func TestUUIDAssignToByteSlice(t *testing.T) { // nolint:gocyclo
	var (
		src      = pgxcompat.UUID{UUID: uuid.UUID("00010203-0405-0607-0809-0a0b0c0d0e0f"), Status: pgtype.Present}
		dst      []byte
		expected = []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	)

	err := src.AssignTo(&dst)
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(dst, expected) {
		t.Errorf("expected %v to assign %v, but result was %v", src, expected, dst)
	}
}

func TestUUIDAssignToBinaryUUID(t *testing.T) { // nolint:gocyclo
	var (
		src      = pgxcompat.UUID{UUID: uuid.UUID("00010203-0405-0607-0809-0a0b0c0d0e0f"), Status: pgtype.Present}
		dst      SomeUUIDType
		expected = [16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	)

	err := src.AssignTo(&dst)
	if err != nil {
		t.Error(err)
	}

	if dst != expected {
		t.Errorf("expected %v to assign %v, but result was %v", src, expected, dst)
	}
}

func TestUUIDAssignToString(t *testing.T) { // nolint:gocyclo
	var (
		src      = pgxcompat.UUID{UUID: uuid.UUID("00010203-0405-0607-0809-0a0b0c0d0e0f"), Status: pgtype.Present}
		dst      string
		expected = "00010203-0405-0607-0809-0a0b0c0d0e0f"
	)

	err := src.AssignTo(&dst)
	if err != nil {
		t.Error(err)
	}

	if dst != expected {
		t.Errorf("expected %v to assign %v, but result was %v", src, expected, dst)
	}
}

func TestUUIDAssignToWrappedUUID(t *testing.T) { // nolint:gocyclo
	var (
		src      = pgxcompat.UUID{UUID: uuid.UUID("00010203-0405-0607-0809-0a0b0c0d0e0f"), Status: pgtype.Present}
		dst      SomeUUIDWrapper
		expected = [16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	)

	err := src.AssignTo(&dst)
	if err != nil {
		t.Error(err)
	}

	if dst.SomeUUIDType != expected {
		t.Errorf("expected %v to assign %v, but result was %v", src, expected, dst)
	}
}

func TestUUID_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		src     pgxcompat.UUID
		want    []byte
		wantErr bool
	}{
		{
			name: "Marshal valid UUID to JSON",
			src: pgxcompat.UUID{
				UUID:   uuid.UUID("1d485a7a-6d18-4599-8c6c-34425616887a"),
				Status: pgtype.Present,
			},
			want:    []byte(`"1d485a7a-6d18-4599-8c6c-34425616887a"`),
			wantErr: false,
		},
		{
			name: "Marshal undefined UUID to json",
			src: pgxcompat.UUID{
				UUID:   uuid.EmptyUUID,
				Status: pgtype.Undefined,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Marshal null UUID to JSON",
			src: pgxcompat.UUID{
				UUID:   uuid.EmptyUUID,
				Status: pgtype.Null,
			},
			want:    []byte("null"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.src.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalJSON() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUUID_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		want    *pgxcompat.UUID
		src     []byte
		wantErr bool
	}{
		{
			name: "Unmarshal JSON UUID",
			want: &pgxcompat.UUID{
				UUID:   uuid.UUID("1d485a7a-6d18-4599-8c6c-34425616887a"),
				Status: pgtype.Present,
			},
			src:     []byte(`"1d485a7a-6d18-4599-8c6c-34425616887a"`),
			wantErr: false,
		},
		{
			name: "Unmarshal JSON null",
			want: &pgxcompat.UUID{
				UUID:   uuid.EmptyUUID,
				Status: pgtype.Null,
			},
			src:     []byte("null"),
			wantErr: false,
		},
		{
			name: "Unmarshal invalid JSON UUID",
			want: &pgxcompat.UUID{
				UUID:   uuid.EmptyUUID,
				Status: pgtype.Undefined,
			},
			src:     []byte("1d485a7a-6d18-4599-8c6c-34425616887a"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := &pgxcompat.UUID{}
			if err := got.UnmarshalJSON(tt.src); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() run = %v, error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UnmarshalJSON() run = %v, got = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
