package run

import (
	"MyDockerDemo/mydocker/cgroup"
	"MyDockerDemo/mydocker/cgroup/subsystem"
	"MyDockerDemo/mydocker/container"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

func Run(tty bool, cmdArray []string, config *subsystem.ResourceConfig) {
	parent, writePipe := container.NewParentProcess(tty)
	if err := parent.Start(); err != nil {
		log.Error(err)
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
	_ = parent.Wait()
	os.Exit(-1)
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
