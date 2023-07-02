package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func NewParentProcess(tty bool, containerName, rootUrl, mntUrl string, volume string, envSlice []string) (*exec.Cmd, *os.File) {
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
	} else {
		dirUrl := fmt.Sprintf(DefaultInfoLocation, containerName)
		if err := os.MkdirAll(dirUrl, 0622); err != nil {
			log.Errorf("mkdir dir %s, err: %v", dirUrl, err)
			return nil, nil
		}

		stdLogFilePath := dirUrl + ContainerLogFile
		stdLogFile, err := os.Create(stdLogFilePath)
		if err != nil {
			log.Errorf("create file %s, err: %v", stdLogFilePath, err)
			return nil, nil
		}

		cmd.Stdout = stdLogFile

	}
	cmd.ExtraFiles = []*os.File{readPipe}
	if err := newWorkSpace(rootUrl, mntUrl, volume, containerName); err != nil {
		log.Errorf("new work space err: %v", err)
		return nil, nil
	}
	cmd.Dir = mntUrl + containerName
	cmd.Env = append(os.Environ(), envSlice...)
	return cmd, writePipe
}

func newWorkSpace(rootUrl string, mntUrl string, volume string, containerName string) error {
	if err := createReadOnlyLayer(rootUrl); err != nil {
		return err
	}
	if err := createWriteLayer(rootUrl, containerName); err != nil {
		return err
	}
	if err := createMountPoint(rootUrl, mntUrl, containerName); err != nil {
		return err
	}
	if err := mountExtractVolume(rootUrl, mntUrl, volume, containerName); err != nil {
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
func createWriteLayer(rootUrl, containerName string) error {
	writeUrl := rootUrl + "writeLayer/" + containerName + "/"
	exist, err := pathExist(writeUrl)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if !exist {
		if err := os.MkdirAll(writeUrl, 0777); err != nil {
			return fmt.Errorf("create write layer failed: %v", err)
		}
	}
	return nil
}

func createTmpWorkDirLayer(rootUrl, containerName string) error {
	writeUrl := rootUrl + "temp/workdir/" + containerName + "/"
	exist, err := pathExist(writeUrl)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if !exist {
		if err := os.MkdirAll(writeUrl, 0777); err != nil {
			return fmt.Errorf("create write layer failed: %v", err)
		}
	}
	return nil
}
func createTmpMntWorkDirLayer(rootUrl, containerName string) error {
	writeUrl := rootUrl + "temp/mnt_workdir/" + containerName + "/"
	exist, err := pathExist(writeUrl)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if !exist {
		if err := os.MkdirAll(writeUrl, 0777); err != nil {
			return fmt.Errorf("create write layer failed: %v", err)
		}
	}
	return nil
}
func createTmpMntLowerDirLayer(rootUrl, containerName string) error {
	writeUrl := rootUrl + "temp/mnt_lowerdir/" + containerName + "/"
	exist, err := pathExist(writeUrl)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if !exist {
		if err := os.MkdirAll(writeUrl, 0777); err != nil {
			return fmt.Errorf("create write layer failed: %v", err)
		}
	}
	return nil
}

func createMountPoint(rootUrl string, mntUrl string, containerName string) error {
	if err := createTmpWorkDirLayer(rootUrl, containerName); err != nil {
		return err
	}
	mountPath := mntUrl + containerName + "/"
	exist, err := pathExist(mountPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if !exist {
		if err := os.MkdirAll(mountPath, 0777); err != nil {
			return fmt.Errorf("mkdir faild: %v", err)
		}
	}
	dirs := "lowerdir=" + rootUrl + "busybox" + ",upperdir=" + rootUrl + "writeLayer/" + containerName + ",workdir=" + rootUrl + "temp/workdir/" + containerName + "/"
	log.Infof("mount", "-t", "overlay", "-o", dirs, "overlay", mountPath)
	fmt.Println("mount", "-t", "overlay", "-o", dirs, "overlay", mountPath)
	cmd := exec.Command("mount", "-t", "overlay", "-o", dirs, "overlay", mountPath)

	// 把writeLayer和busybox目录mount到mnt目录下
	//dirs := "dirs=" + rootUrl + "writeLayer/" + containerName + ":" + rootUrl + "busybox"
	//cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mountPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mnt dir err: %v", err)
	}
	return nil
}
func mountExtractVolume(rootUrl, mntUrl, volume, containerName string) error {
	if volume == "" {
		return nil
	}
	volumeUrls := strings.Split(volume, ":")
	length := len(volumeUrls)
	if length != 2 || volumeUrls[0] == "" || volumeUrls[1] == "" {
		return fmt.Errorf("volume parameter input is not corrent")
	}
	if err := createTmpMntWorkDirLayer(rootUrl, containerName); err != nil {
		return err
	}
	if err := createTmpMntLowerDirLayer(rootUrl, containerName); err != nil {
		return err
	}
	return mountVolume(mntUrl+containerName+"/", volumeUrls, rootUrl+"temp/mnt_workdir/"+containerName+"/", rootUrl+"temp/mnt_lowerdir/"+containerName+"/")
}

func mountVolume(mntUrl string, volumeUrls []string, workdir string, lowerdir string) error {
	parentUrl := volumeUrls[0]
	exist, err := pathExist(parentUrl)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if !exist {
		// 使用mkdir all 递归创建文件夹
		if err := os.MkdirAll(parentUrl, 0777); err != nil {
			return fmt.Errorf("mkdir parent dir err: %v", err)
		}
	}
	containerUrl := mntUrl + volumeUrls[1]
	if err := os.MkdirAll(containerUrl, 0777); err != nil {
		return fmt.Errorf("mkdir container volume err: %v", err)
	}
	dirs := "lowerdir=" + lowerdir + ",upperdir=" + parentUrl + ",workdir=" + workdir
	log.Infof("mount", "-t", "overlay", "-o", dirs, "overlay", containerUrl)
	fmt.Println("mount", "-t", "overlay", "-o", dirs, "overlay", containerUrl)
	cmd := exec.Command("mount", "-t", "overlay", "-o", dirs, "overlay", containerUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mount volume err: %v", err)
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
