//go:build !windows && !linux

package sysproxy

import "fmt"

func Enable(addr string) error {
	return fmt.Errorf("%w: set the proxy manually to %s",ErrUnsupported,addr)
}

func Disable() error {
	return nil
}