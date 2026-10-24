//go:build windows

package launcher

import "os/exec"

func configure(cmd *exec.Cmd) {}

// Windows has no SIGTERM for other processes; closing the child's stdin
// (done by the caller) makes it shut down gracefully.
func interrupt(cmd *exec.Cmd) {}
