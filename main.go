package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/concourse/fly/commands"
)

var targetFlag = cli.StringFlag{
	Name:  "target, t",
	Value: "http://192.168.100.4:8080",
	Usage: "named target you have saved to your .flyrc file",
}

var taskConfigFlag = cli.StringFlag{
	Name:  "config, c",
	Usage: "task configuration file",
}

var inputFlag = cli.StringSliceFlag{
	Name:  "input, i",
	Value: &cli.StringSlice{},
}

var inputsFromPipelineFlag = cli.StringFlag{
	Name:  "inputs-from-pipeline, ifp",
	Usage: "pipeline whose job the task will base its inputs on",
}

var inputsFromJobFlag = cli.StringFlag{
	Name:  "inputs-from-job, ifj",
	Usage: "job the task will base its inputs on",
}

var insecureFlag = cli.BoolFlag{
	Name:  "insecure, k",
	Usage: "allow insecure SSL connections and transfers",
}

var excludeIgnoredFlag = cli.BoolFlag{
	Name:  "exclude-ignored, x",
	Usage: "exclude vcs-ignored files from the build's inputs",
}

var privilegedFlag = cli.BoolFlag{
	Name:  "privileged, p",
	Usage: "run the task with full privileges",
}

var pipelineFlag = cli.StringFlag{
	Name:  "pipeline, p",
	Usage: "the name of the pipeline to act upon",
}

var checkFlag = cli.StringFlag{
	Name:  "check, c",
	Usage: "name of a resource's checking container to hijack",
}

var stepTypeFlag = cli.StringFlag{
	Name:  "step-type, t",
	Usage: "type of step to hijack. one of get, put, or task.",
}

var stepNameFlag = cli.StringFlag{
	Name:  "step-name, n",
	Usage: "name of step to hijack (e.g. build, unit, resource name)",
}

var varFlag = cli.StringSliceFlag{
	Name:  "var, v",
	Value: &cli.StringSlice{},
	Usage: "variable flag that can be used for filling in template values in configuration (i.e. -var secret=key)",
}

var varFileFlag = cli.StringSliceFlag{
	Name:  "vars-from, vf",
	Value: &cli.StringSlice{},
	Usage: "variable flag that can be used for filling in template values in configuration from a YAML file",
}

var executeFlags = []cli.Flag{
	taskConfigFlag,
	inputFlag,
	excludeIgnoredFlag,
	privilegedFlag,
	inputsFromPipelineFlag,
	inputsFromJobFlag,
}

func jobFlag(verb string) cli.StringFlag {
	return cli.StringFlag{
		Name:  "job, j",
		Usage: fmt.Sprintf("if specified, %s builds of the given job", verb),
	}
}

func buildFlag(verb string) cli.StringFlag {
	return cli.StringFlag{
		Name:  "build, b",
		Usage: fmt.Sprintf("%s a specific build of a job", verb),
	}
}

var pipelineConfigFlag = cli.StringFlag{
	Name:  "config, c",
	Usage: "pipeline configuration file",
}

var jsonFlag = cli.BoolFlag{
	Name:  "json, j",
	Usage: "print config as json instead of yaml",
}

var pausedFlag = cli.BoolFlag{
	Name:  "paused",
	Usage: "should the pipeline start out as paused or unpaused (true/false)",
}

var apiFlag = cli.StringFlag{
	Name:  "api",
	Usage: "api url to target",
}

var usernameFlag = cli.StringFlag{
	Name:  "username, user",
	Usage: "username for the api",
}

var passwordFlag = cli.StringFlag{
	Name:  "password, pass",
	Usage: "password for the api",
}

var certFlag = cli.StringFlag{
	Name:  "cert",
	Usage: "directory to your cert",
}

var githubPersonalAccessTokenFlag = cli.StringFlag{
	Name:  "github-personal-access-token",
	Usage: "generated token from github to authenticate",
}

func main() {
	app := cli.NewApp()
	app.Name = "fly"
	app.Usage = "Concourse CLI"
	app.Version = `¯\_(ツ)_/¯ (run "fly -t <target> sync")`
	app.Flags = []cli.Flag{
		insecureFlag,
		targetFlag,
	}
	app.Commands = []cli.Command{
		{
			Name:      "execute",
			ShortName: "e",
			Usage:     "Execute a build",
			Flags:     executeFlags,
			Action:    commands.Execute,
		},
		{
			Name:      "destroy-pipeline",
			ShortName: "d",
			Usage:     "Destroy a pipeline",
			ArgsUsage: "PIPELINE_NAME",
			Action:    commands.DestroyPipeline,
		},
		takeControl("hijack"),
		takeControl("intercept"),
		{
			Name:      "watch",
			ShortName: "w",
			Usage:     "Stream a build's log",
			Flags: []cli.Flag{
				buildFlag("watches"),
				jobFlag("watches"),
				pipelineFlag,
			},
			Action: commands.Watch,
		},
		{
			Name:      "configure",
			ShortName: "c",
			Usage:     "Update or download current configuration",
			ArgsUsage: "PIPELINE_NAME",
			Flags: []cli.Flag{
				pipelineConfigFlag,
				jsonFlag,
				varFlag,
				varFileFlag,
				pausedFlag,
			},
			Action: commands.Configure,
		},
		{
			Name:      "sync",
			ShortName: "s",
			Usage:     "Download and replace the current fly from the target",
			Action:    commands.Sync,
		},
		{
			Name:      "save-target",
			Usage:     "Save a fly target to the .flyrc",
			Action:    commands.SaveTarget,
			ArgsUsage: "TARGET_NAME",
			Flags: []cli.Flag{
				apiFlag,
				usernameFlag,
				passwordFlag,
				certFlag,
				githubPersonalAccessTokenFlag,
			},
		},
		{
			Name:      "checklist",
			ShortName: "l",
			Usage:     "Print a Checkman checkfile for the pipeline configuration",
			ArgsUsage: "PIPELINE_NAME",
			Action:    commands.Checklist,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}

func takeControl(commandName string) cli.Command {
	return cli.Command{
		Name:      commandName,
		ShortName: "i",
		Usage:     "Execute an interactive command in a build's container",
		Flags: []cli.Flag{
			jobFlag(commandName + "s"),
			buildFlag(commandName + "s"),
			pipelineFlag,
			stepTypeFlag,
			stepNameFlag,
			checkFlag,
		},
		Action: commands.Hijack,
	}
}
