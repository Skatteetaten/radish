package auroraenv

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"

	"bytes"
	"github.com/magiconair/properties"
	"github.com/plaid/go-envvar/envvar"
	"github.com/sirupsen/logrus"
	"io"
	"path"
)

// local SYMLINK_FOLDER=$1    //=$HOME
// local CONFIG_BASE_DIR=$2   //=$HOME/config
// local COMPLETE_VERSION=$3  //=$AURORA_VERSION
// local APP_VERSION=$4       //=$APP_VERSION

//EnvData : Struct for the required elements in the configuration json
type EnvData struct {
	HomeFolder    string `envvar:"HOME"`
	AuroraVersion string `envvar:"AURORA_VERSION"`
	AppVersion    string `envvar:"APP_VERSION"`
}

//generateShellScript :
func GenerateEnvScript() (string, error) {
	vars := EnvData{}
	if err := envvar.Parse(&vars); err != nil {
		logrus.Fatal(err)
		return "", errors.Wrap(err, "Error parsing environment vars")
	}

	configBaseDir := vars.HomeFolder + "/config"

	type configFile struct {
		shouldMask bool
		basedir    string
		dir        string
	}

	configDirs := []configFile{
		{
			shouldMask: true,
			basedir:    configBaseDir,
			dir:        "secrets",
		}, {
			shouldMask: true,
			basedir:    configBaseDir,
			dir:        "secret",
		}, {
			shouldMask: false,
			basedir:    configBaseDir,
			dir:        "configmaps",
		}, {
			shouldMask: false,
			basedir:    configBaseDir,
			dir:        "configmap",
		},
	}

	buffer := &bytes.Buffer{}

	for _, dir := range configDirs {
		path := path.Join(dir.basedir, dir.dir)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			logrus.Debugf("No configdir in %s", path)
			continue
		}
		configVersion, err := findConfigVersion(vars.AppVersion, path)
		if err != nil {
			return "", errors.Wrap(err, "Error reading config")
		} else if configVersion == "" {
			logrus.Infof("No config in %s", dir)
			continue
		}
		err = exportPropertiesAsEnvVars(buffer, path+"/"+configVersion+".properties", dir.shouldMask)
		if err != nil {
			return "", err
		}
	}
	return buffer.String(), nil

}

func findConfigVersion(appVersion string, configLocation string) (string, error) {
	//appVersion example: 1.2.0
	//configLocation example: /u01/config/secrets
	var versions []string
	if appVersion == "" {
		logrus.Info("App version is empty. Only look for latest.properties")
		versions = []string{"latest"}
	} else {
		splitVersion := strings.Split(appVersion, ".")
		majorVersion := splitVersion[0]
		minorVersion := splitVersion[0] + "." + splitVersion[1]
		versions = []string{appVersion, minorVersion, majorVersion, "latest"}
	}
	logrus.Debugf("Looking for files in order: ", versions)
	for _, version := range versions {
		if _, err := os.Stat(configLocation + "/" + version + ".properties"); err == nil {
			logrus.Debugf("Using version %s", version)
			return version, nil
		} else if !os.IsNotExist(err) {
			return "", errors.Wrap(err, "Error finding configfile")
		}
	}
	return "", nil
}

func exportPropertiesAsEnvVars(writer io.Writer, filepath string, maskValue bool) error {
	logrus.Debugf("Reading file %s", filepath)
	p, err := properties.LoadFile(filepath, properties.UTF8)
	if err != nil {
		return errors.Wrap(err, "Error reading properties file")
	}
	for _, key := range p.Keys() {
		val := p.MustGetString(key)
		fmt.Fprintf(writer, "export %s=%s\n", key, val)
		if maskValue {
			logrus.Debugf("export %s=******\n", key)
		} else {
			logrus.Debugf("export %s=%s\n", key, val)
		}
		//TODO need to handle panic? can't I think.. must check conditions before calling if so
	}
	return nil
}
