package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"

	"github.com/SKF/go-utility/v2/cmd/getTokens/show"

	"github.com/SKF/go-utility/v2/stages"

	"github.com/SKF/go-utility/v2/cmd/getTokens/ssoClient"

	"github.com/SKF/go-utility/v2/cmd/getTokens/model"
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

	usr, err := user.Current()
	if err != nil {
		panic(fmt.Errorf("failed to get current user: %w", err))
	}

	file, err := ioutil.ReadFile(path.Join(usr.HomeDir, fmt.Sprintf(".skf/%s.json", environ)))
	if err != nil {
		err = fmt.Errorf("Failed to get credentials: %w", err)
		panic(err)
	}

	var config model.Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		panic(fmt.Errorf("failed to parse config: %w", err))
	}

	sso := ssoClient.Client{}

	tokens, err := sso.SignInInitiate(config)
	if err != nil {
		panic(err)
	}

	accessTokenPath := path.Join(usr.HomeDir, ".skf/accesstoken")
	writeFile(accessTokenPath, tokens.AccessToken)

	identityTokenPath := path.Join(usr.HomeDir, ".skf/identitytoken")
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
