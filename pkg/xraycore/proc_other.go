//go:build !windows

package xraycore

import "os/exec"

func hideWindow(cmd *exec.Cmd) { }