package nginx

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

const ninxConfigFile = `
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
          proxy_pass http://localhost:9090;
          client_max_body_size 10m;
       }

       location /web/ {
          root /u01/static;
          add_header SomeHeader "SomeValue";
       }
    }
}
`

const ninxConfigFileWithCustomEnvParams = `
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
          proxy_pass http://127.0.0.1:9099;
          client_max_body_size 10m;
       }

       location /web/ {
          root /u01/static;
          add_header SomeHeader "SomeValue";
       }
    }
}
`

func TestGenerateNginxConfigurationFromDefaultTemplate(t *testing.T) {
	err := GenerateNginxConfiguration("testdata/testRadishConfig.json", "testdata/nginx.conf")
	assert.Equal(t, nil, err)

	data, err := ioutil.ReadFile("testdata/nginx.conf")
	assert.Equal(t, nil, err)

	s := string(data[:])
	assert.Equal(t, s, ninxConfigFile)
}

func TestGenerateNginxConfigurationFromDefaultTemplateWithEnvParams(t *testing.T) {
	os.Setenv("PROXY_PASS_HOST", "127.0.0.1")
	os.Setenv("PROXY_PASS_PORT", "9099")
	err := GenerateNginxConfiguration("testdata/testRadishConfig.json", "testdata/nginx.conf")
	assert.Equal(t, nil, err)

	data, err := ioutil.ReadFile("testdata/nginx.conf")
	assert.Equal(t, nil, err)

	s := string(data[:])
	assert.Equal(t, s, ninxConfigFileWithCustomEnvParams)
}

func TestGenerateNginxConfigurationNoContent(t *testing.T) {
	err := GenerateNginxConfiguration("", "testdata/nginx.conf")
	assert.NotEmpty(t, err)
}
