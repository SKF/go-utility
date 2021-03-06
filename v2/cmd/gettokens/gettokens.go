package gettokens

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/SKF/go-utility/v2/auth"
	"github.com/SKF/go-utility/v2/cmd/gettokens/tokenstorage"
)

type Storage interface {
	GetTokens(stage string) (auth.Tokens, error)
	SetTokens(stage string, tokens auth.Tokens) error
}

type Handler struct {
	in       io.Reader
	out      io.Writer
	inReader *bufio.Reader
	storage  Storage

	stage string

	signIn      func(ctx context.Context, username, password string) (auth.Tokens, error)
	signInToken func(ctx context.Context, refreshToken string) (auth.Tokens, error)
}

func New(stage string, in io.Reader, out io.Writer) Handler {
	h := Handler{}

	if in == nil {
		h.in = os.Stdin
	} else {
		h.in = in
	}

	if out == nil {
		h.out = os.Stdout
	} else {
		h.out = out
	}

	h.stage = stage

	h.inReader = bufio.NewReader(h.in)

	return h
}

func (h *Handler) SignIn() (auth.Tokens, error) {
	tokens, err := h.storage.GetTokens(h.stage)
	if err == nil {
		newToken, err := h.signInToken(context.Background(), tokens.RefreshToken)
		if err != nil {
			return newToken, err
		}

		err = h.storage.SetTokens(h.stage, newToken)

		return newToken, err
	}

	username, err := h.readLine("please enter username")
	if err != nil {
		return auth.Tokens{}, fmt.Errorf("failed to get username: %w", err)
	}

	password, err := h.readLine("please enter password")
	if err != nil {
		return auth.Tokens{}, fmt.Errorf("failed to get password: %w", err)
	}

	newtokens, err := h.signIn(context.Background(), username, password)
	if err != nil {
		return auth.Tokens{}, fmt.Errorf("failed to sign in: %w", err)
	}

	err = h.storage.SetTokens(h.stage, newtokens)
	return newtokens, err

}

func (h *Handler) GetTokens() auth.Tokens {
	return auth.Tokens{}
}

func (h *Handler) readLine(prompt string) (string, error) {
	_, err := fmt.Fprintf(h.out, prompt+"\n")
	if err != nil {
		return "", err
	}

	input, err := h.inReader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read line: %w", err)
	}

	if len(input) == 0 {
		return "", fmt.Errorf("no input provided")
	}

	input = input[:len(input)-1]
	return input, nil
}

func (h Handler) WithSignIn(signIn func(ctx context.Context, username string, password string) (auth.Tokens, error)) Handler {
	h.signIn = signIn

	return h
}

func (h Handler) WithSignInToken(signInToken func(ctx context.Context, refreshToken string) (auth.Tokens, error)) Handler {
	h.signInToken = signInToken

	return h
}

func (h Handler) WithStorage(s tokenstorage.Storage) Handler {
	h.storage = s

	return h
}
