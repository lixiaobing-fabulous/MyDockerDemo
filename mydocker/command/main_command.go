package command

import (
	"MyDockerDemo/mydocker/cgroup/subsystem"
	"MyDockerDemo/mydocker/container"
	"MyDockerDemo/mydocker/run"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"strings"
)

var InitCommand = cli.Command{
	Name:  "init",
	Usage: "Init container process run user's process in container. Do not call it outside",
	Action: func(context *cli.Context) error {
		log.Infof("init come on")
		cmd := context.Args().Get(0)
		args := strings.Split(context.Args().Get(1), " ")
		log.Infof("command: %s, args: %s", cmd, args)
		return container.RunContainerInitProcess()
	},
}

var RunCommand = cli.Command{
	Name:  "run",
	Usage: `Create a container with namespace and cgroups limit mydocker run -ti [command]`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "ti",
			Usage: "enable tty",
		},
		cli.StringFlag{
			Name:  "mem",
			Usage: "memory limit",
		},
		cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit",
		},
		cli.StringFlag{
			Name:  "cpushare",
			Usage: "cpushare limit",
		},
	},
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container command")
		}
		var cmdArray []string
		for _, arg := range context.Args() {
			cmdArray = append(cmdArray, arg)
		}
		tty := context.Bool("ti")
		resConfig := &subsystem.ResourceConfig{
			MemoryLimit: context.String("mem"),
			CpuShare:    context.String("cpuShare"),
			CpuSet:      context.String("cpuSet"),
		}
		run.Run(tty, cmdArray, resConfig)
		return nil
	},
}
