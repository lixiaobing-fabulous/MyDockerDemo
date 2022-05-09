package container

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"syscall"
)

func RunContainerInitProcess(command string, args []string) error {
	log.Infof("command %s, args %s", command, args)

	// private 方式挂载，不影响宿主机的挂载
	//err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	//if err != nil {
	//	log.Errorf("private 方式挂载 failed: %v", err)
	//	return err
	//}

	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	err := syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	if err != nil {
		return err
	}
	path, err := exec.LookPath(command)
	if err != nil {
		log.Errorf("can't find exec path: %s %v", command, err)
		return err
	}
	log.Infof("find path: %s", path)
	if err := syscall.Exec(path, args, os.Environ()); err != nil {
		log.Errorf("syscall exec err: %v", err.Error())
	}
	return nil
}
