package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/SKF/go-utility/v2/cmd/getTokens/config"

	"github.com/SKF/go-utility/v2/cmd/getTokens/show"

	"github.com/SKF/go-utility/v2/stages"

	"github.com/SKF/go-utility/v2/cmd/getTokens/ssoClient"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Expecting environment as first param\n")
		fmt.Printf("got: %v\n", os.Args)
		return
	}

	environ := os.Args[1]
	if !validEnviron(environ) {
		fmt.Printf("invalid environment: %s", environ)
		return
	}
	fmt.Printf("using environment: %s\n", environ)

	configDir, err := config.GetConfigDir()
	if err != nil {
		panic(err)
	}

	cfg, err := config.Read(environ)
	if err != nil {
		panic(err)
	}

	sso := ssoClient.Client{}

	tokens, err := sso.SignInInitiate(cfg)
	if err != nil {
		panic(err)
	}

	accessTokenPath := path.Join(configDir, "accesstoken")
	writeFile(accessTokenPath, tokens.AccessToken)

	identityTokenPath := path.Join(configDir, "identitytoken")
	writeFile(identityTokenPath, tokens.IdentityToken)

	show.Show(fmt.Sprintf("accesstoken: %s\n\nidentityToken: %s", tokens.AccessToken, tokens.IdentityToken))
}

func writeFile(accessTokenPath, token string) {
	const authorizationHeader = "authorization: "

	err := ioutil.WriteFile(accessTokenPath, []byte(authorizationHeader+token), 0600)
	if err != nil {
		panic(err)
	}

	fmt.Printf("token updated in %s\n", accessTokenPath)
}

func validEnviron(environ string) bool {
	switch environ {
	case stages.StageSandbox, stages.StageTest, stages.StageStaging, stages.StageProd:
		return true
	default:
		return false
	}
}
