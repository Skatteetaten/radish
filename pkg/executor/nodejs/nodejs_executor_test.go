package nodejs

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

    keepalive_timeout  65;

    #gzip  on;

    index index.html;

    server {
       listen 8080;

       location /api {
          proxy_pass http://${PROXY_PASS_HOST}:${PROXY_PASS_PORT};
          client_max_body_size 10m;
       }

       location /web/ {
          root /u01/static;
       }
    }
}
`

func TestGenerateNginxConfigurationFromDefaultTemplate(t *testing.T) {
	os.Setenv("RADISH_NODEJS_APP", "true")
	os.Setenv("RADISH_CONFIGURABLE_PROXY", "true")
	os.Setenv("RADISH_WEB_APP_PATH", "/web")
	os.Setenv("RADISH_NODEJS_OVERRIDES", "{\"client_max_body_size\":\"10m\"}")

	err := GenerateNginxConfiguration("", "", "testdata/nginx.conf")
	assert.Equal(t, nil, err)

	data, err := ioutil.ReadFile("testdata/nginx.conf")
	assert.Equal(t, nil, err)

	s := string(data[:])
	assert.Equal(t, s, ninxConfigFile)
}

func TestGenerateNginxConfigurationFromFile(t *testing.T) {
	os.Setenv("RADISH_NODEJS_APP", "true")
	os.Setenv("RADISH_CONFIGURABLE_PROXY", "true")
	os.Setenv("RADISH_WEB_APP_PATH", "/web")
	os.Setenv("RADISH_NODEJS_OVERRIDES", "{\"client_max_body_size\":\"10m\"}")

	err := GenerateNginxConfiguration("testdata/testconfig.template", "", "testdata/nginx.conf")
	assert.Equal(t, nil, err)

	data, err := ioutil.ReadFile("testdata/nginx.conf")
	assert.Equal(t, nil, err)

	s := string(data[:])
	assert.Equal(t, s, ninxConfigFile)
}

func TestGenerateNginxConfigurationFromContent(t *testing.T) {
	os.Setenv("RADISH_NODEJS_APP", "true")
	os.Setenv("RADISH_CONFIGURABLE_PROXY", "true")
	os.Setenv("RADISH_WEB_APP_PATH", "/web")
	os.Setenv("RADISH_NODEJS_OVERRIDES", "{\"client_max_body_size\":\"10m\"}")

	data, err := ioutil.ReadFile("testdata/testconfig.template")
	err = GenerateNginxConfiguration("", string(data[:]), "testdata/nginx.conf2")
	assert.Equal(t, nil, err)

	data, err = ioutil.ReadFile("testdata/nginx.conf2")
	assert.Equal(t, nil, err)

	s := string(data[:])
	assert.Equal(t, s, ninxConfigFile)
}

func TestGenerateNginxConfigurationNoContent(t *testing.T) {
	err := GenerateNginxConfiguration("", "", "testdata/nginx.conf")
	assert.True(t, err == nil)
}
