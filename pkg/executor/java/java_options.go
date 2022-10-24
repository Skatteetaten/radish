package java

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/kballard/go-shellquote"
	"github.com/sirupsen/logrus"
	"github.com/skatteetaten/radish/pkg/util"
)

// ArgumentsContext :
type ArgumentsContext struct {
	Arguments    []string
	Environment  func(string) (string, bool)
	Descriptor   descriptor
	CGroupLimits util.CGroupLimits
}

// ArgumentModificator :
type ArgumentModificator interface {
	shouldModifyArguments(context ArgumentsContext) bool
	modifyArguments(context ArgumentsContext) []string
}

// Java8ArgumentsModificators :
var Java8ArgumentsModificators = []ArgumentModificator{
	&environmentJavaOptionsOverride{},
	&descriptorJavaOptionsOverride{},
	&enableExitOnOom{},
	&debugOptions{},
	&diagnosticsOptions{},
	&jolokiaOptions{},
	&appDynamicsOptions{},
	&otelOptions{},
	&cpuCoreTuning{},
	&java8MemoryOptions{},
	&metaspaceOptions{},
	&heapDumpOptions{},
}

// Java11ArgumentsModificators :
var Java11ArgumentsModificators = []ArgumentModificator{
	&environmentJavaOptionsOverride{},
	&descriptorJavaOptionsOverride{},
	&enableExitOnOom{},
	&debugOptions{},
	&java11PlusDiagnosticsOptions{},
	&jolokiaOptions{},
	&appDynamicsOptions{},
	&otelOptions{},
	&java11PlusMemoryOptions{},
	&heapDumpOptions{},
}

// Java17ArgumentsModificators :
var Java17ArgumentsModificators = Java11ArgumentsModificators

type java11PlusDiagnosticsOptions struct {
	diagnosticsOptions
}

func (m *java11PlusDiagnosticsOptions) modifyArguments(context ArgumentsContext) []string {
	args := make([]string, 0, 3)
	args = append(args, "-XX:NativeMemoryTracking=summary",
		"-Xlog:gc",
		"-XX:+UnlockDiagnosticVMOptions")
	args = append(args, context.Arguments...)
	return args
}

type diagnosticsOptions struct {
}

func (m *diagnosticsOptions) shouldModifyArguments(context ArgumentsContext) bool {
	value, exists := context.Environment("ENABLE_JAVA_DIAGNOSTICS")
	return exists && strings.ToUpper(value) == "TRUE"
}

func (m *diagnosticsOptions) modifyArguments(context ArgumentsContext) []string {
	args := make([]string, 0, 3)
	args = append(args, "-XX:NativeMemoryTracking=summary",
		"-XX:+PrintGC",
		"-XX:+PrintGCDateStamps",
		"-XX:+PrintGCTimeStamps",
		"-XX:+UnlockDiagnosticVMOptions")
	args = append(args, context.Arguments...)
	return args
}

type enableExitOnOom struct {
}

func (m *enableExitOnOom) shouldModifyArguments(context ArgumentsContext) bool {
	value, exists := context.Environment("ENABLE_EXIT_ON_OOM")
	return exists && len(value) > 0
}

func (m *enableExitOnOom) modifyArguments(context ArgumentsContext) []string {
	return append([]string{"-XX:+ExitOnOutOfMemoryError"}, context.Arguments...)
}

type debugOptions struct {
}

func (m *debugOptions) shouldModifyArguments(context ArgumentsContext) bool {
	value, exists := context.Environment("ENABLE_REMOTE_DEBUG")
	return exists && strings.ToUpper(value) == "TRUE"
}

func (m *debugOptions) modifyArguments(context ArgumentsContext) []string {
	args := make([]string, 0)
	portAsString, exists := context.Environment("DEBUG_PORT")
	var port int
	var err error
	if port, err = strconv.Atoi(portAsString); err != nil || !exists {
		port = 5005
	}
	debugArgument := fmt.Sprintf("-agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=%d", port)
	args = append([]string{debugArgument}, context.Arguments...)
	return args
}

type environmentJavaOptionsOverride struct {
}

func (m *environmentJavaOptionsOverride) shouldModifyArguments(context ArgumentsContext) bool {
	_, exists := context.Environment("JAVA_OPTIONS")
	return exists
}

func (m *environmentJavaOptionsOverride) modifyArguments(context ArgumentsContext) []string {
	options, _ := context.Environment("JAVA_OPTIONS")
	splittedArgs, err := shellquote.Split(options)
	if err != nil {
		logrus.Error("Unable to parse JAVA_OPTONS from environment", options, err)
	}
	args := append(context.Arguments, splittedArgs...)
	return args
}

type descriptorJavaOptionsOverride struct {
}

func (m *descriptorJavaOptionsOverride) shouldModifyArguments(context ArgumentsContext) bool {
	return len(context.Descriptor.Data.JavaOptions) != 0
}

func (m *descriptorJavaOptionsOverride) modifyArguments(context ArgumentsContext) []string {
	options := context.Descriptor.Data.JavaOptions
	splittedArgs, err := shellquote.Split(options)
	if err != nil {
		logrus.Errorf("Unable to parse args from radish descriptor: %s %s", options, err)
	}
	args := append(context.Arguments, splittedArgs...)
	return args
}

var cpuCoreArguments = []string{"-XX:ParallelGCThreads",
	"-XX:ConcGCThreads",
	"-Djava.util.concurrent.ForkJoinPool.common.parallelism"}

type cpuCoreTuning struct {
}

func (m *cpuCoreTuning) shouldModifyArguments(context ArgumentsContext) bool {
	return !containsArgument(context.Arguments, cpuCoreArguments...) && context.CGroupLimits.HasCoreLimit()
}

func (m *cpuCoreTuning) modifyArguments(context ArgumentsContext) []string {
	args := removeArguments(context.Arguments, memoryArguments)
	limits := context.CGroupLimits
	if limits.HasCoreLimit() {
		args = append([]string{fmt.Sprintf("-XX:ParallelGCThreads=%d", limits.MaxCoresEstimated)}, args...)
		args = append([]string{fmt.Sprintf("-XX:ConcGCThreads=%d", limits.MaxCoresEstimated)}, args...)
		args = append([]string{fmt.Sprintf("-Djava.util.concurrent.ForkJoinPool.common.parallelism=%d", limits.MaxCoresEstimated)}, args...)
	}
	return args
}

type jolokiaOptions struct {
}

func (m *jolokiaOptions) shouldModifyArguments(context ArgumentsContext) bool {
	value, exists := context.Environment("ENABLE_JOLOKIA")
	return exists && (strings.ToUpper(value) == "TRUE")
}

func (m *jolokiaOptions) modifyArguments(context ArgumentsContext) []string {
	jolokiaPath, exists := context.Environment("JOLOKIA_PATH")
	args := make([]string, 0)
	if !exists {
		logrus.Warn("Jolokia was supposed to be enabled, but no Jolokia-path found")
		return context.Arguments
	}
	jolokiaArgument := fmt.Sprintf("-javaagent:%s=host=0.0.0.0,port=8778,protocol=https", jolokiaPath)
	args = append([]string{jolokiaArgument}, context.Arguments...)
	return args
}

type otelOptions struct {
}

func (m *otelOptions) shouldModifyArguments(context ArgumentsContext) bool {
	value, exists := context.Environment("ENABLE_OTEL_TRACE")
	return exists && strings.ToUpper(value) == "TRUE"
}

func (m *otelOptions) modifyArguments(context ArgumentsContext) []string {
	opentelemetryBaseDir, exists := context.Environment("OPENTELEMETRY_AGENT_BASE_DIR")
	args := make([]string, 0)
	if !exists {
		logrus.Error("Opentelemetry was supposed to be enabled, but no path found")
		return context.Arguments
	}

	appName, exists := context.Environment("APP_NAME")
	if !exists {
		logrus.Error("Required value APP_NAME not found. Opentelemetry agent is disabled")
		return context.Arguments
	}
	namespace, exists := context.Environment("POD_NAMESPACE")
	if !exists {
		logrus.Error("Required value POD_NAMESPACE not found. Opentelemetry agent is disabled")
		return context.Arguments
	}
	cluster, exists := context.Environment("OPENSHIFT_CLUSTER")
	if !exists {
		logrus.Error("Required value POD_NAMESPACE not found. Opentelemetry agent is disabled")
		return context.Arguments
	}
	// uniqueHostId used to identify POD's by AppD machine agent
	otelAgentArguments := fmt.Sprintf("-javaagent:%s/opentelemetry-javaagent.jar", opentelemetryBaseDir)
	args = append([]string{otelAgentArguments})

	args = append(args,
		fmt.Sprintf("-Dotel.resource.attributes=service.name=%s,service.namespace=%s,service.cluster=%s", appName, namespace, cluster),
	)

	args = append(args, context.Arguments...)
	return args
}

type appDynamicsOptions struct {
}

func (m *appDynamicsOptions) shouldModifyArguments(context ArgumentsContext) bool {
	value, exists := context.Environment("ENABLE_APPDYNAMICS")
	return exists && strings.ToUpper(value) == "TRUE"
}

func (m *appDynamicsOptions) modifyArguments(context ArgumentsContext) []string {
	appDynamicsBaseDir, exists := context.Environment("APPDYNAMICS_AGENT_BASE_DIR")
	args := make([]string, 0)
	if !exists {
		logrus.Error("AppDynamics was supposed to be enabled, but no path found")
		return context.Arguments
	}
	// Need to set app, tier and node name.
	// For daemonsets some variables are not present, eks. POD_NAME.
	agentAppName, exists := context.Environment("APPDYNAMICS_AGENT_APPLICATION_NAME")
	if !exists {
		appNameSpace, exists := context.Environment("POD_NAMESPACE")
		if !exists {
			logrus.Error("AppDynamics has no APPLICATION_NAME associated to it. Agent will not be enabled!")
			return context.Arguments
		}
		agentAppName = appNameSpace
	}

	// Two AppD controller clusters represent all OCP clusters. We need at least som unique app_name
	openshiftCluster, exists := context.Environment("OPENSHIFT_CLUSTER")
	if exists {
		appDynamicsEnableClusterSuffix, exists := context.Environment("APPDYNAMICS_ENABLE_CLUSTER_SUFFIX")
		if !exists || exists && strings.ToUpper(appDynamicsEnableClusterSuffix) != "FALSE" {
			agentAppName += "-" + openshiftCluster
		}
	}

	agentTierName, exists := context.Environment("APPDYNAMICS_AGENT_TIER_NAME")
	if !exists {
		appName, exists := context.Environment("APP_NAME")
		if exists {
			agentTierName = appName
		} else {
			appServiceName, exists := context.Environment("SERVICE_NAME")
			if !exists {
				logrus.Error("AppDynamics has no TIER_NAME associated to it. Agent will not be enabled!")
				return context.Arguments
			}
			agentTierName = appServiceName
		}
	}

	agentNodeName, exists := context.Environment("APPDYNAMICS_AGENT_NODE_NAME")
	if !exists {
		appPodName, exists := context.Environment("POD_NAME")
		if exists {
			agentNodeName = appPodName
		} else {
			appHostName, exists := context.Environment("HOSTNAME")
			if !exists {
				logrus.Error("AppDynamics has no NODE_NAME associated to it. Agent will not be enabled!")
				return context.Arguments
			}
			agentNodeName = appHostName
		}
	}

	agentNodeMarkAsHistorical, exists := context.Environment("APPDYNAMICS_JVM_SHUTDOWN_MARK_NODE_AS_HISTORICAL")
	if !exists {
		agentNodeMarkAsHistorical = "true"
	}

	// uniqueHostId used to identify POD's by AppD machine agent
	appDynamicsArgument := fmt.Sprintf("-javaagent:%s/javaagent.jar", appDynamicsBaseDir)
	args = append([]string{appDynamicsArgument})
	args = append(args, "-Dappdynamics.agent.applicationName="+agentAppName,
		"-Dappdynamics.agent.tierName="+agentTierName,
		"-Dappdynamics.agent.nodeName="+agentNodeName,
		"-Dappdynamics.agent.uniqueHostId="+agentNodeName,
		"-Dappdynamics.jvm.shutdown.mark.node.as.historical="+agentNodeMarkAsHistorical)

	appDynamicsAnalyticsAgentURL, exists := context.Environment("APPDYNAMICS_ANALYTICS_AGENT_URL")
	if exists {
		args = append(args, "-Dappdynamics.analytics.agent.url="+appDynamicsAnalyticsAgentURL)
	}

	appHomeDir, exists := context.Environment("HOME")
	if exists {
		args = append(args, "-Dappdynamics.agent.logs.dir="+appHomeDir+"/logs/appdynamics")
	}

	args = append(args, context.Arguments...)

	return args
}

var memoryArguments = []string{"-Xmx", "-XX:+UseCGroupMemoryLimitForHeap", "-Xms"}

type java8MemoryOptions struct {
}

func (m *java8MemoryOptions) shouldModifyArguments(context ArgumentsContext) bool {
	return !containsArgument(context.Arguments, memoryArguments...)
}

func (m *java8MemoryOptions) modifyArguments(context ArgumentsContext) []string {
	args := removeArguments(context.Arguments, memoryArguments)
	memRatio, exists := context.Environment("JAVA_MAX_MEM_RATIO")
	var fraction int
	if exists {
		ratioInPercent, err := strconv.Atoi(memRatio)
		if err != nil {
			logrus.Warnf("Trying to parse JAVA_MAX_MEM_RATIO, but could not parse it %s", err)
		} else {
			fraction = 100 / ratioInPercent
		}
	} else {
		fraction = 4
	}
	limits := context.CGroupLimits
	if limits.HasMemoryLimit() {
		args = append([]string{fmt.Sprintf("-Xmx%dm", limits.MemoryFractionInMB(fraction))}, args...)
		args = append([]string{fmt.Sprintf("-Xms%dm", limits.MemoryFractionInMB(fraction))}, args...)
	}
	return args
}

type java11PlusMemoryOptions struct {
}

func (m *java11PlusMemoryOptions) shouldModifyArguments(context ArgumentsContext) bool {
	return !containsArgument(context.Arguments, memoryArguments...)
}

func (m *java11PlusMemoryOptions) modifyArguments(context ArgumentsContext) []string {
	args := removeArguments(context.Arguments, memoryArguments)
	maxMemory, exists := context.Environment("JAVA_MAX_RAM_PERCENTAGE")
	var percent float64
	if exists {
		memInPercent, err := strconv.ParseFloat(maxMemory, 32)
		if err != nil {
			logrus.Warnf("Trying to parse JAVA_MAX_RAM_PERCENTAGE, but could not parse it %s. Defaults to 75 percent", err)
			percent = 75.0
		} else {
			percent = memInPercent
		}
	} else {
		//Recomended heap size
		percent = 75.0
	}
	limits := context.CGroupLimits
	if limits.HasMemoryLimit() {
		args = append([]string{fmt.Sprintf("-XX:MaxRAMPercentage=%.1f", percent)}, args...)
	}
	return args
}

var metaspaceArguments = []string{"-XX:MaxMetaspaceSize"}

type metaspaceOptions struct {
}

func (m *metaspaceOptions) shouldModifyArguments(context ArgumentsContext) bool {
	if containsArgument(context.Arguments, metaspaceArguments...) {
		return false
	}
	_, exists := context.Environment("JAVA_MAX_METASPACE_RATIO")
	return exists
}

func (m *metaspaceOptions) modifyArguments(context ArgumentsContext) []string {
	args := removeArguments(context.Arguments, metaspaceArguments)
	memRatio, exists := context.Environment("JAVA_MAX_METASPACE_RATIO")
	var fraction int
	if exists {
		ratioInPercent, err := strconv.Atoi(memRatio)
		if err != nil {
			logrus.Warnf("Trying to parse JAVA_MAX_METASPACE_RATIO, but could not parse it %s", err)
		} else {
			fraction = 100 / ratioInPercent
		}
	} else {
		//Fraction of 7 is approx 15% as in old scripts
		fraction = 7
	}
	limits := context.CGroupLimits
	if limits.HasMemoryLimit() {
		args = append([]string{fmt.Sprintf("-XX:MaxMetaspaceSize=%dm", limits.MemoryFractionInMB(fraction))}, args...)
	}
	return args
}

var heapDumpPathArgs = []string{"-XX:HeapDumpPath"}
var heapDumpOnArgs = []string{"-XX:+HeapDumpOnOutOfMemoryError"}

type heapDumpOptions struct {
}

func (m *heapDumpOptions) shouldModifyArguments(context ArgumentsContext) bool {
	isHeapDumpPath := containsArgument(context.Arguments, heapDumpPathArgs...)
	isHeapDumpOn := containsArgument(context.Arguments, heapDumpOnArgs...)

	if isHeapDumpPath && isHeapDumpOn {
		return false
	}
	return true
}

func (m *heapDumpOptions) modifyArguments(context ArgumentsContext) []string {
	args := make([]string, 0)

	if !containsArgument(context.Arguments, heapDumpPathArgs...) {
		javaHeapDumpPath, exists := context.Environment("JAVA_HEAP_DUMP_PATH")
		if exists {
			args = append([]string{fmt.Sprintf("-XX:HeapDumpPath=%s", javaHeapDumpPath)})
		} else {
			args = append([]string{"-XX:HeapDumpPath=/tmp"})
		}
	}

	if !containsArgument(context.Arguments, heapDumpOnArgs...) {
		javaHeapDumpOnOutOfMemoryError, exists := context.Environment("JAVA_HEAP_DUMP_ON_OUT_OF_MEMORY_ERROR")
		if !exists || exists && strings.ToUpper(javaHeapDumpOnOutOfMemoryError) != "FALSE" {
			args = append(args, "-XX:+HeapDumpOnOutOfMemoryError")
		}
	}

	args = append(args, context.Arguments...)
	return args
}

func removeArguments(arguments []string, argumentsToRemove []string) []string {
	ret := make([]string, 0)
outer:
	for _, arg := range arguments {
		for _, candidate := range argumentsToRemove {
			if strings.HasPrefix(arg, candidate) {
				continue outer
			}
		}
		ret = append(ret, arg)
	}
	return ret
}

func containsArgument(arguments []string, argument ...string) bool {
	for _, arg := range arguments {
		for _, candidate := range argument {
			if strings.HasPrefix(arg, candidate) {
				return true
			}
		}
	}
	return false
}

func applyArguments(modificators []ArgumentModificator, ctx ArgumentsContext) []string {
	for _, mod := range modificators {
		if mod.shouldModifyArguments(ctx) {
			logrus.Debugf("Arguments before modificator %s is %+v", reflect.TypeOf(mod), ctx.Arguments)
			args := mod.modifyArguments(ctx)
			ctx.Arguments = args
			logrus.Debugf("Arguments after modificator %s is %+v", reflect.TypeOf(mod), ctx.Arguments)
		}
	}
	return ctx.Arguments
}
