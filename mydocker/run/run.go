package run

import (
	"MyDockerDemo/mydocker/cgroup"
	"MyDockerDemo/mydocker/cgroup/subsystem"
	"MyDockerDemo/mydocker/container"
	log "github.com/sirupsen/logrus"
	"os"
)

func Run(tty bool, cmdArray []string, config *subsystem.ResourceConfig) {
	parent := container.NewParentProcess(tty, cmdArray)
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

	log.Infof("parent process run")
	_ = parent.Wait()
	os.Exit(-1)
}
