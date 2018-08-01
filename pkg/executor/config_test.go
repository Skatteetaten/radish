package executor

import (
	"bytes"
	"github.com/skatteetaten/radish/pkg/util"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func envFunc(key string) (string, bool) {
	env := make(map[string]string)
	env["DISABLE_JOLOKIA"] = "true"
	k, e := env[key]
	return k, e
}

func TestBuildArgline(t *testing.T) {
	dat, err := ioutil.ReadFile("testdata/testconfig.json")
	assert.NoError(t, err)
	desc, err := unmarshallDescriptor(bytes.NewBuffer(dat))
	assert.NoError(t, err)
	args, err := buildArgline(desc, envFunc, util.CGroupLimits{
		MaxCoresEstimated:  2,
		MemoryLimitInBytes: 2 * 1024 * 1024 * 1024,
	})
	assert.NoError(t, err)
	assert.Contains(t, args, "-cp")
	assert.Contains(t, args, "testdata/lib/lib1.jar:testdata/lib/lib2.jar:testdata/lib/lib2/lib4.jar")
}
