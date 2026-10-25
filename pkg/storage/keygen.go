package storage

import (
	"crypto/rand"
	"encoding/hex"
	"errors"

	"github.com/zalando/go-keyring"
)

const (
	service = "CryMeVpn"
	keyName = "secret-master-key"
)

func GetOnCreateKey() ([]byte, error) {
	secret, err := keyring.Get(service,keyName)

	if err == nil && len(secret) > 0 {
		return hex.DecodeString(secret)
	}

	if err != nil && !errors.Is(err, keyring.ErrNotFound) {
	return nil, err
}
	keyBytes := make([]byte,32)

	if _, err := rand.Read(keyBytes); err != nil {
		return nil,err
	}

	hexKey := hex.EncodeToString(keyBytes)

	if err := keyring.Set(service,keyName,hexKey); err != nil {
		return nil, err
	}

	return keyBytes,nil

}