package nodejs

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
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

func TestGenerateNginxConfiguration(t *testing.T) {
	err := GenerateNginxConfiguration("testdata/testconfig.json", "testdata/")
	assert.Equal(t, nil, err)

	data, err := ioutil.ReadFile("testdata/nginx.conf")
	assert.Equal(t, nil, err)

	s := string(data[:])
	assert.Equal(t, s, ninxConfigFile)
}
