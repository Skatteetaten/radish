package executor

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"

	"github.com/sirupsen/logrus"
)

//MemoryStrategy :
type MemoryStrategy string

const (
	//FAILSAFE :
	FAILSAFE MemoryStrategy = "Failsafe"
	//LARGER_HEAP :
	LARGER_HEAP MemoryStrategy = "LargerHeap"
)

type javaExitHandler struct {
}

type generatedJavaExecutor struct {
	javaExitHandler
	MemoryStrategy MemoryStrategy
}

//NewJavaExecutor :
func NewJavaExecutor(strategy MemoryStrategy) Executor {
	return &generatedJavaExecutor{
		javaExitHandler: javaExitHandler{},
		MemoryStrategy:  FAILSAFE,
	}
}

func (m *generatedJavaExecutor) Execute(args []string) *exec.Cmd {
	//TODO Generate Exec string and read all the config and secret stuff

	environ := os.Environ()
	processedEnviron := make([]string, len(environ))
	processedEnviron = append(processedEnviron, "EKSTRA_KONFIG=CONFIG")
	cmd := exec.Command("java", args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = processedEnviron
	return cmd
}

func (m *javaExitHandler) HandleExit(exitCode int, pid int) int {
	if exitCode == int(syscall.SIGABRT)+128 {
		logrus.Info("Java is out of memory! Bummer")
		printCoreFileToStdOut(pid)
		return exitCode
	}
	if exitCode == int(syscall.SIGINT)+128 {
		logrus.Info("Java terminated successfully from a SIGINT")
		return 0
	}
	if exitCode == int(syscall.SIGTERM)+128 {
		logrus.Info("Java terminated successfully from a SIGTERM")
		return 0
	}
	logrus.Info("%d", exitCode)
	return exitCode
}

func printCoreFileToStdOut(pid int) {
	report := fmt.Sprintf("hs_err_pid%d.log", pid)
	crashReport, err := ioutil.ReadFile(report)
	if err != nil {
		logrus.Errorf("Error reading crash report %s", report)
		return
	}
	logrus.Info(string(crashReport))
}
