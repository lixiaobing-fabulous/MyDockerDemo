package run

import (
	_ "MyDockerDemo/mydocker/nsenter"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
)

const EnvExecPid = "mydocker_pid"
const EnvExecCmd = "mydocker_cmd"

func ExecContainer(containerName string, commandArray []string) error {
	pid, err := getContainerPidByName(containerName)
	if err != nil {
		return err
	}
	cmdStr := strings.Join(commandArray, " ")
	log.Infof("container pid %s", pid)
	log.Infof("command %s", cmdStr)
	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := os.Setenv(EnvExecPid, pid); err != nil {
		return fmt.Errorf("setenv %s err: %v", EnvExecPid, err)
	}
	if err := os.Setenv(EnvExecCmd, cmdStr); err != nil {
		return fmt.Errorf("setenv %s err: %v", EnvExecCmd, err)
	}
	envs, err := getEnvsByPid(pid)
	if err != nil {
		return err
	}
	cmd.Env = append(os.Environ(), envs...)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("exec container %s err: %v", containerName, err)
	}
	return nil

}
