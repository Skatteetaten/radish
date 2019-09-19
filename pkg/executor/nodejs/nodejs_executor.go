package nodejs

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
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
func GenerateNginxConfiguration(nginxTemplatePath string, radishDescriptorPath string, nginxPath string) error {
	var openshiftConfig map[string]interface{}
	var template string = nginxConfigTemplate

	if nginxTemplatePath != "" {
		templatePathDat, err := ioutil.ReadFile(nginxTemplatePath)
		if err != nil {
			return err
		}

		if templatePathDat != nil {
			template = string(templatePathDat[:])
		} else {
			fmt.Println("No template added, using default")
		}
	}

	if radishDescriptorPath != "" {
		descriptorPathDat, err := ioutil.ReadFile(radishDescriptorPath)
		err = json.Unmarshal(descriptorPathDat, &openshiftConfig)
		if err != nil {
			return fmt.Errorf("Error mapping descriptor to json")
		}
	} else {
		return fmt.Errorf("Radish descriptor is missing")
	}

	input, err := mapDataDescToTemplateInput(openshiftConfig)
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

func mapDataDescToTemplateInput(openshiftConfig map[string]interface{}) (*executor.TemplateInput, error) {
	var spa = false
	var hasNodeJSApplication = false
	var openshiftConfigPath string = ""
	var extraStaticHeaders = make(map[string]string)
	var openshiftConfigNodeJSOverrides = make(map[string]string)

	if openshiftConfig["web"] == nil {
		return nil, fmt.Errorf("No web element in openshift.json file")
	}

	if openshiftConfig["web"].(map[string]interface{})["webapp"] == nil {
		return nil, fmt.Errorf("No web.webapp element in openshift.json file")
	}

	if openshiftConfig["web"].(map[string]interface{})["nodejs"] == nil {
		return nil, fmt.Errorf("No web.nodejs element in openshift.json file")
	}

	configurableProxy := false
	for key, value := range openshiftConfig["web"].(map[string]interface{}) {
		if key == "configurableProxy" {
			configurableProxy = value.(bool)
			break
		}
	}

	for key, value := range openshiftConfig["web"].(map[string]interface{})["webapp"].(map[string]interface{}) {
		if key == "path" {
			openshiftConfigPath = value.(string)
		} else if key == "headers" {
			for k, v := range value.(map[string]interface{}) {
				extraStaticHeaders[k] = v.(string)
			}
		}
	}

	for key, value := range openshiftConfig["web"].(map[string]interface{})["nodejs"].(map[string]interface{}) {
		if key == "disableTryfiles" {
			spa = value.(bool)
		} else if key == "main" {
			hasNodeJSApplication = true
		} else if key == "overrides" {
			for k, v := range value.(map[string]interface{}) {
				openshiftConfigNodeJSOverrides[k] = v.(string)
			}
		}
	}

	path := "/"
	if len(strings.TrimPrefix(openshiftConfigPath, "/")) > 0 {
		path = "/" + strings.TrimPrefix(openshiftConfigPath, "/")
	}

	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	err := whitelistOverrides(openshiftConfigNodeJSOverrides)
	if err != nil {
		return nil, err
	}

	env := make(map[string]string)
	env["PROXY_PASS_HOST"] = "localhost"
	env["PROXY_PASS_PORT"] = "9090"

	return &executor.TemplateInput{
		HasNodeJSApplication: hasNodeJSApplication,
		NginxOverrides:       openshiftConfigNodeJSOverrides,
		ConfigurableProxy:    configurableProxy,
		ExtraStaticHeaders:   extraStaticHeaders,
		SPA:                  spa,
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
