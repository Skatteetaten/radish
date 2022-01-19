package java

import (
	"github.com/stretchr/testify/assert"
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

func TestBuildClasspath(t *testing.T) {
	executor := NewJavaExecutor()
	descriptor := "testdata/testconfig.json"
	cp, err := executor.BuildClasspath(descriptor)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"testdata/lib/lib1.jar:testdata/lib/lib2.jar:testdata/lib/lib2/lib4.jar",
		cp,
	)
}

func TestBuildClasspathSubpath(t *testing.T) {
	executor := NewJavaExecutor()
	descriptor := "testdata/testconfig-subpath.json"
	cp, err := executor.BuildClasspath(descriptor)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"testdata/lib/lib1.jar:testdata/lib/lib2.jar:testdata/lib/lib2/lib4.jar:testdata/lib/lib3/lib4/lib6.jar:testdata/lib/lib3/lib5/lib7.jar:testdata/lib/lib3/lib5/lib8.jar:testdata/lib/lib3/lib5.jar",
		cp,
	)
}

func TestThatUnknownJavaVersionCausesPanic(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Errorf("The code did not panic")
		}
	}()

	resolveArgumentModificators(javaVersionLookupFor("UNKNOWN VERSION"))
}

func TestJava8ArgumentModificators(t *testing.T) {
	argumentModificators := resolveArgumentModificators(javaVersionLookupFor("8"))
	assert.Equal(t, Java8ArgumentsModificators, argumentModificators)
}

func TestJava11ArgumentModificators(t *testing.T) {
	argumentModificators := resolveArgumentModificators(javaVersionLookupFor("11"))
	assert.Equal(t, Java11ArgumentsModificators, argumentModificators)
}

func TestJava17ArgumentModificators(t *testing.T) {
	argumentModificators := resolveArgumentModificators(javaVersionLookupFor("17"))
	assert.Equal(t, Java17ArgumentsModificators, argumentModificators)
}

func javaVersionLookupFor(javaVersion string) func(string) string {
	return func(s string) string {
		return javaVersion
	}
}
