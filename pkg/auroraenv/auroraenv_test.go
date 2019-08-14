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
validkey1=value1
1notvalidkey=value1
1not_valid_key=value1
not.valid.key=value1
validkey2=value2
valid_key_3=value3
`), 0644)

	expectedEnvScript := `export validkey1=value1
export validkey2=value2
export valid_key_3=value3
`
	envscript, err := GenerateEnvScript()
	assert.NoError(t, err)
	assert.Equal(t, envscript, expectedEnvScript)
}

func TestFindConfigVersion(t *testing.T) {
	appVersion := "1.2.0"
	splitVersion := strings.Split(appVersion, ".")
	var versions []string
	versions = []string{appVersion, splitVersion[0] + "." + splitVersion[1], splitVersion[0], "latest"}
	testdir, err := ioutil.TempDir("", "radish")
	assert.NoError(t, err)
	defer os.RemoveAll(testdir)
	configLocation := testdir

	filepath := configLocation + "/" + appVersion + ".properties"
	ioutil.WriteFile(filepath, []byte("test text"), 0644)

	version, err := findConfigVersion(versions, configLocation)
	assert.NoError(t, err)

	assert.True(t, strings.HasPrefix(version, appVersion))

}
