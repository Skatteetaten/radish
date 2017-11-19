package eve

import (
	"github.com/nixx/eve/pkg/reaper"
	"github.com/nixx/eve/pkg/signaler"
	"github.com/nixx/eve/pkg/executor"
	"syscall"
	"github.com/Sirupsen/logrus"
	"os"
	"os/signal"
)

func RunEve(args []string) {
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
