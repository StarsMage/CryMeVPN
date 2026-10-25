package deployer

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	//pkg/sshclient
	"CryMeVPN/pkg/sshclient"
)

const (
	DEPLOYErrorZero string = "(deploy) Error #0: error installing Docker on the server."
	DEPLOYErrorOne string = "(deploy) Error #1: error stariting Docker on the server."
	DEPLOYErrorTwo string = "(deploy) Error #2: error exicutions command on the ssh."
	DEPLOYErrorThree string = "(deploy) Error #3: error writing configuration file."
	DEPLOYErrorFour string = "(deploy) Error #4: error restart docker container."
	DEPLOYErrorFive string = "(deploy) Error #5: error create docker container."
)

const (
	CriticalDEPLOYErrorZero string = "(deploy) Critical Error #0: unsupported os."
)

func GenerateNameServer(ip string) string {
	hash := sha256.Sum256([]byte(ip))
	hashStr := hex.EncodeToString(hash[:])
	return hashStr[:7]
}

func CheckDockerExists(client *sshclient.Client) bool {
	check, err := client.RunCommand("if which docker >/dev/null 2>&1; then echo 'dockerSea'; fi") // Check having XRAY on base checking $PATH

	if err != nil {
		return false
	}

	cleanedCheckOut := strings.TrimSpace(check)
	return strings.Contains(cleanedCheckOut,"dockerSea")
}


func DeployAndConfigureXray(client *sshclient.Client, xrayConfig, serverPort, serverIP string) error {
	if err := client.CheckInternet(); err != nil {
		return err
	}
	if resp,_ := client.RunCommand("uname -a"); !(strings.Contains(resp,"Linux")) {
		return errors.New(CriticalDEPLOYErrorZero)
	}

	if !CheckDockerExists(client) {
		_, err := client.RunCommand("curl -fsSL https://get.docker.com | sh")
		if err != nil {
			return errors.New(DEPLOYErrorZero)
		}

		resp,_ := client.RunCommand("systemctl enable --now docker; systemctl is-active docker")
		if strings.TrimSpace(resp) != "active" {
			return errors.New(DEPLOYErrorOne)
		}
	}

	_, err := client.RunCommand("mkdir -p /etc/crymevpn/xray")

	if err != nil {
		return errors.New(DEPLOYErrorTwo)
	}

	encodeConfig := base64.StdEncoding.EncodeToString([]byte(xrayConfig))
	configCommand := fmt.Sprintf("echo '%s' | base64 -d > /etc/crymevpn/xray/config.json",encodeConfig)
	_, err = client.RunCommand(configCommand)
	if err != nil {
		return errors.New(DEPLOYErrorThree)
	}
	
	containerName := GenerateNameServer(serverIP)


	runCmd := fmt.Sprintf(
	"docker rm -f %[1]s >/dev/null 2>&1 || true; "+
		"docker run -d --name %[1]s --restart always "+
		"-v /etc/crymevpn/xray/config.json:/etc/xray/config.json "+
		"--log-opt max-size=1m --log-opt max-file=1 "+ 
		"-p %[2]s:%[2]s teddysun/xray", containerName, serverPort)

	if _, err := client.RunCommand(runCmd); err != nil {
		return errors.New(DEPLOYErrorFive)
	}

	out, err := client.RunCommand(fmt.Sprintf("sleep 2; docker inspect -f '{{.State.Running}}' %s", containerName)) 

	if err != nil || !strings.Contains(out, "true") {
		return errors.New(DEPLOYErrorFive)
	}

	return nil
}