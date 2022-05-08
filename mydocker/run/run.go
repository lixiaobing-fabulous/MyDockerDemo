package run

import (
	"MyDockerDemo/mydocker/container"
	log "github.com/sirupsen/logrus"
	"os"
)

func Run(tty bool, command string) {
	parent := container.NewParentProcess(tty, command)
	if err := parent.Start(); err != nil {
		log.Error(err)
		return
	}
	log.Infof("parent process run")
	_ = parent.Wait()
	os.Exit(-1)
}
