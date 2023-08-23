package jwk

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"

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
var getKeySetsLock = &sync.Mutex{}

type JWKeySets []JWKeySet

func (kss JWKeySets) LookupKeyID(keyID string) (JWKeySet, error) {
	for _, ks := range kss {
		if ks.KeyID == keyID && ks.Use == "sig" {
			return ks, nil
		}
	}

	if err := RefreshKeySets(); err != nil {
		return JWKeySet{}, err
	}

	for _, ks := range kss {
		if ks.KeyID == keyID && ks.Use == "sig" {
			return ks, nil
		}
	}

	return JWKeySet{}, fmt.Errorf("unable to find public key")
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

func (ks JWKeySet) GetPublicKey() (*rsa.PublicKey, error) {
	decodedE, err := base64.RawURLEncoding.DecodeString(ks.Exp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode key set `exp`: %w", err)
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
		return nil, fmt.Errorf("failed to decode key set `mod`: %w", err)
	}

	pubKey.N.SetBytes(decodedN)

	return pubKey, nil
}

func GetKeySets() (JWKeySets, error) {
	getKeySetsLock.Lock()
	defer getKeySetsLock.Unlock()

	if len(keySets) == 0 {
		if err := RefreshKeySets(); err != nil {
			return JWKeySets{}, err
		}
	}

	return keySets, nil
}

func RefreshKeySets() error {
	url, err := getKeySetsURL()
	if err != nil {
		return fmt.Errorf("failed to get key sets URL: %w", err)
	}

	resp, err := http.Get(url) // nolint: gosec
	if err != nil {
		return fmt.Errorf("failed to fetch key sets: %w", err)
	}

	defer resp.Body.Close()

	var data map[string]JWKeySets

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non 200 status code when fetching key sets, %d", resp.StatusCode)
	}

	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return fmt.Errorf("failed to unmarshal key sets: %w", err)
	}

	if keys, present := data["Keys"]; present {
		// keys from Cognito
		keySets = keys
	} else if keys, present := data["keys"]; present {
		// keys from Cognito
		keySets = keys
	} else if keys, present := data["data"]; present {
		// keys from SSO-API
		keySets = keys
	} else {
		return fmt.Errorf("failed to find key sets in response")
	}

	return err
}

func getKeySetsURL() (string, error) {
	if config == nil && KeySetURL == "" {
		return "", fmt.Errorf("jwk is not configured")
	}

	if config == nil {
		return KeySetURL, nil
	}

	if !allowedStages[config.Stage] {
		return "", fmt.Errorf("stage %s is not allowed", config.Stage)
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
