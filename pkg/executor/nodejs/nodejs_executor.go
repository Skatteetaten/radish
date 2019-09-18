package nodejs

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/plaid/go-envvar/envvar"
	"github.com/sirupsen/logrus"
	"github.com/skatteetaten/radish/pkg/executor"
	"github.com/skatteetaten/radish/pkg/util"
	"io/ioutil"
	"regexp"
	"strings"
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

    keepalive_timeout  75;

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

//GenerateNginxConfiguration :
func GenerateNginxConfiguration(radishDescriptorPath string, radishDescriptor string, nginxPath string) error {
	var err error
	var dat []byte = nil
	var template string = nginxConfigTemplate

	descriptor := Descriptor{}
	if err := envvar.Parse(&descriptor); err != nil {
		return errors.Wrap(err, "Couldn't parse environment variables")
	}

	if radishDescriptorPath != "" {
		dat, err = ioutil.ReadFile(radishDescriptorPath)
		if err != nil {
			return err
		}
	} else if radishDescriptor != "" {
		dat = []byte(radishDescriptor)
	}

	if dat != nil {
		template = string(dat[:])
	} else {
		fmt.Println("No template added, using default")
	}

	input, err := mapDataDescToTemplateInput(descriptor)
	if err != nil {
		return fmt.Errorf("Error mapping data to template")
	}

	writer := util.NewTemplateWriter(input, "nginx.conf", template)
	if writer == nil {
		return errors.Wrap(err, "Error creating nginx configuration")
	}

	fileWriter := util.NewFileWriter(nginxPath)
	fileWriter(writer)
	return nil
}

func mapDataDescToTemplateInput(descriptor Descriptor) (*executor.TemplateInput, error) {
	path := "/"
	if len(strings.TrimPrefix(descriptor.Data.WebappPath, "/")) > 0 {
		path = "/" + strings.TrimPrefix(descriptor.Data.WebappPath, "/")
	} else if len(strings.TrimPrefix(descriptor.Data.Path, "/")) > 0 {
		logrus.Warnf("web.path in openshift.json is deprecated. Please use web.webapp.path when setting path: %s", descriptor.Data.Path)
		path = "/" + strings.TrimPrefix(descriptor.Data.Path, "/")
	}

	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	extraHeaders := descriptor.Data.ExtraHeaders
	extraHeadersMap := make(map[string]string)
	err := json.Unmarshal([]byte(extraHeaders), &extraHeadersMap)
	if err != nil {
		logrus.Warnf("Could not convert ExtraHeaders to hashmap, proceed without overrides.")
		extraHeadersMap = make(map[string]string)
	}

	overrides := descriptor.Data.NodeJSOverrides
	overridesMap := make(map[string]string)
	err = json.Unmarshal([]byte(overrides), &overridesMap)
	if err != nil {
		logrus.Warnf("Could not convert NodeJSOverrides to hashmap, proceed without overrides.")
		overridesMap = make(map[string]string)
	}

	err = whitelistOverrides(overridesMap)
	if err != nil {
		return nil, err
	}

	env := make(map[string]string)
	env["PROXY_PASS_HOST"] = "localhost"
	env["PROXY_PASS_PORT"] = "9090"

	return &executor.TemplateInput{
		HasNodeJSApplication: descriptor.Data.HasNodeJSApplication,
		NginxOverrides:       overridesMap,
		ConfigurableProxy:    descriptor.Data.ConfigurableProxy,
		ExtraStaticHeaders:   extraHeadersMap,
		SPA:                  descriptor.Data.SPA,
		Path:                 path,
		Env:                  env,
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
