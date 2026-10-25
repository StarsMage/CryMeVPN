package xraycore

import (
	"CryMeVPN/pkg/core"
	"CryMeVPN/pkg/storage"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type XrayManage struct {
	cmd *exec.Cmd
	ctx context.Context
	cancel context.CancelFunc
}

func NewXrayManage() *XrayManage {
	return &XrayManage{}
}

func (x *XrayManage) StartXray() error {
	x.Stop()

	pathJson,err := storage.GetJsonPath()

	if err != nil {
		return err
	}

	binPath, err := core.GetXrayPath()

	if err != nil {
		return err
	}

	parentCTX := context.Background()
	x.ctx, x.cancel = context.WithCancel(parentCTX)

	cmd := exec.CommandContext(x.ctx, binPath,"run","-config",pathJson)
	hideWindow(cmd)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("xray start error: %w",err)
	}

	x.cmd = cmd 

	done := make(chan error,1)

	go func() { done <- cmd.Wait()}()

	select {
	case err := <-done:
		x.cmd = nil
		details := strings.TrimSpace(stderr.String())
		return fmt.Errorf("xray exited immediately: %v, stderr: %s",err,details)
	case <-time.After(1500*time.Millisecond):
		return nil
	}

}

func (x *XrayManage) Stop() error {
	if x.cancel != nil {
		x.cancel()
	}
	if x.cmd != nil && x.cmd.Process != nil {
		_ = x.cmd.Process.Kill()
		x.cmd = nil
	}
	return nil
}