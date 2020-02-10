# Radish
<img align="right" width=280px src="https://images.pexels.com/photos/244393/pexels-photo-244393.jpeg?cs=srgb&dl=close-up-colors-farm-produce-244393.jpg&fm=jpg">
For the dozers of Fraggle Rock, radishes are the beginning of everything they build, because dozer sticks are made from it.
For Aurora, Radish is the beginning of every running java application. 

The first task of Radish is a small process wrapper that:

* Forwards signals
* Reap child processes (PID 1)
* Rewrites exit codes from JVM
* Handles crash reports (currently passing to stdout)
* Generates JVM arguments based on Cgroup-limits and runtime config. See [source](pkg/executor/java_options.go)

The execution is based on a [radish descriptor](pkg/executor/testdata/testconfig.json)

The second task of Radish is a CLI to accomplish a number of tasks:

```
  generateEnvScript          Use to set environment variables from appropriate properties files, based on app- and aurora versions.
  generateNginxConfiguration Use to generate Nginx configuration files based on a Radish descriptor
  generateSplunkStanzas      Use to generate Splunk stanzas. If a stanza template file is provided, use it, if not, use default stanzas.
  printCP                    Prints complete classpath Radish will use with java application
  runJava                    Runs a Java process with Radish

```

# Config read by Radish

| Environment variable |Description |
| ---| ---| 
| JAVA_OPTIONS | Checked for already set options |
| JAVA_MAX_MEM_RATIO | adjust the ratio in percent of memory set as XMX. Default 25%. Remember that memory is more than heap.|
| ENABLE_REMOTE_DEBUG| turn on remote debuging on DEBUG_PORT (default 5005) |
| ENABLE_EXIT_ON_OOM | If set to a non-empty string, the JVM will exit on OutOfMemoryError. Default is off. |
| ENABLE_JAVA_DIAGNOSTICS | If set to a non-empty string, the JVM is started with diagnostics flags set. Default is off.| 
| ENABLE_JOLOKIA | Enables the Jolokia-agent if set.|
| SPLUNK_INDEX | Splunk Index to use for application logging.|
| SPLUNK_AUDIT_INDEX | Splunk Index for audit logs.|
| SPLUNK_APPDYNAMICS_INDEX | Splunk Index for APM logs.|
| SPLUNK_ATS_INDEX | Splunk Index for ATS/STS logs.|
| SPLUNK_BLACKLIST | Rules for which incomming files that should be excluded. Must be a valid PCRE2 regular expression.|
| IGNORE_NGINX_EXCLUDE | Ignore Nginx exclude for Nginx configuration.|
| PROXY_PASS_HOST | Proxypass host for Nginx configuration.|
| PROXY_PASS_PORT | Proxypass port for Nginx configuration.|
| NGINX_WORKER_CONNECTIONS | Number of worker connections for Nginx configuration. |

# Build:

Install go dep (https://github.com/golang/dep)

* dep ensure
* make


# Usage - Process wrapper mode:

Example:

`radish runJava`

Will search for the radish descriptor in the following locations:

* The location given in the optional arg
* The environment variable RADISH_DESCRIPTOR
* /u01/app/radish.json
* /radish.json

# Usage - CLI mode

See help text - type radish -h

