package executor

import (
	"encoding/json"
	"github.com/skatteetaten/radish/pkg/util"
	"io"
	"os"
	"path"
	"strings"

	"github.com/kballard/go-shellquote"
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
	if len(descriptor.Data.JavaOptions) != 0 {
		splittedArgs, err := shellquote.Split(descriptor.Data.JavaOptions)
		if err != nil {
			logrus.Error("Unable to parse args from radish descriptor: %s %s", descriptor.Data.JavaOptions, err)
		}
		args = append(args, splittedArgs...)
	}
	args = applyArguments(ArgumentsModificators, ArgumentsContext{
		Arguments:    args,
		Environment:  env,
		CGroupLimits: cgl,
	})
	args = append(args, descriptor.Data.MainClass)
	if len(strings.TrimSpace(descriptor.Data.ApplicationArgs)) != 0 {
		args = append(args, descriptor.Data.ApplicationArgs)
	}
	return args, nil
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
