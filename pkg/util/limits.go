package util

import (
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"
)

//https://www.kernel.org/doc/Documentation/cgroup-v1/memory.txt
//https://www.kernel.org/doc/Documentation/scheduler/sched-bwc.txt
const (
	cfsPeriodUs            = "/sys/fs/cgroup/cpu/cpu.cfs_period_us"
	cfsQuotaUs             = "/sys/fs/cgroup/cpu/cpu.cfs_quota_us"
	memAndSwapLimitInBytes = "/sys/fs/cgroup/memory/memory.memsw.limit_in_bytes"
	memLimitInBytes        = "/sys/fs/cgroup/memory/memory.limit_in_bytes"
)

//CGroupLimits :
type CGroupLimits struct {
	MaxCoresEstimated  int
	MemoryLimitInBytes int
}

//ReadCGroupLimits :
func ReadCGroupLimits() CGroupLimits {

	ret := CGroupLimits{}
	periodUs := readCGroupLimit(cfsPeriodUs)
	quotaUs := readCGroupLimit(cfsQuotaUs)
	memLimitInBytes := readCGroupLimit(memLimitInBytes)
	memAndSwapLimitInBytes := readCGroupLimit(memAndSwapLimitInBytes)

	if periodUs == -1 || quotaUs == -1 {
		ret.MaxCoresEstimated = -1
	} else {
		ret.MaxCoresEstimated = int(math.Ceil(float64(periodUs) / float64(quotaUs)))
	}

	if memLimitInBytes > memAndSwapLimitInBytes {
		ret.MemoryLimitInBytes = memLimitInBytes
	} else {
		ret.MemoryLimitInBytes = memAndSwapLimitInBytes
	}

	return ret
}

func readCGroupLimit(cgroupFilePath string) int {
	dat, err := ioutil.ReadFile(cgroupFilePath)
	if os.IsNotExist(err) {
		logrus.Debugf("File %s does not exist", cgroupFilePath)
		return -1
	}
	if err != nil {
		logrus.Errorf("Could not read %s because of: %s, defaulting to -1", cgroupFilePath, err)
		return -1
	}

	parsed, err := strconv.Atoi(strings.TrimSpace(string(dat)))
	if err != nil {
		logrus.Errorf("Could not parse %s because of: %s defaulting to -1", dat, err)
		return -1
	}
	return parsed
}

//HasMemoryLimit :
func (e CGroupLimits) HasMemoryLimit() bool {
	return e.MemoryLimitInBytes > 0
}

//HasCoreLimit :
func (e CGroupLimits) HasCoreLimit() bool {
	return e.MaxCoresEstimated > 0
}

//MemoryFractionInMB :
func (e CGroupLimits) MemoryFractionInMB(fraction int) interface{} {
	//If memory limits is above a sensible default (64G), we set it to 1G
	if e.MemoryLimitInBytes > 2<<37 {
		return 1024
	}
	return e.MemoryLimitInBytes / (1024 * 1024 * fraction)
}
