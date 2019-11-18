package nginx

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"github.com/skatteetaten/radish/pkg/executor"
	"github.com/skatteetaten/radish/pkg/util"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
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

	{{.Gzip}}

    index index.html;

    server {
       listen 8080;

       location /api {
          {{if or .HasNodeJSApplication .ConfigurableProxy}}proxy_pass http://{{.ProxyPassHost}}:{{.ProxyPassPort}};{{else}}return 404;{{end}}{{range $key, $value := .NginxOverrides}}
         {{$key}} {{$value}};{{end}}
      }
    {{range $value := .Exclude}}
	   location {{$value}} {  
		  return 404;
	   }
    {{end}}
{{if .SPA}}
       location {{.Path}} {
          root /u01/static;
          try_files $uri {{.Path}}index.html;{{else}}
       location {{.Path}} {
          root /u01/static;{{end}}{{range $key, $value := .ExtraStaticHeaders}}
          add_header {{$key}} "{{$value}}";{{end}}
	   }
	   
	   {{.Locations}}
    }
}
`

//GenerateNginxConfiguration :
func GenerateNginxConfiguration(openshiftConfigPath string, nginxPath string) error {
	var openshiftConfig OpenshiftConfig

	if openshiftConfigPath != "" {
		data, err := ioutil.ReadFile(openshiftConfigPath)
		if err != nil {
			return fmt.Errorf("Error reading file: " + openshiftConfigPath)
		}

		openshiftConfig, err = UnmarshallOpenshiftConfig(bytes.NewBuffer(data))
		if err != nil {
			return fmt.Errorf("Error mapping openshift json to internal structure")
		}
	} else {
		return fmt.Errorf("OpenshiftConfigPath is missing. Will not generate nginx configuration with radish")
	}

	fileWriter := util.NewFileWriter(nginxPath)

	err := generateNginxConfiguration(openshiftConfig, fileWriter)

	if err != nil {
		return errors.Wrap(err, "Error writing nginx configuration")
	}

	return nil
}

func generateNginxConfiguration(openshiftConfig OpenshiftConfig, fileWriter util.FileWriter) error {

	input, err := mapDataDescToTemplateInput(openshiftConfig)
	if err != nil {
		return fmt.Errorf("Error mapping data to template")
	}

	writer := util.NewTemplateWriter(input, "nginx.conf", nginxConfigTemplate)
	if writer == nil {
		return errors.New("Error creating nginx configuration")
	}

	err = fileWriter(writer, "nginx.conf")

	if err != nil {
		return errors.Wrap(err, "Error writing nginx configuration")
	}

	return nil
}

func mapDataDescToTemplateInput(openshiftConfig OpenshiftConfig) (*executor.TemplateInput, error) {
	documentRoot := "/u01/static"
	path := "/"
	if len(strings.TrimPrefix(openshiftConfig.Web.WebApp.Path, "/")) > 0 {
		path = "/" + strings.TrimPrefix(openshiftConfig.Web.WebApp.Path, "/")
	}

	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	err := whitelistOverrides(openshiftConfig.Web.Nodejs.Overrides)
	if err != nil {
		return nil, err
	}

	exclude := openshiftConfig.Web.Exclude
	ignoreExclude := os.Getenv("IGNORE_NGINX_EXCLUDE")
	if strings.EqualFold("true", ignoreExclude) {
		exclude = []string{}
	}

	proxyPassHost := os.Getenv("PROXY_PASS_HOST")
	if proxyPassHost == "" {
		proxyPassHost = "localhost"
	}

	proxyPassPort := os.Getenv("PROXY_PASS_PORT")
	if proxyPassPort == "" {
		proxyPassPort = "9090"
	}

	nginxGzipForTemplate := nginxGzipMapToString(openshiftConfig.Web.Gzip)
	nginxLocationForTemplate := nginxLocationsMapToString(openshiftConfig.Web.Locations, documentRoot, path)

	return &executor.TemplateInput{
		HasNodeJSApplication: openshiftConfig.Web.Nodejs.Main != "",
		NginxOverrides:       openshiftConfig.Web.Nodejs.Overrides,
		ConfigurableProxy:    openshiftConfig.Web.ConfigurableProxy,
		ExtraStaticHeaders:   openshiftConfig.Web.WebApp.Headers,
		SPA:                  !openshiftConfig.Web.WebApp.DisableTryfiles,
		Path:                 path,
		Gzip:                 nginxGzipForTemplate,
		Exclude:              exclude,
		Locations:            nginxLocationForTemplate,
		ProxyPassHost:        proxyPassHost,
		ProxyPassPort:        proxyPassPort,
	}, nil
}

/*

We sanitize the input.... Don't want to large inputs.

For example; Accepting very large client_max_body_size would make a DOS attack very easy to implement...

*/
var allowedNginxOverrides = map[string]func(string) error{
	"client_max_body_size": func(s string) error {
		// between 1 and 50
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

func (m nginxLocations) sort() []string {
	index := []string{}
	for k := range m {
		index = append(index, k)
	}
	sort.Strings(index)
	return index
}

func (m headers) sort() []string {
	index := []string{}
	for k := range m {
		index = append(index, k)
	}
	sort.Strings(index)
	return index
}

func nginxLocationsMapToString(m nginxLocations, documentRoot string, path string) string {
	sumLocations := ""
	indentN1 := strings.Repeat(" ", 8)
	indentN2 := strings.Repeat(" ", 12)

	for _, k := range m.sort() {
		v := m[k]
		singleLocation := fmt.Sprintf("%slocation %s%s {\n", indentN1, path, k)
		singleLocation = fmt.Sprintf("%s%sroot %s;\n", singleLocation, indentN2, documentRoot)
		gZipUse := strings.TrimSpace(v.Gzip.Use)
		gZipVary := strings.TrimSpace(v.Gzip.Vary)
		if gZipUse == "on" {
			singleLocation = fmt.Sprintf("%s%sgzip on;\n", singleLocation, indentN2)
			if v.Gzip.MinLength > 0 {
				singleLocation = fmt.Sprintf("%s%sgzip_min_length %d;\n", singleLocation, indentN2, v.Gzip.MinLength)
			}
			if gZipVary != "" {
				singleLocation = fmt.Sprintf("%s%sgzip_vary %s;\n", singleLocation, indentN2, gZipVary)
			}
			if v.Gzip.Proxied != "" {
				singleLocation = fmt.Sprintf("%s%sgzip_proxied %s;\n", singleLocation, indentN2, v.Gzip.Proxied)
			}
			if v.Gzip.Types != "" {
				singleLocation = fmt.Sprintf("%s%sgzip_types %s;\n", singleLocation, indentN2, v.Gzip.Types)
			}
			if v.Gzip.Disable != "" {
				singleLocation = fmt.Sprintf("%s%sgzip_disable \"%s\";\n", singleLocation, indentN2, v.Gzip.Disable)
			}
		} else if gZipUse == "off" {
			singleLocation = fmt.Sprintf("%s%sgzip off;\n", singleLocation, indentN2)
		}

		for _, k2 := range v.Headers.sort() {
			singleLocation = fmt.Sprintf("%s%sadd_header %s \"%s\";\n", singleLocation, indentN2, k2, v.Headers[k2])
		}

		singleLocation = fmt.Sprintf("%s%s}\n", singleLocation, indentN1)
		sumLocations = sumLocations + singleLocation
	}
	return sumLocations
}

func nginxGzipMapToString(gzip nginxGzip) string {
	sumGzip := ""
	indent := strings.Repeat(" ", 4)
	if gzip.Use == "on" {
		sumGzip = fmt.Sprintf("%s%sgzip on;\n", sumGzip, indent)
		if gzip.MinLength > 0 {
			sumGzip = fmt.Sprintf("%s%sgzip_min_length %d;\n", sumGzip, indent, gzip.MinLength)
		}
		if gzip.Vary != "" {
			sumGzip = fmt.Sprintf("%s%sgzip_vary %s;\n", sumGzip, indent, gzip.Vary)
		}
		if gzip.Proxied != "" {
			sumGzip = fmt.Sprintf("%s%sgzip_proxied %s;\n", sumGzip, indent, gzip.Proxied)
		}
		if gzip.Types != "" {
			sumGzip = fmt.Sprintf("%s%sgzip_types %s;\n", sumGzip, indent, gzip.Types)
		}
		if gzip.Disable != "" {
			sumGzip = fmt.Sprintf("%s%sgzip_disable \"%s\";\n", sumGzip, indent, gzip.Disable)
		}
	} else {
		sumGzip = fmt.Sprintf("%s%sgzip off;\n", sumGzip, indent)
	}
	return sumGzip
}
