package auroraenv

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"path"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSetAuroraEnv(t *testing.T) {
	os.Setenv("AURORA_VERSION", "1.2.0-b1.4.3-flange-8.152.18")
	os.Setenv("APP_VERSION", "1.2.0")
	logrus.SetLevel(logrus.DebugLevel)
	testdir, err := ioutil.TempDir("", "radish")
	os.Setenv("HOME", testdir)
	secretPath := path.Join(testdir, "config/secrets")
	os.MkdirAll(secretPath, 0755)
	defer os.RemoveAll(testdir)
	filepath := path.Join(secretPath, "/1.2.properties")
	ioutil.WriteFile(filepath, []byte(`
key1=value1
key2=val2
`), 0644)

	expectedEnvScript := `export key1=value1
export key2=val2
`
	envscript, err := GenerateEnvScript()
	assert.NoError(t, err)
	assert.Equal(t, envscript, expectedEnvScript)
}

func TestFindConfigVersion(t *testing.T) {
	appVersion := "1.2.0"
	testdir, err := ioutil.TempDir("", "radish")
	assert.NoError(t, err)
	defer os.RemoveAll(testdir)
	configLocation := testdir

	filepath := configLocation + "/" + appVersion + ".properties"
	ioutil.WriteFile(filepath, []byte("test text"), 0644)

	version, err := findConfigVersion(appVersion, configLocation)
	assert.NoError(t, err)

	assert.True(t, strings.HasPrefix(version, appVersion))

}
