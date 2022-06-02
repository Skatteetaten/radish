package nodejs

import (
	"os/exec"
)

//Executor :
type Executor interface {
	PrepareForNodeJSRun(mainJavaScriptFile string) *exec.Cmd
}

type nodeJSExecutor struct {
	nodeJSExitHandler
}

type nodeJSExitHandler struct {
}

//NewNodeJSExecutor :
func NewNodeJSExecutor() Executor {
	return nodeJSExecutor{
		nodeJSExitHandler{},
	}
}

func (m nodeJSExecutor) PrepareForNodeJSRun(mainJavaScriptFile string) *exec.Cmd {
	cmd := exec.Command("node", mainJavaScriptFile)
	return cmd
}
