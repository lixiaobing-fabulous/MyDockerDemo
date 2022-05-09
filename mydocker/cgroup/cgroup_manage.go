package cgroup

import "MyDockerDemo/mydocker/cgroup/subsystem"

type CgroupManager struct {
	CgroupName string
	Resource   *subsystem.ResourceConfig
}

func NewCgroupManager(cgroupName string) *CgroupManager {
	return &CgroupManager{
		CgroupName: cgroupName,
	}
}

func (c *CgroupManager) Apply(pid int) error {
	for _, ins := range subsystem.SubsystemIns {
		err := ins.Apply(c.CgroupName, pid)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *CgroupManager) Set(res *subsystem.ResourceConfig) error {
	for _, ins := range subsystem.SubsystemIns {
		err := ins.Set(c.CgroupName, res)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *CgroupManager) Destroy() error {
	for _, ins := range subsystem.SubsystemIns {
		err := ins.Remove(c.CgroupName)
		if err != nil {
			return err
		}
	}
	return nil
}
