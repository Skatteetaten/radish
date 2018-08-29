package executor

import (
	"encoding/json"
	"github.com/skatteetaten/radish/pkg/util"
	"io"
	"os"
	"path"
	"strings"

	"github.com/drone/envsubst"
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

type Type struct {
	Type    string `json:"Type"`
	Version string `json:"Version"`
}

type JavaDescriptorData struct {
	Basedir               string   `json:"Basedir"`
	PathsToClassLibraries []string `json:"PathsToClassLibraries"`
	MainClass             string   `json:"MainClass"`
	ApplicationArgs       string   `json:"ApplicationArgs"`
	JavaOptions           string   `json:"JavaOptions"`
}

type JavaDescriptor struct {
	Type
	Data JavaDescriptorData
}

func buildArgline(descriptor JavaDescriptor, env func(string) (string, bool), cgl util.CGroupLimits) ([]string, error) {
	args := make([]string, 0, 10)
	classpath, err := createClasspath(descriptor.Data.Basedir, descriptor.Data.PathsToClassLibraries)
	if err != nil {
		return nil, err
	} else if len(classpath) == 0 {
		logrus.Warn("No classpath found... Probably a configuration issue?")
	} else {
		args = append(args, "-cp", strings.Join(classpath, ":"))
	}
	args = applyArguments(ArgumentsModificators, ArgumentsContext{
		Arguments:    args,
		Environment:  env,
		CGroupLimits: cgl,
		Descriptor:   descriptor,
	})
	args = append(args, descriptor.Data.MainClass)
	if len(strings.TrimSpace(descriptor.Data.ApplicationArgs)) != 0 {
		args = append(args, descriptor.Data.ApplicationArgs)
	}

	return expandArgumentsAgainstEnv(args, env), nil
}

func expandArgumentsAgainstEnv(args []string, env func(string) (string, bool)) []string {
	argsAfterSubstitution := make([]string, len(args), len(args))
	for i, arg := range args {
		substituted, err := envsubst.Eval(arg, func(key string) string {
			value, _ := env(key)
			return value
		})
		if err != nil {
			logrus.Warnf("Error substituting in arg %s", arg)
			argsAfterSubstitution[i] = arg
		} else {
			argsAfterSubstitution[i] = substituted
		}

	}
	return argsAfterSubstitution
}

func createClasspath(basedir string, patterns []string) ([]string, error) {
	cp := make([]string, 0, 10)
	for _, pattern := range patterns {
		p := path.Join(basedir, pattern)
		fi, err := os.Stat(p)
		if os.IsNotExist(err) {
			logrus.Debugf("Trying to build classpath from %s but it does not exist", p)
			continue
		}
		if err != nil {
			logrus.Warnf("Trying to build classpath from %s but it was an error", p, err)
			continue
		}
		if fi.IsDir() {
			files, err := ioutil.ReadDir(p)
			if err != nil {
				logrus.Warn("Can not list content of directory %s", p)
			}
			for _, file := range files {
				if file.Mode().IsRegular() {
					cp = append(cp, path.Join(p, file.Name()))
				}
			}
		} else if fi.Mode().IsRegular() {
			cp = append(cp, p)
		}
	}
	return cp, nil
}

func unmarshallDescriptor(buffer io.Reader) (JavaDescriptor, error) {
	var data JavaDescriptor
	err := json.NewDecoder(buffer).Decode(&data)
	return data, err
}
