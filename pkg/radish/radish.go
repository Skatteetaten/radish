package radish

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/skatteetaten/radish/pkg/executor"
	"github.com/skatteetaten/radish/pkg/reaper"
	"github.com/skatteetaten/radish/pkg/signaler"
)

//RunRadish : main executor for Radish
func RunRadish(args []string) {
	logrus.Infof("args: %d", args)
	if args[1] == "bash" {
		logrus.Info("Inside bash if")
		cmd := exec.Command("/usr/bin/tail", "-f", "/dev/null")
		cmd.Start()
		pid := cmd.Process.Pid
		logrus.Infof("pid: %d", pid)
		var wstatus syscall.WaitStatus
		syscall.Wait4(int(pid), &wstatus, 0, nil)
		exitCode := wstatus.ExitStatus()
		logrus.Infof("Exit code bash %d", exitCode)
		os.Exit(int(exitCode))
	} else {
		e := executor.NewJavaExecutor(executor.FAILSAFE)
		cmd := e.Execute(args)
		err := cmd.Start()
		if err != nil {
			logrus.Fatalf("Error starting: %s", err)
		}
		reaper.Start()
		signal.Ignore(syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL)
		pid := cmd.Process.Pid
		signaler.Start(cmd.Process)
		var wstatus syscall.WaitStatus
		//processState, error := cmd.Process.Wait()
		//processState.
		syscall.Wait4(int(pid), &wstatus, 0, nil)
		logrus.Infof("Exit code %d", wstatus.ExitStatus())
		exitCode := e.HandleExit(wstatus.ExitStatus(), pid)
		os.Exit(int(exitCode))
	}

}
