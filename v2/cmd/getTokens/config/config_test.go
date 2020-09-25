package config_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/SKF/go-utility/v2/cmd/getTokens/config"
)

func TestRead(t *testing.T) {
	mockfile := bytes.NewReader([]byte(`
sandbox:
  Username: Apa
  SSOURL: https://sso.askd.com
  RefreshToken: ;alskjgas;dgjasd
`))

	cfg, err := config.ReadFile(mockfile, "sandbox")

	require.NoError(t, err)
	require.NotEmpty(t, cfg.Username)
	require.NotEmpty(t, cfg.SSOURL)
	require.NotEmpty(t, cfg.RefreshToken)
}
