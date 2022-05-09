package container

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func NewParentProcess(tty bool, cmdArray []string) *exec.Cmd {
	commands := strings.Join(cmdArray, " ")
	log.Infof("conmand all is %s", commands)
	cmd := exec.Command("/proc/self/exe", "init", cmdArray[0], commands)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd
}
