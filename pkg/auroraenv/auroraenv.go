package auroraenv

import (
	"fmt"
	"github.com/pkg/errors"
	"os"
	"regexp"
	"strings"

	"bytes"
	"io"
	"path"

	"github.com/magiconair/properties"
	"github.com/plaid/go-envvar/envvar"
	"github.com/sirupsen/logrus"
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

//GenerateEnvScript :
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
	//appVersion example: 1.2.0
	//configLocation example: /u01/config/secrets
	var versions []string
	// If match we have a semantic version with minor and patch with optional meta
	if isFullSemanticVersion(vars.AppVersion) {
		appVersion := getVersionOnly(vars.AppVersion)
		if appVersion != vars.AppVersion {
			logrus.Infof("Only using version info ^d+.d+.d+ from version %s for config files check", vars.AppVersion)
		}
		splitVersion := strings.Split(appVersion, ".")
		majorVersion := splitVersion[0]
		minorVersion := splitVersion[0] + "." + splitVersion[1]
		versions = []string{appVersion, minorVersion, majorVersion}
	} else {
		logrus.Infof("No valid version in form ^d+.d+.d+ found in version %s. Using latest prefix for config file check", vars.AppVersion)
	}
	versions = append(versions, "latest")
	logrus.Infof("Looking for config files in version order prefix: %s", versions)

	buffer := &bytes.Buffer{}
	for _, dir := range configDirs {
		path := path.Join(dir.basedir, dir.dir)
		logrus.Debugf("Processing dir: %s", path)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			logrus.Infof("No configdir %s", path)
			continue
		}
		configVersion, err := findConfigVersion(versions, path)
		if err != nil {
			logrus.Debug("Error reading config")
			return "", errors.Wrap(err, "Error reading config")
		} else if configVersion == "" {
			logrus.Infof("No config in %s", dir.dir)
			continue
		}
		err = exportPropertiesAsEnvVars(buffer, path+"/"+configVersion+".properties", dir.shouldMask)
		if err != nil {
			logrus.Debugf("Returning with error after export: %s", err.Error())
			return "", err
		}
	}
	result := buffer.String()
	return result, nil

}

func findConfigVersion(versions []string, configLocation string) (string, error) {
	for _, version := range versions {
		if _, err := os.Stat(configLocation + "/" + version + ".properties"); err == nil {
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
	envCounter := 0
	for _, key := range p.Keys() {
		if isValidEnvironmentVariable(key) {
			val := p.MustGetString(key)
			fmt.Fprintf(writer, "export %s=%s\n", key, val)
			if maskValue {
				logrus.Debugf("export %s=******", key)
			} else {
				logrus.Debugf("export %s=%s", key, val)
			}
			envCounter++
		} else {
			logrus.Warnf("Variable %s does not validate and will not be exported", key)
		}
	}
	if envCounter > 0 {
		logrus.Infof("Exported %d environment variables from %s", envCounter, filepath)
	}
	return nil
}

var validEnvironmentVariable = regexp.MustCompile(`^[_[:alpha:]][_[:alpha:][:digit:]]*$`)

func isValidEnvironmentVariable(envVar string) bool {
	return validEnvironmentVariable.MatchString(envVar)
}

var versionWithMinorAndPatch = regexp.MustCompile(`^([0-9]+\.[0-9]+\.[0-9]+$|^[0-9]+\.[0-9]+\.[0-9]+).*`)

func isFullSemanticVersion(versionString string) bool {
	return versionWithMinorAndPatch.MatchString(versionString)
}

func getVersionOnly(versionString string) string {
	matches := versionWithMinorAndPatch.FindStringSubmatch(versionString)
	if matches == nil {
		return versionString
	}
	return matches[1]
}
