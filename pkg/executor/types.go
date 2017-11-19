package executor

import "os/exec"

type Executor interface {
	/**
	Execute a process
	 */
	Execute(args []string) *exec.Cmd

	/**
	Handles exit code and rewrites exit code
	 */
	HandleExit(exitCode int, pid int) int
}
