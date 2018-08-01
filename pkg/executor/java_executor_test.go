package executor

import (
	"syscall"
	"testing"
)

func TestHandleExit(t *testing.T) {
	executor := NewJavaExecutor()

	oomCode := int(syscall.SIGABRT) + 128
	oomHandled := executor.HandleExit(oomCode, 1)
	if oomHandled != 134 {
		t.Errorf("Handler returned wrong value. Got %d, want %d", oomHandled, 134)
	}

	termCode := int(syscall.SIGTERM) + 128
	termHandled := executor.HandleExit(termCode, 1)
	if termHandled != 0 {
		t.Errorf("Handler returned wrong value. Got %d, want %d", termHandled, 0)
	}

	sigintCode := int(syscall.SIGINT) + 128
	sigIntHandled := executor.HandleExit(sigintCode, 1)
	if sigIntHandled != 0 {
		t.Errorf("Handler returned wrong value. Got %d, want %d", sigIntHandled, 0)
	}

}
