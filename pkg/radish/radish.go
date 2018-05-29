package radish

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/skatteetaten/radish/pkg/executor"
	"github.com/skatteetaten/radish/pkg/reaper"
	"github.com/skatteetaten/radish/pkg/signaler"
)

//RunRadish : main executor for Radish
func RunRadish(args []string) {
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
