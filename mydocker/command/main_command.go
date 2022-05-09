package command

import (
	"MyDockerDemo/mydocker/cgroup/subsystem"
	"MyDockerDemo/mydocker/container"
	"MyDockerDemo/mydocker/run"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
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
		// 添加-v标签
		cli.StringFlag{
			Name:  "v",
			Usage: "volume",
		},
		// 添加-d标签
		cli.BoolFlag{
			Name:  "d",
			Usage: "detach container",
		},
		// 提供run后面的-name指定容器名字参数
		cli.StringFlag{
			Name:  "name",
			Usage: "container name",
		},
		cli.StringSliceFlag{
			Name:  "e",
			Usage: "set environment",
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
		detach := context.Bool("d")
		resConfig := &subsystem.ResourceConfig{
			MemoryLimit: context.String("mem"),
			CpuShare:    context.String("cpuShare"),
			CpuSet:      context.String("cpuSet"),
		}
		volume := context.String("v")
		containerName := context.String("name")
		envSlice := context.StringSlice("e")

		run.Run(tty, detach, cmdArray, resConfig, volume, containerName, envSlice)
		return nil
	},
}

var CommitCommand = cli.Command{
	Name:  "commit",
	Usage: "commit a container into image",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("Missing container name")
		}
		imageName := context.Args().Get(0)
		return run.CommitContainer(imageName)
	},
}

var ListCommand = cli.Command{
	Name:  "ps",
	Usage: "list all the container",
	Action: func(context *cli.Context) error {
		return run.ListContainers()
	},
}

var LogCommand = cli.Command{
	Name:  "logs",
	Usage: "print logs of a container",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("Missing container name")
		}
		containerName := context.Args().Get(0)
		return run.LogContainer(containerName)
	},
}

var ExecCommand = cli.Command{
	Name:  "exec",
	Usage: "exec a command into container",
	Action: func(context *cli.Context) {
		if os.Getenv(run.EnvExecPid) != "" {
			log.Infof("pid callback pid %d", os.Getgid())
			return
		}

		// 我们希望命令格式是docker exec 容器名 命令
		if len(context.Args()) < 2 {
			log.Errorf("missing container name or command")
			return
		}

		containerName := context.Args().Get(0)
		var commandArray []string
		for _, arg := range context.Args().Tail() {
			commandArray = append(commandArray, arg)
		}

		// 执行命令
		if err := run.ExecContainer(containerName, commandArray); err != nil {
			log.Errorf("%v", err)
		}
	},
}

var StopCommand = cli.Command{
	Name:  "stop",
	Usage: "stop container",
	Action: func(context *cli.Context) {
		if len(context.Args()) < 1 {
			log.Errorf("missing container name")
			return
		}
		containerName := context.Args().Get(0)
		if err := run.StopContainer(containerName); err != nil {
			log.Errorf("stop container err: %v", err)
		}
	},
}
var RemoveCommand = cli.Command{
	Name:  "rm",
	Usage: "remove container",
	Action: func(context *cli.Context) {
		if len(context.Args()) < 1 {
			log.Errorf("missing container name")
			return
		}
		containerName := context.Args().Get(0)
		if err := run.RemoveContainer(containerName); err != nil {
			log.Errorf("%v", err)
		}
	},
}
