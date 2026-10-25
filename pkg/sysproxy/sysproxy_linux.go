package sysproxy

import (
	"net"
	"os"
	"strings"
	"fmt"
)

type backend interface {
	Enable(host, port string) error
	Disable() error
}

var current backend

func detectGUI() backend {
	desktops := strings.Split(strings.ToUpper(os.Getenv("XDG_CURRENT_DESKTOP")), ":")

	for _, d := range desktops {
		switch d {
			case "GNOME", "UNITY", "BUDGIE", "PANTHEON":
				return &gnome{}
		}
	}

	return nil
}

func Enable(addr string) error {
	host, port, err := net.SplitHostPort(addr)

	if err != nil {
		return err
	}

	b := detectGUI()

	if b == nil {
		return fmt.Errorf("%w: set the proxy manually to %s", ErrUnsupported, addr)
	}

	if err := b.Enable(host,port); err != nil {
		return err
	}

	current = b
	return nil
}

func Disable() error {
	if current == nil {
		return nil
	}

	err := current.Disable()
	current = nil
	return err
}