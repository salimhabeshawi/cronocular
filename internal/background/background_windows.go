//go:build windows

package background

import (
	"os/exec"
	"syscall"
)

const detachedProcess = 0x00000008

func prepareDetached(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: detachedProcess | syscall.CREATE_NEW_PROCESS_GROUP,
	}
}
