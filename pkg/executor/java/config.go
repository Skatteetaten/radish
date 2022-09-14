package java

import (
	"encoding/json"
	"github.com/kballard/go-shellquote"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/skatteetaten/radish/pkg/util"

	"io/ioutil"

	"github.com/drone/envsubst"
	"github.com/sirupsen/logrus"
)

type descriptorData struct {
	Basedir               string
	PathsToClassLibraries []string
	MainClass             string
	ApplicationArgs       string
	JavaOptions           string
	StartScript           string
}

type descriptor struct {
	Type string
	Data descriptorData
}

func buildArgline(desc descriptor, env func(string) (string, bool),
	argumentModificators []ArgumentModificator, cgl util.CGroupLimits) ([]string, error) {
	args := make([]string, 0, 10)
	classpath, err := createClasspath(desc.Data.Basedir, desc.Data.PathsToClassLibraries)
	if err != nil {
		return nil, err
	} else if len(classpath) == 0 {
		logrus.Warn("No classpath found... Probably a configuration issue?")
	} else {
		args = append(args, "-cp", strings.Join(classpath, ":"))
	}
	args = applyArguments(argumentModificators, ArgumentsContext{
		Arguments:    args,
		Environment:  env,
		CGroupLimits: cgl,
		Descriptor:   desc,
	})
	args = append(args, desc.Data.MainClass)
	if len(strings.TrimSpace(desc.Data.ApplicationArgs)) != 0 {
		splittedArgs, err := shellquote.Split(desc.Data.ApplicationArgs)
		if err == nil {
			args = append(args, splittedArgs...)
		} else {
			logrus.Warnf("Error parsing args: %s", err)
			args = append(args, desc.Data.ApplicationArgs)
		}
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
		if strings.HasSuffix(p, "/**") {
			wcp, err := walkClasspath(strings.TrimSuffix(p, "/**"))
			if err != nil {
				logrus.Warnf("Can not walk directory %s", p)
				continue
			}
			cp = append(cp, wcp...)
		} else {
			fi, err := os.Stat(p)
			if os.IsNotExist(err) {
				logrus.Debugf("Trying to build classpath from %s but it does not exist.", p)
				continue
			}
			if err != nil {
				logrus.Warnf("Trying to build classpath from %s but it was an error: %s", p, err)
				continue
			}
			if fi.IsDir() {
				files, err := ioutil.ReadDir(p)
				if err != nil {
					logrus.Warnf("Can not list content of directory %s", p)
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
	}
	return cp, nil
}

func walkClasspath(path string) ([]string, error) {
	var jarfiles []string

	err := filepath.Walk(path,
		func(subpath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(info.Name(), ".jar") {
				jarfiles = append(jarfiles, subpath)
			}
			return nil
		})
	return jarfiles, err
}

func unmarshallDescriptor(buffer io.Reader) (descriptor, error) {
	var data descriptor
	err := json.NewDecoder(buffer).Decode(&data)
	return data, err
}
