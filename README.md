# Radish

<img align="right" width=280px src="https://images.pexels.com/photos/244393/pexels-photo-244393.jpeg?cs=srgb&dl=close-up-colors-farm-produce-244393.jpg&fm=jpg">
For the dozers of Fraggle Rock, radishes are the beginning of everything they build, because dozer sticks are made from it.
For Aurora, Radish is the beginning of every running java application. 

The first task of Radish is a small process wrapper that:

* Forwards signals
* Reap child processes (PID 1)
* Rewrites exit codes from JVM
* Handles crash reports (currently passing to stdout)
* Generates JVM arguments based on Cgroup-limits and runtime config. See [source](pkg/executor/java/java_options.go)

The execution is based on a [radish descriptor](pkg/executor/testdata/testconfig.json)

The second task of Radish is a CLI to accomplish a number of tasks:

```
  generateEnvScript          Use to set environment variables from appropriate properties files, based on app- and aurora versions.
  generateNginxConfiguration Use to generate Nginx configuration files based on a Radish descriptor
  printCP                    Prints complete classpath Radish will use with java application
  runJava                    Runs a Java process with Radish
  runNnginx                  Runs a Nginx process with support for logrotate. 
```

# Config read by Radish

| Environment variable     | Description                                                                                                                                                                                                                                     |
|--------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------| 
| JAVA_OPTIONS             | Checked for already set options                                                                                                                                                                                                                 |
| JAVA_MAX_MEM_RATIO       | ONLY Java 8: adjust the ratio in percent of memory set as XMX. Default 25%. Remember that memory is more than heap.                                                                                                                             |
| JAVA_MAX_RAM_PERCENTAGE  | Java 11+: adjust the max amount of memory set as XX:MaxRAMPercentage. Default 75.0. Remember that memory is more than heap.                                                                                                                     |
| ENABLE_REMOTE_DEBUG      | turn on remote debuging on DEBUG_PORT (default 5005)                                                                                                                                                                                            |
| ENABLE_EXIT_ON_OOM       | If set to a non-empty string, the JVM will exit on OutOfMemoryError. Default is off.                                                                                                                                                            |
| ENABLE_JAVA_DIAGNOSTICS  | If set to a non-empty string, the JVM is started with diagnostics flags set. Default is off.                                                                                                                                                    | 
| ENABLE_JOLOKIA           | Enables the Jolokia-agent if set.                                                                                                                                                                                                               |
| SPLUNK_INDEX             | Splunk Index to use for application logging.                                                                                                                                                                                                    |
| SPLUNK_AUDIT_INDEX       | Splunk Index for audit logs.                                                                                                                                                                                                                    |
| SPLUNK_APPDYNAMICS_INDEX | Splunk Index for APM logs.                                                                                                                                                                                                                      |
| SPLUNK_ATS_INDEX         | Splunk Index for ATS/STS logs.                                                                                                                                                                                                                  |
| SPLUNK_BLACKLIST         | Rules for which incomming files that should be excluded. Must be a valid PCRE2 regular expression.                                                                                                                                              |
| IGNORE_NGINX_EXCLUDE     | Ignore Nginx exclude for Nginx configuration.                                                                                                                                                                                                   |
| PROXY_PASS_HOST          | Proxypass host for Nginx configuration. Default localhost.                                                                                                                                                                                      |
| PROXY_PASS_PORT          | Proxypass port for Nginx configuration. Default 9090.                                                                                                                                                                                           |
| NGINX_WORKER_CONNECTIONS | Number of worker connections for Nginx configuration. Default 1024.                                                                                                                                                                             |
| NGINX_WORKER_PROCESSES   | Number of worker processes for Nginx configuration. Default 1.                                                                                                                                                                                  |
| RADISH_SIGNAL_FORWARD_DELAY | The delay in second from a signal is received by radish until it is sent to the child process. Default is 0                                                                                                                                     |
| NGINX_PROXY_READ_TIMEOUT | Read timeout configuration. Default is 60                                                                                                                                                                                                       |
| NGINX_LOG_STRATEGY       | Nginx indexing strategy is either set to `file` or `stdout`. Note: The `stdout` strategy is only available in OCP3 clusters.                                                                                                                    
| ENABLE_OTEL_TRACE        | Enables Opentelemetry tracing via agent if set to true. For additional config parameters see https://github.com/open-telemetry/opentelemetry-java/blob/main/sdk-extensions/autoconfigure/README.md#otlp-exporter-both-span-and-metric-exporters |

# Build:

We use go modules

Dependencies are managed via `go.mod`. Remember to run `go mod tidy` after dependency update.

The build is orchestrated on Jenkins, with Jenkinsfile.

### Local build

We use nginx when validating the generated nginx configuration, thus nginx is required to run tests locally.

Nginx can be installed in one of the following ways:

Mac:

`brew install nginx`

Linux (Ubuntu)

`sudo apt install nginx`

Build commands:
```
make                       # Build and test
make binary                # Build and skip tests
```

# Usage - Process wrapper mode:

Example:

`radish runJava`

Will search for the radish descriptor in the following locations:

* The location given in the optional arg
* The environment variable RADISH_DESCRIPTOR
* /u01/app/radish.json
* /radish.json

Example:

`radish runNginx --nginxPath --nginxPath=/tmp/nginx/nginx.conf`

Will start the nginx server with nginx configuration located at /tmp/nginx/nginx.conf

If `NGINX_LOG_STRATEGY` is set to `file` logs are written to `/u01/logs/nginx.log` and `/u01/logs/nginx.access` in
addition to stdout /stderr

# Usage - CLI mode

See help text - type radish -h

## New java version

For a new java version, the java options are listed here: https://chriswhocodes.com/vm-options-explorer.html.
