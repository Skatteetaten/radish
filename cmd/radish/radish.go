package radish

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/sirupsen/logrus"
)

//RunPlaceholder
var RunPlaceholder = &cobra.Command{
	Use:   "bash",
	Short: "Run radish with a placeholder process",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Info("Inside bash placeholder")
		comd := exec.Command("/usr/bin/tail", "-f", "/dev/null")
		comd.Start()
		pid := comd.Process.Pid
		logrus.Infof("pid: %d", pid)
		var wstatus syscall.WaitStatus
		syscall.Wait4(int(pid), &wstatus, 0, nil)
		exitCode := wstatus.ExitStatus()
		logrus.Infof("Exit code bash %d", exitCode)
		os.Exit(int(exitCode))

	},
}
