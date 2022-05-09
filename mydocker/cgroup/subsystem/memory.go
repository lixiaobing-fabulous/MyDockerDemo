package subsystem

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type MemorySubsystem struct {
}

func (m MemorySubsystem) Name() string {
	return "memory"
}

func (m MemorySubsystem) Set(cgroupName string, res *ResourceConfig) error {
	cgroupPath, err := GetCgroupPath(m.Name(), cgroupName)
	if err != nil {
		return err
	}
	log.Infof("%s cgroup path: %s", m.Name(), cgroupPath)
	limitFilePath := path.Join(cgroupPath, "memory.limit_in_bytes")
	if err := ioutil.WriteFile(limitFilePath, []byte(res.MemoryLimit), 0644); err != nil {
		return fmt.Errorf("set memory cgroup failed: %v", err)
	}
	return nil
}

func (m MemorySubsystem) Apply(cgroupName string, pid int) error {
	cgroupPath, err := GetCgroupPath(m.Name(), cgroupName)
	if err != nil {
		return err
	}
	log.Infof("%s cgroup path: %s", m.Name(), cgroupPath)
	// 将PID加入该cgroup
	limitFilePath := path.Join(cgroupPath, "tasks")
	if err := ioutil.WriteFile(limitFilePath, []byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf("add pid to cgroup failed: %v", err)
	}
	return nil
}

func (m MemorySubsystem) Remove(cgroupName string) error {
	cgroupPath, err := GetCgroupPath(m.Name(), cgroupName)
	if err != nil {
		return err
	}
	log.Infof("%s cgroup path: %s", m.Name(), cgroupPath)
	return os.RemoveAll(cgroupPath)
}
