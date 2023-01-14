package api

import (
	"errors"
	"github.com/nbd-wtf/go-nostr"
	"os"
)

const keyFile = "nostr.key"

func createKeyFile(privateKey string) error {
	return os.WriteFile(keyFile, []byte(privateKey), 0644)
}

func GetPrivateKey() (string, error) {
	if _, err := os.Stat(keyFile); errors.Is(err, os.ErrNotExist) {
		privateKey := nostr.GeneratePrivateKey()
		err = createKeyFile(privateKey)
		if err != nil {
			return "", err
		}
		return privateKey, nil
	}
	dat, err := os.ReadFile(keyFile)
	if err != nil {
		return "", err
	}
	privateKey := string(dat)
	return privateKey, nil

}
