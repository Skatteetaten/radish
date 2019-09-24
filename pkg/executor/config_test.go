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
	env["ENABLE_JOLOKIA"] = "true"
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
	args, err := buildArgline(desc, envFunc, Java8ArgumentsModificators, limits)
	assert.NoError(t, err)
	assert.Contains(t, args, "testdata/lib/lib1.jar:testdata/lib/lib2.jar:testdata/lib/lib2/lib4.jar")
}

func TestBuildArgLineFromDescriptorSubpath(t *testing.T) {
	dat, err := ioutil.ReadFile("testdata/testconfig-subpath.json")
	assert.NoError(t, err)
	desc, err := unmarshallDescriptor(bytes.NewBuffer(dat))
	limits := util.CGroupLimits{
		MaxCoresEstimated:  2,
		MemoryLimitInBytes: 2 * 1024 * 1024 * 1024,
	}
	assert.NoError(t, err)
	args, err := buildArgline(desc, envFunc, Java8ArgumentsModificators, limits)
	assert.NoError(t, err)
	assert.Contains(t, args, "testdata/lib/lib1.jar:testdata/lib/lib2.jar:testdata/lib/lib2/lib4.jar:testdata/lib/lib3/lib4/lib6.jar:testdata/lib/lib3/lib5/lib7.jar:testdata/lib/lib3/lib5/lib8.jar:testdata/lib/lib3/lib5.jar")
}

func TestExpanstionOfVariablesAgainstEnv(t *testing.T) {
	desc := JavaDescriptor{
		Data: JavaDescriptorData{
			JavaOptions: "-Dtest=${SOME} -Dtest2=${OTHER} \"this should not be splitt\"",
		},
	}
	args, err := buildArgline(desc, envFunc, Java8ArgumentsModificators, util.ReadCGroupLimits())
	assert.NoError(t, err)
	assert.Contains(t, args, "-Dtest=Jallaball")
	assert.Contains(t, args, "-Dtest2=Ballejall")
	assert.Contains(t, args, "this should not be splitt")
}

func TestTokenizationOfShellQuotedArgs(t *testing.T) {
	desc := JavaDescriptor{
		Data: JavaDescriptorData{
			ApplicationArgs: "arg1 arg2 \"this should not be splitt\"",
			MainClass:       "Class",
		},
	}
	args, err := buildArgline(desc, envFunc, []ArgumentModificator{}, util.ReadCGroupLimits())
	assert.NoError(t, err)
	assert.Len(t, args, 4)
	assert.Equal(t, args[0], "Class")
	assert.Equal(t, args[1], "arg1")
	assert.Equal(t, args[2], "arg2")
	assert.Equal(t, args[3], "this should not be splitt")

	desc.Data.ApplicationArgs = "arg1"
	args, err = buildArgline(desc, envFunc, []ArgumentModificator{}, util.ReadCGroupLimits())
	assert.NoError(t, err)
	assert.Len(t, args, 2)
	assert.Equal(t, args[0], "Class")
	assert.Equal(t, args[1], "arg1")
}
