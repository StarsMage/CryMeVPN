package xraygen

import (
	"bytes"
	"text/template"

	//haven't name
	_ "embed"
)

//go:embed client.json.template
var rawTemplateClient string

type ClientStruct struct {
    Address     string
    Port        uint16
    UUID        string
    Flow        string
    SNI         string
    Fingerprint string
    PublicKey   string
    ShortID     string
    SocksPort   uint16
    HTTPPort    uint16
}

const (
    LocalSocksPort uint16 = 10808
    LocalHTTPPort  uint16 = 10809
    FlowVision     string = "xtls-rprx-vision"
)

const (
	FingerChrome  string = "chrome"
	FingerFirefox string = "firefox"
	FingerSafari  string = "safari"
	FingerEdge    string = "edge"
)

func GenerateClientConfigOnBaseTemplate(p ClientStruct) (string, error) {
	tmpl, err := template.New("xray").Parse(rawTemplateClient)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	if attemptExecute := tmpl.Execute(&buf,p); attemptExecute != nil {
		return "",attemptExecute
	}

	return buf.String(),nil
}
