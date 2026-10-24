//go:build !windows

package launcher

import (
	"os/exec"
	"syscall"
)

func configure(cmd *exec.Cmd) {}

func interrupt(cmd *exec.Cmd) {
	_ = cmd.Process.Signal(syscall.SIGTERM)
}
