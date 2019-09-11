package nodejs

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/skatteetaten/radish/pkg/util"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
)

const nginxConfigTemplate string = `
worker_processes  1;
error_log stderr;

events {
    worker_connections  1024;
}


http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;

    log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
                      '$status $body_bytes_sent "$http_referer" '
                      '"$http_user_agent" "$http_x_forwarded_for"';

    access_log  /dev/stdout;

    sendfile        on;
    #tcp_nopush     on;

    keepalive_timeout  65;

    #gzip  on;

    index index.html;

    server {
       listen 8080;

       location /api {
          {{if or .HasNodeJSApplication .ConfigurableProxy}}proxy_pass http://${PROXY_PASS_HOST}:${PROXY_PASS_PORT};{{else}}return 404;{{end}}{{range $key, $value := .NginxOverrides}}
          {{$key}} {{$value}};{{end}}
       }
{{if .SPA}}
       location {{.Path}} {
          root /u01/static;
          try_files $uri {{.Path}}index.html;{{else}}
       location {{.Path}} {
          root /u01/static;{{end}}{{range $key, $value := .ExtraStaticHeaders}}
          add_header {{$key}} "{{$value}}";{{end}}
       }
    }
}
`

//BuildNginx :
func BuildNginx(radishDescriptor string, nginxPath string) {
	dat, err := ioutil.ReadFile(radishDescriptor)
	if err != nil {
		return nil, err
	}

	completeDockerName := baseImage.GetCompleteDockerTagName()
	input, err := mapOpenShiftJSONToTemplateInput(dockerSpec, v, completeDockerName, imageBuildTime, auroraVersion)

	err = util.NewTemplateWriter(input, "NgnixConfiguration", nginxConfigTemplate, nginxPath)
	if err != nil {
		return errors.Wrap(err, "Error creating nginx configuration")
	}
}

func mapOpenShiftJSONToTemplateInput(dockerSpec config.DockerSpec, v *openshiftJson, completeDockerName string, imageBuildTime string, auroraVersion *runtime.AuroraVersion) (*templateInput, error) {
	labels := make(map[string]string)
	if v.DockerMetadata.Labels != nil {
		for k, v := range v.DockerMetadata.Labels {
			labels[k] = v
		}
	}
	labels["version"] = string(auroraVersion.GetAppVersion())
	labels["maintainer"] = findMaintainer(v.DockerMetadata)

	path := "/"
	if v.Aurora.Webapp != nil && len(strings.TrimPrefix(v.Aurora.Webapp.Path, "/")) > 0 {
		path = "/" + strings.TrimPrefix(v.Aurora.Webapp.Path, "/")
	} else if len(strings.TrimPrefix(v.Aurora.Path, "/")) > 0 {
		logrus.Warnf("web.path in openshift.json is deprecated. Please use web.webapp.path when setting path: %s", v.Aurora.Path)
		path = "/" + strings.TrimPrefix(v.Aurora.Path, "/")
	}

	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	var nodejsMainfile string
	var overrides map[string]string
	var err error
	if v.Aurora.NodeJS != nil {
		nodejsMainfile = strings.TrimSpace(v.Aurora.NodeJS.Main)
		overrides = v.Aurora.NodeJS.Overrides
		err = whitelistOverrides(overrides)
		if err != nil {
			return nil, err
		}
	}

	var static string
	var spa bool
	var extraHeaders map[string]string

	if v.Aurora.Webapp == nil {
		static = v.Aurora.Static
		spa = v.Aurora.SPA
		extraHeaders = nil

	} else {
		static = v.Aurora.Webapp.StaticContent
		spa = v.Aurora.Webapp.DisableTryfiles == false
		extraHeaders = v.Aurora.Webapp.Headers
	}

	env := make(map[string]string)
	env["MAIN_JAVASCRIPT_FILE"] = "/u01/application/" + nodejsMainfile
	env["PROXY_PASS_HOST"] = "localhost"
	env["PROXY_PASS_PORT"] = "9090"
	env[docker.IMAGE_BUILD_TIME] = imageBuildTime
	env[docker.ENV_APP_VERSION] = string(auroraVersion.GetAppVersion())
	env[docker.ENV_AURORA_VERSION] = string(auroraVersion.GetCompleteVersion())
	env[docker.ENV_PUSH_EXTRA_TAGS] = dockerSpec.PushExtraTags.ToStringValue()
	if auroraVersion.Snapshot {
		env[docker.ENV_SNAPSHOT_TAG] = auroraVersion.GetGivenVersion()
	}

	return &templateInput{
		Baseimage:            completeDockerName,
		HasNodeJSApplication: len(nodejsMainfile) != 0,
		NginxOverrides:       overrides,
		ConfigurableProxy:    v.Aurora.ConfigurableProxy,
		Static:               static,
		ExtraStaticHeaders:   extraHeaders,
		SPA:                  spa,
		Path:                 path,
		Labels:               labels,
		Env:                  env,
		PackageDirectory:     "package",
	}, nil
}

/*

We sanitize the input.... Don't want to large inputs.

For example; Accepting very large client_max_body_size would make a DOS attack very easy to implement...

*/
var allowedNginxOverrides = map[string]func(string) error{
	"client_max_body_size": func(s string) error {
		// between 1 and 20
		match, err := regexp.MatchString("^([1-9]|[1-4][0-9]|[5][0])m$", s)
		if err != nil {
			return err
		}
		if !match {
			return errors.New("Value on client_max_body_size should be on the form Nm where N is between 1 and 50")
		}
		return nil
	},
}

func whitelistOverrides(overrides map[string]string) error {
	if overrides == nil {
		return nil
	}

	for key, value := range overrides {
		validateFunc, exists := allowedNginxOverrides[key]
		if !exists {
			return errors.New("Config " + key + " is not allowed to override with Architect.")
		}
		var err error
		if err = validateFunc(value); err != nil {
			return err
		}
	}
	return nil
}
