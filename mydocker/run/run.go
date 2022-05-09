package run

import (
	"MyDockerDemo/mydocker/cgroup"
	"MyDockerDemo/mydocker/cgroup/subsystem"
	"MyDockerDemo/mydocker/container"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func Run(tty, detach bool, cmdArray []string, config *subsystem.ResourceConfig, volume string, containerName string) {
	pwd, err := os.Getwd()
	if err != nil {
		log.Errorf("Run get pwd err: %v", err)
		return
	}
	mntUrl := pwd + "/mnt/"
	rootUrl := pwd + "/"
	parent, writePipe := container.NewParentProcess(tty, containerName, rootUrl, mntUrl, volume)
	if err := parent.Start(); err != nil {
		log.Error(err)
		deleteWorkSpace(rootUrl, mntUrl, volume)
		return
	}

	// 记录容器信息
	containerName, err = recordContainerInfo(parent.Process.Pid, cmdArray, containerName)
	if err != nil {
		log.Errorf("record contariner info err: %v", err)
		return
	}

	cgroupManager := cgroup.NewCgroupManager("mydocker-cgroup")
	defer cgroupManager.Destroy()
	if err := cgroupManager.Apply(parent.Process.Pid); err != nil {
		log.Errorf("cgroup apply err: %v", err)
		return
	}
	if err := cgroupManager.Set(config); err != nil {
		log.Errorf("cgoup set err: %v", err)
		return
	}
	sendInitCommand(cmdArray, writePipe)
	log.Infof("parent process run")
	if !detach {
		_ = parent.Wait()
		deleteWorkSpace(rootUrl, mntUrl, volume)
		deleteContainerInfo(containerName)
	}
	os.Exit(-1)
}
func deleteContainerInfo(containerName string) {
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	if err := os.RemoveAll(dirUrl); err != nil {
		log.Errorf("remove dir %s err: %v", dirUrl, err)
	}
}

func recordContainerInfo(pid int, cmdArray []string, containerName string) (string, error) {
	id := randStringBytes(10)
	createTime := time.Now().Format("2000-01-01 00:00:00")
	command := strings.Join(cmdArray, " ")
	if containerName == "" {
		containerName = id
	}
	containerInfo := &container.ContainerInfo{
		ID:         id,
		Pid:        strconv.Itoa(pid),
		Command:    command,
		CreateTime: createTime,
		Status:     container.RUNNING,
		Name:       containerName,
	}
	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		return "", fmt.Errorf("container info to json string err: %v", err)
	}
	jsonStr := string(jsonBytes)
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	if err := os.MkdirAll(dirUrl, 0622); err != nil {
		return "", fmt.Errorf("mkdir %s err: %v", dirUrl, err)
	}
	fileName := dirUrl + "/" + container.ConfigName
	file, err := os.Create(fileName)
	defer file.Close()
	if err != nil {
		return "", fmt.Errorf("create file %s, err: %v", fileName, err)
	}

	if _, err := file.WriteString(jsonStr); err != nil {
		return "", fmt.Errorf("file write string err: %v", err)
	}
	return containerName, nil
}

func randStringBytes(n int) string {
	letterBytes := "1234567890"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
func deleteWorkSpace(rootUrl, mntUrl, volume string) {
	unmountVolume(mntUrl, volume)
	deleteMountPoint(mntUrl)
	deleteWriteLayer(rootUrl)
}
func unmountVolume(mntUrl string, volume string) {
	if volume == "" {
		return
	}
	volumeUrls := strings.Split(volume, ":")
	if len(volumeUrls) != 2 || volumeUrls[0] == "" || volumeUrls[1] == "" {
		return
	}

	// 卸载容器内的 volume 挂载点的文件系统
	containerUrl := mntUrl + volumeUrls[1]
	cmd := exec.Command("umount", containerUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("ummount volume failed: %v", err)
	}
}
func deleteMountPoint(mntUrl string) {
	cmd := exec.Command("umount", mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("deleteMountPoint umount %s err : %v", mntUrl, err)
	}
	if err := os.RemoveAll(mntUrl); err != nil {
		log.Errorf("deleteMountPoint remove %s err : %v", mntUrl, err)
	}
}
func deleteWriteLayer(rootUrl string) {
	writeUrl := rootUrl + "writeLayer/"
	if err := os.RemoveAll(writeUrl); err != nil {
		log.Errorf("deleteMountPoint remove %s err : %v", writeUrl, err)
	}
}

func sendInitCommand(array []string, writePipe *os.File) {
	command := strings.Join(array, " ")
	log.Infof("all command is : %s", command)
	if _, err := writePipe.WriteString(command); err != nil {
		log.Errorf("write pipe write string err: %v", err)
		return
	}
	if err := writePipe.Close(); err != nil {
		log.Errorf("write pipe close err: %v", err)
	}
}
