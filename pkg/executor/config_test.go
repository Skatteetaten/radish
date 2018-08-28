package executor

import (
	"bytes"
	"github.com/skatteetaten/radish/pkg/util"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

var env = make(map[string]string)

func envFunc(key string) (string, bool) {
	env["DISABLE_JOLOKIA"] = "true"
	env["SOME"] = "Jallaball"
	env["OTHER"] = "Ballejall"
	k, e := env[key]
	return k, e
}

func TestBuildArgLineFromDescriptor(t *testing.T) {
	dat, err := ioutil.ReadFile("testdata/testconfig.json")
	assert.NoError(t, err)
	desc, err := unmarshallDescriptor(bytes.NewBuffer(dat))
	limits := util.CGroupLimits{
		MaxCoresEstimated:  2,
		MemoryLimitInBytes: 2 * 1024 * 1024 * 1024,
	}
	assert.NoError(t, err)
	args, err := buildArgline(desc, envFunc, limits)
	assert.NoError(t, err)
	assert.Contains(t, args, "testdata/lib/lib1.jar:testdata/lib/lib2.jar:testdata/lib/lib2/lib4.jar")
	assert.Contains(t, args, "testdata/lib/lib1.jar:testdata/lib/lib2.jar:testdata/lib/lib2/lib4.jar")
}

func TestExpanstionOfVariablesAgainstEnv(t *testing.T) {
	desc := JavaDescriptor{
		Data: JavaDescriptorData{
			JavaOptions: "-Dtest=${SOME} -Dtest2=${OTHER} \"this should not be splitt\"",
		},
	}
	args, err := buildArgline(desc, envFunc, util.ReadCGroupLimits())
	assert.NoError(t, err)
	assert.Contains(t, args, "-Dtest=Jallaball")
	assert.Contains(t, args, "-Dtest2=Ballejall")
	assert.Contains(t, args, "this should not be splitt")
}
