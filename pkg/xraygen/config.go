package xraygen

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"text/template"

	//haven't name
	_ "embed"

	//github
	"github.com/google/uuid"

	//golang
	"golang.org/x/crypto/curve25519"
)

//go:embed config.json.template
var rawTemplate string

const (
	XRAYGENErrorZero string = "(xraygen) Error #0: error parse config.json.template ."
	XRAYGENErrorOne string = "(xraygen) Error #1: error execute attempt create config.json."
)

const (
	CloudFlare string = "1.1.1.1"
	Quad9 string = "9.9.9.9"
	AdGuard string = "94.140.14.14"
)

type ConfigStruct struct {
	Port uint16
	Destination string
	ServerName string
	DNS string
	UUID string
	PrivateKey string
	ShortID string
}

type Credetinals struct {
	UUID string
	PrivateKey string
	OpenKey string
	ShortID string
}

func GenerateConfigOnBaseTemplate(p ConfigStruct) (string, error) {
	tmpl, err := template.New("xray").Parse(rawTemplate)
	if err != nil {
		return "", errors.New(XRAYGENErrorZero)
	}

	var buf bytes.Buffer

	if attemptExecute := tmpl.Execute(&buf,p); attemptExecute != nil {
		return "",errors.New(XRAYGENErrorOne)
	}

	return buf.String(),nil
}

func GenerateCredentials() (Credetinals, error) {
	//UUID v4
	newUUID := uuid.New().String()

	// X25519 keys
	privateKeyBytes := make([]byte,32)

	if _, err := rand.Read(privateKeyBytes); err != nil {
		return Credetinals{},err
	}

	privateKeyBytes[0] &= 248
	privateKeyBytes[31] &= 127
	privateKeyBytes[31] |= 64

	var publicKey [32]byte
	var privcKey [32]byte
	copy(privcKey[:], privateKeyBytes)

	curve25519.ScalarBaseMult(&publicKey,&privcKey)

	// ShortID
	shordIDBytes := make([]byte, 4)
	if _, err := rand.Read(shordIDBytes); err != nil {
		return Credetinals{}, err
	}

	return Credetinals{
		UUID: newUUID,
		PrivateKey: base64.RawURLEncoding.EncodeToString(privateKeyBytes),
		OpenKey: base64.RawURLEncoding.EncodeToString(publicKey[:]),
		ShortID: hex.EncodeToString(shordIDBytes),
	}, nil
}