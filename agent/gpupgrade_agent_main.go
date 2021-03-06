package main

import (
	"os"
	"os/exec"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpupgrade/agent/services"
	"github.com/greenplum-db/gpupgrade/helpers"
	"github.com/greenplum-db/gpupgrade/utils"
	"github.com/spf13/cobra"
)

func main() {
	//debug.SetTraceback("all")
	//parser := flags.NewParser(&AllServices, flags.HelpFlag|flags.PrintErrors)
	//
	//_, err := parser.Parse()
	//if err != nil {
	//	os.Exit(utils.GetExitCodeForError(err))
	//}
	var logdir, statedir string
	var RootCmd = &cobra.Command{
		Use:   "gpupgrade_agent ",
		Short: "Start the Command Listener (blocks)",
		Long:  `Start the Command Listener (blocks)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			gplog.InitializeLogging("gpupgrade_agent", logdir)

			conf := services.AgentConfig{
				Port:     6416,
				StateDir: statedir,
			}

			commandExecer := func(command string, vars ...string) helpers.Command {
				return exec.Command(command, vars...)
			}
			agentServer := services.NewAgentServer(commandExecer, conf)
			agentServer.Start()

			agentServer.Stop()

			return nil
		},
	}

	RootCmd.Flags().StringVar(&logdir, "log-directory", "", "command_listener log directory")
	RootCmd.Flags().StringVar(&statedir, "state-directory", utils.GetStateDir(), "Agent state directory")

	if err := RootCmd.Execute(); err != nil {
		gplog.Error(err.Error())
		os.Exit(1)
	}
}
