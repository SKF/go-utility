package jwk

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"math/big"
	"net/http"

	"github.com/pkg/errors"
)

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

func (ks JWKeySet) GetPublicKey() (_ *rsa.PublicKey, err error) {
	decodedE, err := base64.RawURLEncoding.DecodeString(ks.Exp)
	if err != nil {
		err = errors.Wrap(err, "failed to decode key set `exp`")
		return
	}

	if len(decodedE) < 4 {
		ndata := make([]byte, 4)
		copy(ndata[4-len(decodedE):], decodedE)
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
	resp, err := http.Get(KeySetURL) // nolint: gosec
	if err != nil {
		err = errors.Wrap(err, "failed to fetch key sets")
		return
	}
	defer resp.Body.Close()

	var data struct {
		Keys JWKeySets `json:"Keys"`
	}

	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		err = errors.Wrap(err, "failed to unmarshal key sets")
		return
	}

	keySets = data.Keys
	return
}
