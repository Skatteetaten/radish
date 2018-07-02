package auroraenv

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/skatteetaten/radish/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestSetAuroraEnv(t *testing.T) {
	os.Setenv("HOME", "envtest")
	os.Setenv("AURORA_VERSION", "1.2.0-b1.4.3-flange-8.152.18")
	os.Setenv("APP_VERSION", "1.2.0")

	configpath := "envtest/config"
	secretspath := configpath + "/secrets"
	util.CreateDirIfNotExist(secretspath)
	filepath := secretspath + "/1.2.properties"
	ioutil.WriteFile(filepath, []byte(`
key1=value1
key2=val2
`), 0644)

	success, err := SetAuroraEnv()
	assert.NoError(t, err)
	assert.True(t, success)

	//cleanup file
	os.RemoveAll("envtest")
}

func TestFindConfigVersion(t *testing.T) {
	auroraVersion := "1.2.0-b1.4.3-flange-8.152.18"
	appVersion := "1.2.0"
	configLocation := "test"

	util.CreateDirIfNotExist(configLocation)
	filepath := configLocation + "/" + appVersion + ".properties"
	ioutil.WriteFile(filepath, []byte("test text"), 0644)

	version, err := findConfigVersion(auroraVersion, appVersion, configLocation)
	assert.NoError(t, err)

	assert.True(t, strings.HasPrefix(version, appVersion))

	//cleanup file
	os.Remove(filepath)

}

func TestExportPropertiesAsEnvVars(t *testing.T) {
	util.CreateDirIfNotExist("test_data")
	filepath := "test_data/test.properties"
	ioutil.WriteFile(filepath, []byte(`
key1=value1
key2=val2
`), 0644)

	output := captureOutputFromFunction(exportPropertiesAsEnvVars, "test_data/test.properties")

	expected := `export key1=value1
export key2=val2
`
	assert.Equal(t, output, expected)

	os.RemoveAll("test_data")
}

func captureOutputFromFunction(f func(param string) (bool, error), param string) string {
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f(param)
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout

	fmt.Printf("Captured: %s", out)

	return string(out)
}
