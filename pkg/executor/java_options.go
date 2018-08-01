package executor

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/skatteetaten/radish/pkg/util"
	"reflect"
	"strconv"
	"strings"
)

type ArgumentsContext struct {
	Arguments    []string
	Environment  func(string) (string, bool)
	CGroupLimits util.CGroupLimits
}

type ArgumentsDeriver interface {
	shouldDeriveArguments(context ArgumentsContext) bool
	deriveArguments(context ArgumentsContext) []string
}

var ArgumentsModificators = []ArgumentsDeriver{
	&debugOptions{},
	&diagnosticsOptions{},
	&jolokiaOptions{},
	&cpuCoreTuning{},
	&memoryOptions{},
}

type diagnosticsOptions struct {
}

func (m *diagnosticsOptions) shouldDeriveArguments(context ArgumentsContext) bool {
	value, exists := context.Environment("ENABLE_DIAGNOSTICS")
	return exists && strings.ToUpper(value) == "TRUE"
}

func (m *diagnosticsOptions) deriveArguments(context ArgumentsContext) []string {
	args := make([]string, 0, 3)
	args = append(args, "-XX:NativeMemoryTracking=summary",
		"-XX:+PrintGC",
		"-XX:+PrintGCDateStamps",
		"-XX:+PrintGCTimeStamps",
		"-XX:+UnlockDiagnosticVMOptions")
	args = append(args, context.Arguments...)
	return args
}

type debugOptions struct {
}

func (m *debugOptions) shouldDeriveArguments(context ArgumentsContext) bool {
	value, exists := context.Environment("ENABLE_REMOTE_DEBUG")
	return exists && strings.ToUpper(value) == "TRUE"
}

func (m *debugOptions) deriveArguments(context ArgumentsContext) []string {
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

var cpuCoreArguments = []string{"-XX:ParallelGCThreads",
	"-XX:ConcGCThreads",
	"-Djava.util.concurrent.ForkJoinPool.common.parallelism"}

type cpuCoreTuning struct {
}

func (m *cpuCoreTuning) shouldDeriveArguments(context ArgumentsContext) bool {
	return !containsArgument(context.Arguments, cpuCoreArguments...) && context.CGroupLimits.HasCoreLimit()
}

func (m *cpuCoreTuning) deriveArguments(context ArgumentsContext) []string {
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

func (m *jolokiaOptions) shouldDeriveArguments(context ArgumentsContext) bool {
	value, exists := context.Environment("DISABLE_JOLOKIA")
	return !exists || !(strings.ToUpper(value) == "TRUE")
}

func (m *jolokiaOptions) deriveArguments(context ArgumentsContext) []string {
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

var memoryArguments = []string{"-Xmx", "-XX:+UseCGroupMemoryLimitForHeap", "-Xms"}

type memoryOptions struct {
}

func (m *memoryOptions) shouldDeriveArguments(context ArgumentsContext) bool {
	return !containsArgument(context.Arguments, memoryArguments...)
}

func (m *memoryOptions) deriveArguments(context ArgumentsContext) []string {
	args := removeArguments(context.Arguments, memoryArguments)
	limits := context.CGroupLimits
	if limits.HasMemoryLimit() {
		args = append([]string{fmt.Sprintf("-Xmx%dm", limits.MemoryFractionInMB(4))}, args...)
		args = append([]string{fmt.Sprintf("-Xms%dm", limits.MemoryFractionInMB(4))}, args...)
	}
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

func applyArguments(modificators []ArgumentsDeriver, ctx ArgumentsContext) []string {
	for _, mod := range modificators {
		if mod.shouldDeriveArguments(ctx) {
			logrus.Debugf("Arguments before modificator %s is %+v", reflect.TypeOf(mod), ctx.Arguments)
			args := mod.deriveArguments(ctx)
			ctx.Arguments = args
			logrus.Debugf("Arguments after modificator %s is %+v", reflect.TypeOf(mod), ctx.Arguments)
		}
	}
	return ctx.Arguments
}
