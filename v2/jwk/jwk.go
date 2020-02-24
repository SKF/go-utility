package jwk

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"math/big"
	"net/http"

	"github.com/pkg/errors"

	"github.com/SKF/go-utility/v2/stages"
)

var config *Config

type Config struct {
	Stage string
}

func Configure(conf Config) {
	config = &conf
}

// KeySetURL is used to configure which URL to fetch JWKs from.
// Deprecated: Use Configure(Config{Stage: "..."}) instead.
var KeySetURL string
var keySets JWKeySets

type JWKeySets []JWKeySet

func (kss JWKeySets) LookupKeyID(keyID string) (ks JWKeySet, err error) {
	for _, ks := range kss {
		if ks.KeyID == keyID && ks.Use == "sig" {
			return ks, nil
		}
	}

	if err = RefreshKeySets(); err != nil {
		return
	}

	for _, ks := range kss {
		if ks.KeyID == keyID && ks.Use == "sig" {
			return ks, nil
		}
	}

	return JWKeySet{}, errors.New("unable to find public key")
}

type JWKeySet struct {
	Algorithm string `json:"alg"`
	Exp       string `json:"e"`
	KeyID     string `json:"kid"`
	KeyType   string `json:"kty"`
	Mod       string `json:"n"`
	Use       string `json:"use"`
}

const smallestExpLengthInBytes = 4

func (ks JWKeySet) GetPublicKey() (_ *rsa.PublicKey, err error) {
	decodedE, err := base64.RawURLEncoding.DecodeString(ks.Exp)
	if err != nil {
		err = errors.Wrap(err, "failed to decode key set `exp`")
		return
	}

	if len(decodedE) < smallestExpLengthInBytes {
		ndata := make([]byte, smallestExpLengthInBytes)
		copy(ndata[smallestExpLengthInBytes-len(decodedE):], decodedE)
		decodedE = ndata
	}

	pubKey := &rsa.PublicKey{
		N: &big.Int{},
		E: int(binary.BigEndian.Uint32(decodedE)),
	}

	decodedN, err := base64.RawURLEncoding.DecodeString(ks.Mod)
	if err != nil {
		err = errors.Wrap(err, "failed to decode key set `mod`")
		return
	}

	pubKey.N.SetBytes(decodedN)

	return pubKey, nil
}

func GetKeySets() (_ JWKeySets, err error) {
	if len(keySets) == 0 {
		err = RefreshKeySets()
	}

	return keySets, err
}

func RefreshKeySets() (err error) {
	url, err := getKeySetsURL()
	if err != nil {
		err = errors.Wrap(err, "failed to get key sets URL")
		return
	}

	resp, err := http.Get(url) // nolint: gosec
	if err != nil {
		err = errors.Wrap(err, "failed to fetch key sets")
		return
	}
	defer resp.Body.Close()

	var data map[string]JWKeySets

	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		err = errors.Wrap(err, "failed to unmarshal key sets")
		return
	}

	if keys, present := data["Keys"]; present {
		// keys from Cognito
		keySets = keys
	} else if keys, present := data["data"]; present {
		// keys from SSO-API
		keySets = keys
	} else {
		return errors.New("failed to find key sets in response")
	}

	return err
}

func getKeySetsURL() (string, error) {
	if config == nil && KeySetURL == "" {
		return "", errors.New("jwk is not configured")
	}

	if config == nil {
		return KeySetURL, nil
	}

	if !allowedStages[config.Stage] {
		return "", errors.Errorf("stage %s is not allowed", config.Stage)
	}

	if config.Stage == stages.StageProd {
		return "https://sso-api.users.enlight.skf.com/jwks", nil
	}

	return "https://sso-api." + config.Stage + ".users.enlight.skf.com/jwks", nil
}

var allowedStages = map[string]bool{
	stages.StageProd:         true,
	stages.StageStaging:      true,
	stages.StageVerification: true,
	stages.StageTest:         true,
	stages.StageSandbox:      true,
}
