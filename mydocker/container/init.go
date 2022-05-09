package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func RunContainerInitProcess() error {

	if err := setUpMount(); err != nil {
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

func setUpMount() error {
	// 首先设置根目录为私有模式，防止影响pivot_root
	if err := syscall.Mount("/", "/", "", syscall.MS_REC|syscall.MS_PRIVATE, ""); err != nil {
		return fmt.Errorf("setUpMount Mount proc err: %v", err)
	}
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current location err: %v", err)
	}
	log.Infof("current location: %s", pwd)

	err = pivotRoot(pwd)
	if err != nil {
		return err
	}

	// mount proc
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	err = syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	if err != nil {
		log.Errorf("proc挂载 failed: %v", err)
		return err
	}
	syscall.Mount("tmpfs", "/dev", "tempfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
	return nil
}

func pivotRoot(root string) error {
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("mount rootfs to itself error: %v", err)
	}
	pivotDir := filepath.Join(root, ".pivot_root")
	if _, err := os.Stat(pivotDir); err == nil {
		// 存在则删除
		if err := os.Remove(pivotDir); err != nil {
			return err
		}
	}
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return fmt.Errorf("mkdir of pivot_root err: %v", err)
	}
	// pivot_root 到新的rootfs，老的old_root现在挂载在rootfs/.pivot_root上
	// 挂载点目前依然可以在mount命令中看到
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("pivot_root err: %v", err)
	}
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir root err: %v", err)
	}
	pivotDir = filepath.Join("/", ".pivot_root")
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount pivot_root dir err: %v", err)
	}
	return os.Remove(pivotDir)
}
