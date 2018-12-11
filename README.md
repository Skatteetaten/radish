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

The second task of Radish is a CLI to accomplish a number of tasks.

* generateSplunkStanzas - based on template (optional) and configuration, generate Splunk stanza file where specified.
* generateEnvScript - Prints a script that can be sources into a shell exposing configuration from secrets as env variables

# Config read by Radish

| Environment variable |Description |
| ---| ---| 
| JAVA_OPTIONS | Checked for already set options |
| JAVA_MAX_MEM_RATIO | adjust the ratio in percent of memory set as XMX. Default 25%. Remember that memory is more than heap.|
| ENABLE_REMOTE_DEBUG| turn on remote debuging on DEBUG_PORT (default 5005) |
| ENABLE_EXIT_ON_OOM | If set to a non-empty string, the JVM will exit on OutOfMemoryError. Default is off. |
| ENABLE_JAVA_DIAGNOSTICS | If set to a non-empty string, the JVM is started with diagnostics flags set. Default is off.| 
| ENABLE_JOLOKIA | Enables the Jolokia-agent if set.|

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

