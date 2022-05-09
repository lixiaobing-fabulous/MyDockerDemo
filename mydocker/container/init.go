package container

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func RunContainerInitProcess() error {
	log.Infof("command %s, args %s", command, args)

	// private 方式挂载，不影响宿主机的挂载
	err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	if err != nil {
		log.Errorf("private 方式挂载 failed: %v", err)
		return err
	}

	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	err = syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	if err != nil {
		return err
	}

	cmdArray := readUserCommand()

	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		log.Errorf("can't find exec path: %s %v", cmdArray[0], err)
		return err
	}
	log.Infof("find path: %s", path)
	if err := syscall.Exec(path, cmdArray, os.Environ()); err != nil {
		log.Errorf("syscall exec err: %v", err.Error())
	}
	return nil
}

func readUserCommand() []string {
	readPipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(readPipe)
	if err != nil {
		log.Errorf("read init argv pipe err: %v", err)
		return nil
	}
	return strings.Split(string(msg), " ")

}
