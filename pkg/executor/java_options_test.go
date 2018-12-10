package executor

import (
	"github.com/skatteetaten/radish/pkg/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJava8Options(t *testing.T) {
	env := make(map[string]string)
	env["ENABLE_JOLOKIA"] = "true"
	env["ENABLE_JAVA_DIAGNOSTICS"] = "true"
	env["ENABLE_REMOTE_DEBUG"] = "true"
	env["APPDYNAMICS_AGENT_BASE_DIR"] = "/opt/appdynamics"
	ctx := createTestContext(env)
	modifiedArgs := applyArguments(Java8ArgumentsModificators, ctx)
	assert.Contains(t, modifiedArgs, "-javaagent:jolokia.jar=host=0.0.0.0,port=8778,protocol=https")
	assert.Contains(t, modifiedArgs, "-Xmx2048m")
	assert.Contains(t, modifiedArgs, "-Xms2048m")
	assert.Contains(t, modifiedArgs, "-Djava.util.concurrent.ForkJoinPool.common.parallelism=4")
	assert.Contains(t, modifiedArgs, "-XX:ConcGCThreads=4")
	assert.Contains(t, modifiedArgs, "-XX:ParallelGCThreads=4")
	assert.Contains(t, modifiedArgs, "-XX:NativeMemoryTracking=summary")
	assert.Contains(t, modifiedArgs, "-XX:+PrintGC")
	assert.Contains(t, modifiedArgs, "-XX:+PrintGCDateStamps")
	assert.Contains(t, modifiedArgs, "-XX:+PrintGCTimeStamps")
	assert.Contains(t, modifiedArgs, "-XX:+UnlockDiagnosticVMOptions")
	assert.Contains(t, modifiedArgs, "-agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=5005")
	assert.NotContains(t, modifiedArgs, "-javaagent:/opt/appdynamics/javaagent.jar")
}

func TestJava11Options(t *testing.T) {
	env := make(map[string]string)
	env["ENABLE_JOLOKIA"] = "true"
	env["ENABLE_JAVA_DIAGNOSTICS"] = "true"
	env["ENABLE_REMOTE_DEBUG"] = "true"
	env["APPDYNAMICS_AGENT_BASE_DIR"] = "/opt/appdynamics"
	ctx := createTestContext(env)
	modifiedArgs := applyArguments(Java11ArgumentsModificators, ctx)
	assert.Contains(t, modifiedArgs, "-javaagent:jolokia.jar=host=0.0.0.0,port=8778,protocol=https")
	assert.Contains(t, modifiedArgs, "-XX:NativeMemoryTracking=summary")
	assert.Contains(t, modifiedArgs, "-Xlog:gc")
	assert.Contains(t, modifiedArgs, "-agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=5005")
	assert.Len(t, modifiedArgs, 7)
}

func TestOptionsJolokia(t *testing.T) {
	env := make(map[string]string)
	env["ENABLE_JOLOKIA"] = "true"
	ctx := createTestContext(env)
	modifiedArgs := applyArguments(Java11ArgumentsModificators, ctx)
	assert.Contains(t, modifiedArgs, "-javaagent:jolokia.jar=host=0.0.0.0,port=8778,protocol=https")
}

func TestOptionsNoJolokia(t *testing.T) {
	env := make(map[string]string)
	ctx := createTestContext(env)
	modifiedArgs := applyArguments(Java11ArgumentsModificators, ctx)
	assert.NotContains(t, modifiedArgs, "-javaagent:jolokia.jar=host=0.0.0.0,port=8778,protocol=https")
}

func TestOptionsAppDynamics(t *testing.T) {
	env := make(map[string]string)
	env["OPENSHIFT_CLUSTER"] = "test"
	env["ENABLE_JOLOKIA"] = "true"
	env["ENABLE_JAVA_DIAGNOSTICS"] = "false"
	env["ENABLE_REMOTE_DEBUG"] = "false"
	env["ENABLE_APPDYNAMICS"] = "true"
	env["APPDYNAMICS_AGENT_BASE_DIR"] = "/opt/appdynamics"
	env["POD_NAMESPACE"] = "mynamespace"
	env["APP_NAME"] = "myappname"
	env["POD_NAME"] = "mypodname"
	ctx := createTestContext(env)
	modifiedArgs := applyArguments(Java8ArgumentsModificators, ctx)
	assert.Contains(t, modifiedArgs, "-javaagent:jolokia.jar=host=0.0.0.0,port=8778,protocol=https")
	assert.Contains(t, modifiedArgs, "-Xmx2048m")
	assert.Contains(t, modifiedArgs, "-Xms2048m")
	assert.Contains(t, modifiedArgs, "-Djava.util.concurrent.ForkJoinPool.common.parallelism=4")
	assert.Contains(t, modifiedArgs, "-XX:ConcGCThreads=4")
	assert.Contains(t, modifiedArgs, "-XX:ParallelGCThreads=4")
	assert.NotContains(t, modifiedArgs, "-XX:NativeMemoryTracking=summary")
	assert.NotContains(t, modifiedArgs, "-agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=5005")
	assert.Contains(t, modifiedArgs, "-javaagent:/opt/appdynamics/javaagent.jar")
	assert.Contains(t, modifiedArgs, "-Dappdynamics.agent.applicationName=mynamespace-test")
	assert.Contains(t, modifiedArgs, "-Dappdynamics.agent.tierName=myappname")
	assert.Contains(t, modifiedArgs, "-Dappdynamics.agent.nodeName=mypodname")
}

func TestReadingOfJavaOptionsInDescriptor(t *testing.T) {
	env["VARIABLE_TO_EXPAND"] = "jallaball"
	ctx := createTestContext(env)
	ctx.Descriptor.Data.JavaOptions = "-Dtest.tull1 -Dtest2"
	args := applyArguments(Java8ArgumentsModificators, ctx)
	assert.Contains(t, args, "-Dtest.tull1")
	assert.Contains(t, args, "-Dtest2")
	ctx.Descriptor.Data.JavaOptions = "\"-Dtest.tull1 -Dtest2\""
	args = applyArguments(Java8ArgumentsModificators, ctx)
	assert.Contains(t, args, "-Dtest.tull1 -Dtest2")
}

func TestReadingOfJavaOptionsInEnv(t *testing.T) {
	env["JAVA_OPTIONS"] = "-Xtulleball -Xjallaball"
	ctx := createTestContext(env)
	args := applyArguments(Java8ArgumentsModificators, ctx)
	assert.Contains(t, args, "-Xtulleball")
	assert.Contains(t, args, "-Xjallaball")
}

func TestJavaDiagnostics(t *testing.T) {
	env["ENABLE_JAVA_DIAGNOSTICS"] = "true"
	ctx := createTestContext(env)
	args := applyArguments(Java8ArgumentsModificators, ctx)
	diagnostics := []string{"-XX:NativeMemoryTracking=summary",
		"-XX:+PrintGC",
		"-XX:+PrintGCDateStamps",
		"-XX:+PrintGCTimeStamps",
		"-XX:+UnlockDiagnosticVMOptions"}
	assert.Subset(t, args, diagnostics)
	env["ENABLE_JAVA_DIAGNOSTICS"] = "0"
	ctx = createTestContext(env)
	args = applyArguments(Java8ArgumentsModificators, ctx)
	for _, d := range diagnostics {
		assert.NotContains(t, args, d)
	}
}

func TestJavaMaxMemRatio(t *testing.T) {
	m := &memoryOptions{}
	env = make(map[string]string)
	ctx := createTestContext(env)
	args := m.modifyArguments(ctx)
	assert.Contains(t, args, "-Xmx2048m")
	assert.Contains(t, args, "-Xms2048m")
	env["JAVA_MAX_MEM_RATIO"] = "50"
	ctx = createTestContext(env)
	args = m.modifyArguments(ctx)
	assert.Contains(t, args, "-Xmx4096m")
	assert.Contains(t, args, "-Xms4096m")
}

func TestJavaMaxMetaspaceMemRatio(t *testing.T) {
	env = make(map[string]string)
	env["JAVA_MAX_METASPACE_RATIO"] = "5"
	ctx := createTestContext(env)
	args := applyArguments(Java8ArgumentsModificators, ctx)
	assert.Contains(t, args, "-XX:MaxMetaspaceSize=409m")
	delete(env, "JAVA_MAX_METASPACE_RATIO")
	ctx = createTestContext(env)
	args = applyArguments(Java8ArgumentsModificators, ctx)
	for _, arg := range args {
		assert.NotRegexp(t, "-XX:MaxMetaspaceSize.*", arg)
	}

}

func TestExitOnOom(t *testing.T) {
	env["ENABLE_EXIT_ON_OOM"] = "1"
	ctx := createTestContext(env)
	args := applyArguments(Java8ArgumentsModificators, ctx)
	assert.Contains(t, args, "-XX:+ExitOnOutOfMemoryError")
}

func createTestContext(env map[string]string) ArgumentsContext {
	desc := JavaDescriptor{}
	limits := util.CGroupLimits{
		MemoryLimitInBytes: 1024 * 1024 * 1024 * 8,
		MaxCoresEstimated:  4,
	}
	env["JOLOKIA_PATH"] = "jolokia.jar"
	desc.Data.JavaOptions = "-Dtest.tull1 -Dtest2"
	ctx := ArgumentsContext{
		CGroupLimits: limits,
		Descriptor:   desc,
		Environment: func(key string) (string, bool) {
			k, e := env[key]
			return k, e
		},
	}
	return ctx
}
