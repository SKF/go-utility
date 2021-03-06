package main

import (
	"context"
	"fmt"

	"github.com/SKF/go-utility/v2/auth"
	"github.com/SKF/go-utility/v2/cmd/gettokens"
)

func main() {
	cfg := auth.Config{Stage: "sandbox"}
	auth.Configure(cfg)

	h := gettokens.New(nil, nil).WithSignIn(auth.SignIn)
	tokens, err := h.SignIn()
	if err != nil {
		panic(err)
	}

	fmt.Printf("tokens: %v\n", tokens)

	ctx := context.Background()
	tokens, err = auth.SignInRefreshToken(ctx, tokens.RefreshToken)
	if err != nil {
		panic(err)
	}

	fmt.Printf("refresh tokens: %+v\n", tokens)
}
