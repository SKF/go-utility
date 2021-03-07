package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/SKF/go-utility/v2/auth"
	"github.com/SKF/go-utility/v2/cmd/gettokens"
	"github.com/SKF/go-utility/v2/cmd/gettokens/tokenstorage"
)

func main() {
	printTokens := flag.Bool("print", false, "print tokens to stdout")
	flag.Parse()

	stage := flag.Arg(0)
	if stage == "" {
		fmt.Printf("please specify stage\n")
		os.Exit(1)
	}

	cfg := auth.Config{Stage: stage}
	auth.Configure(cfg)

	configpath, err := getConfigpath()
	if err != nil {
		fmt.Printf("failed to get path: %w\n", err)
		os.Exit(1)
	}

	f, err := getFile(configpath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	storage := tokenstorage.New(f)

	h := gettokens.New(stage, nil, nil).
		WithSignIn(auth.SignIn).
		WithSignInToken(auth.SignInRefreshToken).
		WithStorage(storage)

	tokens, err := h.SignIn()
	if err != nil {
		panic(err)
	}

	if *printTokens {
		fmt.Printf("accessToken: %s\n", tokens.AccessToken)
		fmt.Printf("identityToken: %s\n", tokens.IdentityToken)
		fmt.Printf("refreshToken: %s\n", tokens.RefreshToken)
	}

	err = os.WriteFile(path.Join(configpath, "accesstoken"), toAuthBytes(tokens.AccessToken), 0600)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(path.Join(configpath, "identitytoken"), toAuthBytes(tokens.IdentityToken), 0600)
	if err != nil {
		panic(err)
	}
}

func toAuthBytes(token string) []byte {
	return []byte(fmt.Sprintf("Authorization: %s", token))
}

func getConfigpath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	p := path.Join(homeDir, ".skf")

	return p, nil
}

func getFile(configPath string) (*os.File, error) {
	p := path.Join(configPath, "config.yaml")
	if _, err := os.Stat(p); os.IsNotExist(err) {
		err := os.MkdirAll(configPath, 0700)
		if err != nil {
			return nil, err
		}
	}

	return os.OpenFile(p, os.O_RDWR|os.O_CREATE, 0600)
}
