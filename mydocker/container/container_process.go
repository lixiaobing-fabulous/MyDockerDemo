package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"syscall"
)

func NewParentProcess(tty bool, rootUrl, mntUrl string) (*exec.Cmd, *os.File) {
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		log.Errorf("create pipe error: %v", err)
		return nil, nil
	}

	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	cmd.ExtraFiles = []*os.File{readPipe}
	if err := newWorkSpace(rootUrl, mntUrl); err != nil {
		log.Errorf("new work space err: %v", err)
		return nil, nil
	}
	cmd.Dir = mntUrl
	return cmd, writePipe
}

func newWorkSpace(rootUrl string, mntUrl string) error {
	if err := createReadOnlyLayer(rootUrl); err != nil {
		return err
	}
	if err := createWriteLayer(rootUrl); err != nil {
		return err
	}
	if err := createMountPoint(rootUrl, mntUrl); err != nil {
		return err
	}
	return nil
}

// 我们直接把busybox放到了工程目录下，直接作为容器的只读层
func createReadOnlyLayer(rootUrl string) error {
	busyboxUrl := rootUrl + "busybox/"
	exist, err := pathExist(busyboxUrl)
	if err != nil {
		return err
	}
	if !exist {
		return fmt.Errorf("busybox dir don't exist: %s", busyboxUrl)
	}
	return nil
}

// 创建一个名为writeLayer的文件夹作为容器的唯一可写层
func createWriteLayer(rootUrl string) error {
	writeUrl := rootUrl + "writeLayer/"
	if err := os.Mkdir(writeUrl, 0777); err != nil {
		return fmt.Errorf("create write layer failed: %v", err)
	}
	return nil
}

func createMountPoint(rootUrl string, mntUrl string) error {
	// 创建mnt文件夹作为挂载点
	if err := os.Mkdir(mntUrl, 0777); err != nil {
		return fmt.Errorf("mkdir faild: %v", err)
	}
	// 把writeLayer和busybox目录mount到mnt目录下
	dirs := "dirs=" + rootUrl + "writeLayer:" + rootUrl + "busybox"
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mmt dir err: %v", err)
	}
	return nil
}

func pathExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, err
	}
	return false, err
}
