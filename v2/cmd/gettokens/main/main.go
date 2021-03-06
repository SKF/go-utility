package main

import (
	"fmt"
	"os"

	"github.com/SKF/go-utility/v2/auth"
	"github.com/SKF/go-utility/v2/cmd/gettokens"
	"github.com/SKF/go-utility/v2/cmd/gettokens/tokenstorage"
)

func main() {
	stage := "sandbox"
	cfg := auth.Config{Stage: stage}
	auth.Configure(cfg)

	f, err := os.OpenFile("tokens.yaml", os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	storage := tokenstorage.New(f)

	h := gettokens.New(stage, nil, nil).
		WithSignIn(auth.SignIn).
		WithSignInToken(auth.SignInRefreshToken).
		WithStorage(storage)

	tokens, err := h.SignIn()
	if err != nil {
		panic(err)
	}

	fmt.Printf("tokens: %+v\n", tokens)

}
