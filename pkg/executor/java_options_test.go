package executor

import (
	"github.com/skatteetaten/radish/pkg/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

//TODO: Negative tests!
func TestOptions(t *testing.T) {
	env := make(map[string]string)
	args := make([]string, 0)

	env["ENABLE_JOLOKIA"] = "true"
	env["JOLOKIA_PATH"] = "jolokia.jar"
	env["ENABLE_DIAGNOSTICS"] = "true"
	env["ENABLE_REMOTE_DEBUG"] = "true"
	env["APPDYNAMICS_AGENT_BASE_DIR"] = "/opt/appdynamics"
	ctx := ArgumentsContext{
		Arguments: args,
		Environment: func(key string) (string, bool) {
			k, e := env[key]
			return k, e
		},
		CGroupLimits: util.CGroupLimits{
			MemoryLimitInBytes: 1024 * 1024 * 1024 * 8,
			MaxCoresEstimated:  4,
		},
	}
	modifiedArgs := applyArguments(ArgumentsModificators, ctx)
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

func TestOptionsAppDynamics(t *testing.T) {
	env := make(map[string]string)
	args := make([]string, 0)

	env["ENABLE_JOLOKIA"] = "true"
	env["JOLOKIA_PATH"] = "jolokia.jar"
	env["ENABLE_DIAGNOSTICS"] = "false"
	env["ENABLE_REMOTE_DEBUG"] = "false"
	env["ENABLE_APPDYNAMICS"] = "true"
	env["APPDYNAMICS_AGENT_BASE_DIR"] = "/opt/appdynamics"
	env["POD_NAMESPACE"] = "mynamespace"
	env["APP_NAME"] = "myappname"
	env["POD_NAME"] = "mypodname"
	ctx := ArgumentsContext{
		Arguments: args,
		Environment: func(key string) (string, bool) {
			k, e := env[key]
			return k, e
		},
		CGroupLimits: util.CGroupLimits{
			MemoryLimitInBytes: 1024 * 1024 * 1024 * 8,
			MaxCoresEstimated:  4,
		},
	}
	modifiedArgs := applyArguments(ArgumentsModificators, ctx)
	assert.Contains(t, modifiedArgs, "-javaagent:jolokia.jar=host=0.0.0.0,port=8778,protocol=https")
	assert.Contains(t, modifiedArgs, "-Xmx2048m")
	assert.Contains(t, modifiedArgs, "-Xms2048m")
	assert.Contains(t, modifiedArgs, "-Djava.util.concurrent.ForkJoinPool.common.parallelism=4")
	assert.Contains(t, modifiedArgs, "-XX:ConcGCThreads=4")
	assert.Contains(t, modifiedArgs, "-XX:ParallelGCThreads=4")
	assert.NotContains(t, modifiedArgs, "-XX:NativeMemoryTracking=summary")
	assert.NotContains(t, modifiedArgs, "-agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=5005")
	assert.Contains(t, modifiedArgs, "-javaagent:/opt/appdynamics/javaagent.jar")
	assert.Contains(t, modifiedArgs, "-Dappdynamics.agent.applicationName=mynamespace")
	assert.Contains(t, modifiedArgs, "-Dappdynamics.agent.tierName=myappname")
	assert.Contains(t, modifiedArgs, "-Dappdynamics.agent.nodeName=mypodname")
}
