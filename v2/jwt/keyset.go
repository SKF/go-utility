package jwt

import (
	"sync"

	"github.com/SKF/go-utility/v2/jwk"
)

var keySetM = &sync.Mutex{}

func getKeySets() (jwk.JWKeySets, error) {
	keySetM.Lock()
	defer keySetM.Unlock()

	return jwk.GetKeySets()
}
