package nginx

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/skatteetaten/radish/pkg/executor"
	"github.com/skatteetaten/radish/pkg/util"
)

const nginxConfigTemplate string = `
worker_processes  {{.WorkerProcesses}};
error_log stderr;

events {
    worker_connections  {{.WorkerConnections}};
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
    proxy_read_timeout {{.ProxyReadTimeout}};

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

	proxyPassHost := getEnvOrDefault("PROXY_PASS_HOST", "localhost")

	proxyPassPort := getEnvOrDefault("PROXY_PASS_PORT", "9090")

	proxyReadTimeout := getEnvOrDefault("NGINX_PROXY_READ_TIMEOUT", "1")

	workerConnections := getEnvOrDefault("NGINX_WORKER_CONNECTIONS", "1024")

	workerProcesses := getEnvOrDefault("NGINX_WORKER_PROCESSES", "1")

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
		WorkerConnections:    workerConnections,
		WorkerProcesses:      workerProcesses,
		ProxyReadTimeout:     proxyReadTimeout,
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

	for _, key := range m.sort() {
		value := m[key]
		singleLocation := fmt.Sprintf("%slocation %s%s {\n", indentN1, path, key)
		singleLocation = fmt.Sprintf("%s%sroot %s;\n", singleLocation, indentN2, documentRoot)

		gZipUse := strings.TrimSpace(value.Gzip.Use)
		if gZipUse == "on" || gZipUse == "off" {
			singleLocation = getGzipConfAsString(value.Gzip, singleLocation, indentN2)
		}

		for _, k2 := range value.Headers.sort() {
			singleLocation = fmt.Sprintf("%s%sadd_header %s \"%s\";\n", singleLocation, indentN2, k2, value.Headers[k2])
		}

		singleLocation = fmt.Sprintf("%s%s}\n", singleLocation, indentN1)
		sumLocations = sumLocations + singleLocation
	}
	return sumLocations
}

func nginxGzipMapToString(gzip nginxGzip) string {
	indent := strings.Repeat(" ", 4)
	return getGzipConfAsString(gzip, "", indent)
}

func getGzipConfAsString(gzip nginxGzip, location string, indent string) string {
	if strings.TrimSpace(gzip.Use) == "on" {
		location = fmt.Sprintf("%s%sgzip on;\n", location, indent)
		if gzip.MinLength > 0 {
			location = fmt.Sprintf("%s%sgzip_min_length %d;\n", location, indent, gzip.MinLength)
		}
		if gzip.Vary != "" {
			location = fmt.Sprintf("%s%sgzip_vary %s;\n", location, indent, strings.TrimSpace(gzip.Vary))
		}
		if gzip.Proxied != "" {
			location = fmt.Sprintf("%s%sgzip_proxied %s;\n", location, indent, gzip.Proxied)
		}
		if gzip.Types != "" {
			location = fmt.Sprintf("%s%sgzip_types %s;\n", location, indent, gzip.Types)
		}
		if gzip.Disable != "" {
			location = fmt.Sprintf("%s%sgzip_disable \"%s\";\n", location, indent, gzip.Disable)
		}
	} else {
		location = fmt.Sprintf("%s%sgzip off;\n", location, indent)
	}
	return location
}

func getEnvOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}
