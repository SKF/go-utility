package fakefile_test

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/SKF/go-utility/v2/cmd/gettokens/tokenstorage/fakefile"
)

func TestNew(t *testing.T) {
	fakefile.New()
}

func TestRead_Read(t *testing.T) {
	s := "mystring"

	f := fakefile.New([]byte(s)...)

	buf := make([]byte, len(s))
	n, err := f.Read(buf)
	require.NoError(t, err)
	require.Equal(t, len(s), n)
	require.Equal(t, []byte(s), buf)
}

func TestRead_ReadAll(t *testing.T) {
	s := "mystring"

	f := fakefile.New([]byte(s)...)

	all, err := io.ReadAll(f)
	require.NoError(t, err)
	require.Equal(t, []byte(s), all)
}

func TestRead_WriteRead(t *testing.T) {
	f := fakefile.New()

	s := []byte("hej")

	n, err := f.Write(s)
	require.NoError(t, err)
	require.Equal(t, len(s), n)

	b, err := io.ReadAll(f)
	require.NoError(t, err)
	require.Equal(t, []byte{}, b)
}

func TestRead_WriteSeekRead(t *testing.T) {
	f := fakefile.New()
	s := []byte("hej")

	n, err := f.Write(s)
	require.NoError(t, err)
	require.Equal(t, len(s), n)

	pos, err := f.Seek(1, 0)
	require.NoError(t, err)
	require.Equal(t, int64(1), pos)

	b, err := io.ReadAll(f)
	require.NoError(t, err)
	require.Equal(t, s[1:], b)
}
