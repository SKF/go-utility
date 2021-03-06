package tokenstorage

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v3"

	"github.com/SKF/go-utility/v2/auth"
)

type Storage struct {
	store io.ReadWriteSeeker
}

var ErrNotFound = fmt.Errorf("tokens not found")

func (s Storage) GetTokens(stage string) (auth.Tokens, error) {
	out, err := s.getAllTokens()

	return out[stage], err
}

func (s Storage) getAllTokens() (map[string]auth.Tokens, error) {
	_, err := s.store.Seek(0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to rewind store: %w", err)
	}

	bytes, err := io.ReadAll(s.store)
	if err != nil {
		return nil, err
	}

	out := map[string]auth.Tokens{}
	err = yaml.Unmarshal(bytes, &out)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (s Storage) SetTokens(stage string, tokens auth.Tokens) error {
	out, err := s.getAllTokens()
	if err != nil {
		return fmt.Errorf("failed to read current config: %w", err)
	}

	out[stage] = tokens

	_, err = s.store.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("failed to store tokens: %w", err)
	}

	bytes, err := yaml.Marshal(out)
	if err != nil {
		return fmt.Errorf("failed to marshal bytes: %w", err)
	}

	_, err = s.store.Write(bytes)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

type Store interface {
	GetTokens(stage string) (auth.Tokens, error)
}

func New(store io.ReadWriteSeeker) Storage {
	return Storage{store: store}
}
