package sshclient

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	//golang.org
	"golang.org/x/crypto/ssh"
)

type Client struct {
	Host string
	User string
	Conn *ssh.Client
}

const (
	SSHErrorZero string = "(ssh)Error #0: error parse ssh key."
	SSHErrorOne string = "(ssh)Error #1: error connect ssh."
	SSHErrorTwo string = "(ssh)Error #2: error create new ssh session."
	SSHErrorThree string = "(ssh)Error #3: error command execution."
	SSHErrorFour string = "(ssh)Error #4: server haven't the internet connection."
)

func (c *Client) CheckInternet() error {
	if c.Conn == nil {
		return errors.New(SSHErrorOne)
	}

	pingSession, err := c.Conn.NewSession()

	if err != nil {
		return errors.New(SSHErrorTwo)
	}

	defer pingSession.Close()

	testNet, err := pingSession.CombinedOutput("ping -c 4 8.8.8.8")

	if err != nil || len(testNet) == 0 {
		return errors.New(SSHErrorFour)
	}

	return nil
}

func ValidPort(port string) (string, error){
	port = strings.TrimSpace(port)

	if port == "" {
		return "",fmt.Errorf("Incorrect port: %s",port)
	}

	parse, err := strconv.ParseUint(port, 10, 16)

	if err != nil || parse == 0 {
		return "",fmt.Errorf("Incorrect port: %s",port)
	}

	return port, nil
}

func Connect(ip,port ,user,privateKeyPEM string) (*Client, error) {
	var authMethod ssh.AuthMethod

	sshPort, err := ValidPort(port)

	if err != nil {
		return nil, err
	}

	if strings.Contains(privateKeyPEM, "BEGIN") && strings.Contains(privateKeyPEM,"PRIVATE KEY") {
		keyBates := []byte(privateKeyPEM)
		signer, err := ssh.ParsePrivateKey(keyBates)
		if err != nil {
			return nil,err
		}
		authMethod = ssh.PublicKeys(signer)
	}else {
		authMethod = ssh.Password(privateKeyPEM)
	}


	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			authMethod,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout: 15*time.Second,
	}

	addr := net.JoinHostPort(ip, sshPort)

	//connected
	conn, err := ssh.Dial("tcp",addr,config)

	if err != nil {
		return nil, errors.New(SSHErrorOne)
	}

	return &Client{
		Host: ip,
		User: user,
		Conn: conn,
	},nil
}

func (c *Client) RunCommand(command string) (string, error) {
	if c.Conn == nil {
		return "", errors.New(SSHErrorOne)
	}
	commandSession, err := c.Conn.NewSession()
	if err != nil {
		return "", errors.New(SSHErrorTwo)
	}

	defer commandSession.Close()

	output, err := commandSession.CombinedOutput(command)

	if err != nil {
		return "",errors.New(SSHErrorThree)
	}

	return string(output),nil
}

func (c *Client) Close() {
	if c.Conn != nil {
		c.Conn.Close()
	}
}