package auroraenv

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"

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

//SetAuroraEnv :
func SetAuroraEnv() (bool, error) {
	vars := EnvData{}
	if err := envvar.Parse(&vars); err != nil {
		logrus.Fatal(err)
		return false, errors.Wrap(err, "Error parsing environment vars")
	}

	configBaseDir := vars.HomeFolder + "/config"

	configDirs := []string{configBaseDir + "/secrets", configBaseDir + "/configmaps"}

	for _, dir := range configDirs {
		configVersion, err := findConfigVersion(vars.AuroraVersion, vars.AppVersion, dir)
		if err != nil {
			logrus.Debugf("No config in %d. Err: %s", dir, err)
			continue
		}
		exportPropertiesAsEnvVars(dir + "/" + configVersion + ".properties")
	}

	return true, nil
}

func findConfigVersion(auroraVersion string, appVersion string, configLocation string) (string, error) {
	//auroraVersion example: 1.2.0-b1.4.3-flange-8.152.18
	//appVersion example: 1.2.0
	//configLocation example: /u01/config/secrets

	splitVersion := strings.Split(appVersion, ".")
	majorVersion := splitVersion[0]
	minorVersion := splitVersion[0] + "." + splitVersion[1]

	versions := []string{auroraVersion, appVersion, minorVersion, majorVersion, "latest"}

	for _, version := range versions {
		if _, err := os.Stat(configLocation + "/" + version + ".properties"); err == nil {
			return version, nil
		}
	}
	return "", errors.New("No config mounted for " + configLocation)
}

func exportPropertiesAsEnvVars(filepath string) (bool, error) {
	p := properties.MustLoadFile(filepath, properties.UTF8)
	for _, key := range p.Keys() {
		val := p.MustGetString(key)
		fmt.Printf("export %s=%v\n", key, val)

		//TODO need to handle panic? can't I think.. must check conditions before calling if so
	}
	return true, nil
}
