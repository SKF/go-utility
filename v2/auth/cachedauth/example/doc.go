package main

import (
	"context"
	"fmt"

	auth "github.com/SKF/go-utility/v2/auth/cachedauth"
)

func main() {
	ctx := context.Background()

	conf := auth.Config{
		Stage: "sandbox",
	}
	auth.Configure(conf)

	if err := auth.SignIn(ctx, "<email_address", "<password>"); err != nil {
		panic(err)
	}

	fmt.Println(auth.GetTokens().AccessToken)
}
