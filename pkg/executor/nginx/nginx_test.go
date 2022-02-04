package nginx

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/skatteetaten/radish/pkg/util"
	"github.com/stretchr/testify/assert"
)

const ninxConfigFile = `
worker_processes  1;
error_log stderr;
error_log /u01/logs/nginx.log;

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

	access_log /u01/logs/nginx.access;
	
	sendfile        on;
	#tcp_nopush     on;
	server_tokens  off;

    keepalive_timeout  75;
    proxy_read_timeout 60;

	gzip_static off;


	index index.html;

	server {
		listen 8080;

		location /api {
		proxy_pass http://localhost:9090;
		proxy_http_version 1.1;
		client_max_body_size 10m;
	}
	

		location /web/ {
			root /u01/static;
			try_files $uri /web/index.html;
			add_header SomeHeader "SomeValue";
		}	

		location =/ {
			if ($request_method = HEAD) {
			return 200;
		}
		}
	
	}
}
`

const nginxConfigFileWithGzipStatic = `
worker_processes  1;
error_log stderr;
error_log /u01/logs/nginx.log;

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
	access_log /u01/logs/nginx.access;

	sendfile        on;
	#tcp_nopush     on;
	server_tokens  off;

    keepalive_timeout  75;
    proxy_read_timeout 60;

	gzip_static on;
	gzip_vary on;
	gzip_proxied any;
	gzip on;

	index index.html;

	server {
		listen 8080;

		location /api {
			proxy_pass http://localhost:9090;
			proxy_http_version 1.1;
			client_max_body_size 10m;
	}
	

		location /web/ {
			root /u01/static;
			try_files $uri /web/index.html;
			add_header SomeHeader "SomeValue";
		}
		
		location =/ {
			if ($request_method = HEAD) {
				return 200;
			}
		}
	
	}
}
`

const ninxConfigFileWithCustomEnvParams = `
worker_processes  1;
error_log stderr;
error_log /u01/logs/nginx.log;

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
	access_log /u01/logs/nginx.access;
	
	sendfile        on;
	#tcp_nopush     on;
	server_tokens  off;

    keepalive_timeout  75;
    proxy_read_timeout 5;

	gzip_static off;


	index index.html;

	server {
		listen 8080;

		location /api {
			proxy_pass http://127.0.0.1:9099;
			proxy_http_version 1.1;
			client_max_body_size 10m;
		}
	

		location /web/ {
			root /u01/static;
			try_files $uri /web/index.html;
			add_header SomeHeader "SomeValue";
		}

		location =/ {
			if ($request_method = HEAD) {
				return 200;
			}
		}
	}
}
`

const nginxConfigWithExclude = `
worker_processes  1;
error_log stderr;
error_log /u01/logs/nginx.log;

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
	access_log /u01/logs/nginx.access;

	sendfile        on;
	#tcp_nopush     on;
	server_tokens  off;

    keepalive_timeout  75;
    proxy_read_timeout 60;

	gzip_static off;


	index index.html;

	server {
		listen 8080;

		location /api {
			proxy_pass http://localhost:9090;
			proxy_http_version 1.1;
			client_max_body_size 10m;
		}
	
		location test/fil1.swf {  
			return 404;
		}
		
		location test/fil2.png {  
			return 404;
		}
		

		location /web/ {	
			root /u01/static;
			try_files $uri /web/index.html;
			add_header SomeHeader "SomeValue";
		}

			location =/ {
			if ($request_method = HEAD) {
				return 200;
			}
		}
		
		
	}
}
`

const nginxConfWithCustomLocations = `
worker_processes  1;
error_log stderr;
error_log /u01/logs/nginx.log;

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
	access_log /u01/logs/nginx.access;

	sendfile        on;
	#tcp_nopush     on;
	server_tokens  off;

    keepalive_timeout  75;
    proxy_read_timeout 60;

		gzip_static off;


	index index.html;

	server {
		listen 8080;

		location /api {
		proxy_pass http://localhost:9090;
			proxy_http_version 1.1;
			client_max_body_size 10m;
		}
		

		location /web/ {
			root /u01/static;
			try_files $uri /web/index.html;
			add_header SomeHeader "SomeValue";
		}
		
				location /web/index.html {
			root /u01/static;
			gzip_static on;
			gzip_vary on;
			gzip_proxied any;
			gzip on;
			add_header Cache-Control "no-cache";
			add_header X-Frame-Options "DENY";
			add_header X-XSS-Protection "1";
		}
		location /web/index/other.html {
			root /u01/static;
			add_header Cache-Control "no-store";
			add_header X-XSS-Protection "1; mode=block";
		}
		location /web/index_other.html {
			root /u01/static;
			add_header Cache-Control "max-age=60";
			add_header X-XSS-Protection "0";
		}

		
		location =/ {
			if ($request_method = HEAD) {
				return 200;
			}
		}

	}
}
`
const nginxConfPrefixWithFileLogging = `
worker_processes  1;
error_log stderr;
error_log /u01/logs/nginx.log;

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
	access_log /u01/logs/nginx.access;

	sendfile        on;
	#tcp_nopush     on;
	server_tokens  off;

    keepalive_timeout  75;
    proxy_read_timeout 60;

	gzip_static off;


	index index.html;
`

const nginxConfPrefix = `
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
	server_tokens  off;

    keepalive_timeout  75;
    proxy_read_timeout 60;

	gzip_static off;


	index index.html;
`

const nginxConfPrefixWithChangedWorkerConnsAndProcesses = `
worker_processes  2;
error_log stderr;

events {
	worker_connections  2048;
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
	server_tokens  off;

    keepalive_timeout  75;
    proxy_read_timeout 60;

	gzip_static off;


	index index.html;
`

const expectedNginxConfFileNoNodejsPartial = `
	server {
		listen 8080;

		location /api {
			return 404;
			
		}
	

		location / {
			root /u01/static;
			try_files $uri /index.html;
		}
		
		
	}
}
`
const expectedNginxConfFilePartial = `
	server {
		listen 8080;

		location /api {
			proxy_pass http://localhost:9090;
			proxy_http_version 1.1;
		}
	

		location / {
			root /u01/static;
			try_files $uri /index.html;
		}
		
		
	}
}
`

const expectedNginxConfFileSpaAndCustomHeaders = `
	server {
		listen 8080;

		location /api {
			proxy_pass http://localhost:9090;
			proxy_http_version 1.1;
		}


		location / {
			root /u01/static;
			try_files $uri /index.html;
			add_header X-Test-Header "Tulleheader";
			add_header X-Test-Header2 "Tulleheader2";
		}
		
		
	}
}
`

const expectedNginxConfFileNoSpaAndCustomHeaders = `
	server {
		listen 8080;

		location /api {
			proxy_pass http://localhost:9090;
			proxy_http_version 1.1;
		}
	

		location / {
			root /u01/static;
			add_header X-Test-Header "Tulleheader";
			add_header X-Test-Header2 "Tulleheader2";
		}

	}
}
`

const expectedNginxConfigWithOverrides = `
	server {
		listen 8080;

		location /api {
			proxy_pass http://localhost:9090;
			proxy_http_version 1.1;
			client_max_body_size 5m;
			proxy_buffer_size 14k;
			proxy_buffers 4 14k;
		}


		location / {
			root /u01/static;
			try_files $uri /index.html;
		}


	}
}
`

func TestGeneratedNginxFileWhenNodeJSIsEnabled(t *testing.T) {
	openshiftJSON := OpenshiftConfig{
		Docker: Docker{
			Maintainer: "Tullebukk",
		},
		Web: Web{
			Nodejs: Nodejs{
				Main: "test.json",
			},
		},
	}
	var actual string
	err := generateNginxConfiguration(openshiftJSON, testFileWriter(&actual))

	assert.NoError(t, err)
	assert.Equal(t, cleanString(nginxConfPrefix+expectedNginxConfFilePartial), cleanString(actual))

	validateNginxConfig(t, actual)
}

func TestGeneratedNginxFileWhenWorkerConnsAndProcessesAreChanged(t *testing.T) {
	os.Setenv("NGINX_WORKER_CONNECTIONS", "2048")
	os.Setenv("NGINX_WORKER_PROCESSES", "2")
	openshiftJSON := OpenshiftConfig{
		Docker: Docker{
			Maintainer: "Tullebukk",
		},
		Web: Web{
			Nodejs: Nodejs{
				Main: "test.json",
			},
		},
	}
	var actual string
	err := generateNginxConfiguration(openshiftJSON, testFileWriter(&actual))

	assert.NoError(t, err)

	assert.Equal(t, cleanString(nginxConfPrefixWithChangedWorkerConnsAndProcesses+expectedNginxConfFilePartial), cleanString(actual))

	validateNginxConfig(t, actual)

	os.Unsetenv("NGINX_WORKER_CONNECTIONS")
	os.Unsetenv("NGINX_WORKER_PROCESSES")
}

func TestGeneratedFilesWhenNodeJSIsDisabled(t *testing.T) {
	openshiftJSON := OpenshiftConfig{
		Docker: Docker{
			Maintainer: "Tullebukk",
		},
		Web: Web{
			WebApp: WebApp{
				Path: "",
			},
		},
	}

	var data string
	err := generateNginxConfiguration(openshiftJSON, testFileWriter(&data))

	assert.NoError(t, err)
	assert.Equal(t, cleanString(nginxConfPrefix+expectedNginxConfFileNoNodejsPartial), cleanString(data))

	validateNginxConfig(t, data)
}

func TestThatCustomHeadersIsPresentInNginxConfig(t *testing.T) {
	openshiftJSON := OpenshiftConfig{
		Docker: Docker{
			Maintainer: "tullebukk",
		},
		Web: Web{
			Nodejs: Nodejs{
				Main: "test.json",
			},
			WebApp: WebApp{
				DisableTryfiles: false,
				Headers: map[string]string{
					"X-Test-Header":  "Tulleheader",
					"X-Test-Header2": "Tulleheader2",
				},
				Content: "pathTilStatic",
			},
		},
	}

	var actual string
	err := generateNginxConfiguration(openshiftJSON, testFileWriter(&actual))

	assert.NoError(t, err)
	assert.Equal(t, cleanString(nginxConfPrefix+expectedNginxConfFileSpaAndCustomHeaders), cleanString(actual))

	openshiftJSON.Web.WebApp.DisableTryfiles = true

	validateNginxConfig(t, actual)

	err = generateNginxConfiguration(openshiftJSON, testFileWriter(&actual))

	assert.NoError(t, err)
	assert.Equal(t, cleanString(nginxConfPrefix+expectedNginxConfFileNoSpaAndCustomHeaders), cleanString(actual))

	validateNginxConfig(t, actual)
}

func TestThatOverrideInNginxIsSet(t *testing.T) {
	openshiftJSON := OpenshiftConfig{
		Docker: Docker{
			Maintainer: "Tullebukk",
		},
		Web: Web{
			Nodejs: Nodejs{
				Main: "test.json",
				Overrides: map[string]string{
					"client_max_body_size": "5m",
					"proxy_buffer_size":    "14k",
					"proxy_buffers":        "4 14k",
				},
			},
		},
	}

	var actual string
	err := generateNginxConfiguration(openshiftJSON, testFileWriter(&actual))

	assert.NoError(t, err)
	assert.Equal(t, cleanString(nginxConfPrefix+expectedNginxConfigWithOverrides), cleanString(actual))

	validateNginxConfig(t, actual)

}

func TestGenerateNginxConfigurationFromDefaultTemplate(t *testing.T) {
	os.Setenv("NGINX_LOG_STRATEGY", "file")
	err := GenerateNginxConfiguration("testdata/testRadishConfig.json", "testdata")
	assert.Equal(t, nil, err)

	data, err := ioutil.ReadFile("testdata/nginx.conf")
	assert.Equal(t, nil, err)

	s := string(data[:])
	assert.Equal(t, cleanString(ninxConfigFile), cleanString(s))

	validateNginxConfig(t, s)

	os.Unsetenv("NGINX_LOG_STRATEGY")
}

func TestGenerateNginxConfigurationFromDefaultTemplateWithGzip(t *testing.T) {
	os.Setenv("NGINX_LOG_STRATEGY", "file")
	err := GenerateNginxConfiguration("testdata/testRadishConfigWithGzipStatic.json", "testdata")
	assert.Equal(t, nil, err)

	data, err := ioutil.ReadFile("testdata/nginx.conf")
	assert.Equal(t, nil, err)

	s := string(data[:])
	assert.Equal(t, cleanString(nginxConfigFileWithGzipStatic), cleanString(s))

	validateNginxConfig(t, s)

	os.Unsetenv("NGINX_LOG_STRATEGY")
}

func TestGenerateNginxConfigurationFromDefaultTemplateWithEnvParams(t *testing.T) {
	os.Setenv("NGINX_PROXY_READ_TIMEOUT", "5")
	os.Setenv("PROXY_PASS_HOST", "127.0.0.1")
	os.Setenv("PROXY_PASS_PORT", "9099")
	os.Setenv("NGINX_LOG_STRATEGY", "file")

	err := GenerateNginxConfiguration("testdata/testRadishConfigWithProxy.json", "testdata")
	assert.Equal(t, nil, err)

	data, err := ioutil.ReadFile("testdata/nginx.conf")
	assert.Equal(t, nil, err)

	s := string(data[:])
	assert.Equal(t, cleanString(ninxConfigFileWithCustomEnvParams), cleanString(s))

	validateNginxConfig(t, s)

	// Clean up env params
	os.Unsetenv("NGINX_PROXY_READ_TIMEOUT")
	os.Unsetenv("PROXY_PASS_HOST")
	os.Unsetenv("PROXY_PASS_PORT")
	os.Unsetenv("NGINX_LOG_STRATEGY")
}

func TestGenerateNginxConfigurationWithProxyShouldFailWhenEnvsAreMissing(t *testing.T) {
	err := GenerateNginxConfiguration("testdata/testRadishConfigWithProxy.json", "testdata")
	if err == nil {
		t.Fail()
	}
}

func TestGenerateNginxConfigurationFromDefaultTemplateWithExclude(t *testing.T) {
	os.Setenv("NGINX_LOG_STRATEGY", "file")
	err := GenerateNginxConfiguration("testdata/testRadishConfigWithExclude.json", "testdata")
	assert.Equal(t, nil, err)

	data, err := ioutil.ReadFile("testdata/nginx.conf")
	assert.Equal(t, nil, err)

	s := string(data[:])
	assert.Equal(t, cleanString(nginxConfigWithExclude), cleanString(s))

	validateNginxConfig(t, s)

	os.Unsetenv("NGINX_LOG_STRATEGY")
}

func TestGenerateNginxConfigurationFromDefaultTemplateWithIgnoreExcludeNginxEnvParam(t *testing.T) {
	os.Setenv("IGNORE_NGINX_EXCLUDE", "true")
	os.Setenv("NGINX_LOG_STRATEGY", "file")
	err := GenerateNginxConfiguration("testdata/testRadishConfigWithExclude.json", "testdata")
	assert.Equal(t, nil, err)

	data, err := ioutil.ReadFile("testdata/nginx.conf")
	assert.Equal(t, nil, err)

	s := string(data[:])
	assert.Equal(t, cleanString(ninxConfigFile), cleanString(s))

	validateNginxConfig(t, s)

	// Clean up env params
	os.Unsetenv("IGNORE_NGINX_EXCLUDE")
	os.Unsetenv("NGINX_LOG_STRATEGY")
}

func TestGenerateNginxConfigurationFromDefaultTemplateWithCustomLocations(t *testing.T) {
	os.Setenv("NGINX_LOG_STRATEGY", "file")
	err := GenerateNginxConfiguration("testdata/testRadishConfigWithCustomLocations.json", "testdata")
	assert.Equal(t, nil, err)

	data, err := ioutil.ReadFile("testdata/nginx.conf")
	assert.Equal(t, nil, err)

	s := string(data[:])
	assert.Equal(t, cleanString(nginxConfWithCustomLocations), cleanString(s))

	validateNginxConfig(t, s)

	os.Unsetenv("NGINX_LOG_STRATEGY")
}

func TestGenerateNginxConfigurationNoContent(t *testing.T) {
	err := GenerateNginxConfiguration("", "testdata")
	assert.NotEmpty(t, err)
}

func testFileWriter(res *string) util.FileWriter {
	return func(writer util.WriterFunc, file ...string) error {
		buffer := new(bytes.Buffer)
		err := writer(buffer)
		if err == nil {
			*res = buffer.String()
		}
		return err
	}
}

func cleanString(in string) string {
	replacer := strings.NewReplacer("\n", "", "\t", "")
	return replacer.Replace(in)
}

func validateNginxConfig(t *testing.T, config string) {
	//Not relevant for syntax checking
	config = strings.Replace(config, "include       /etc/nginx/mime.types;", "", -1)

	file, err := ioutil.TempFile("/tmp", "nginx.*.conf")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())

	err = ioutil.WriteFile(file.Name(), []byte(config), 0)
	if err != nil {
		t.Error(err)
	}

	result, err := exec.Command("nginx", "-t", "-c", file.Name()).CombinedOutput()
	logrus.Infof(string(result))

	if !strings.Contains(string(result), "syntax is ok") && err != nil {
		t.Fatalf("Could not validate nginx.conf: %v", err)
	}

}
