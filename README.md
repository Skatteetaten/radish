# Radish
<img align="right" width=280px src="https://images.pexels.com/photos/244393/pexels-photo-244393.jpeg?cs=srgb&dl=close-up-colors-farm-produce-244393.jpg&fm=jpg">
For the dozers of Fraggle Rock, radishes are the beginning of everything they build, because dozer sticks are made from it.
For Aurora, Radish is the beginning of every running java application. 

The first mode of Radish is a small process wrapper that:

* Forwards signals
* Reap child processes (PID 1)
* Rewrites exit codes from JVM
* Handles crash reports (currently passing to stdout)

The second mode of Radish is a CLI to accomplish a number of tasks.

* generateSplunkStanzas - based on template (optional) and configuration, generate Splunk stanza file where specified.
* generateStartScript - based on configuration file, generate start script for use in Docker container for the application. The start script would typically use the first mode of radish to start the application java process.


# Build:

Install go dep (https://github.com/golang/dep)

* dep ensure
* make


# Usage - Process wrapper mode:

`radish <javaargs> MainClass <args>`

Example:

`radish -XX:+CrashOnOutOfMemoryError -Xmx10m -cp . Main`


# Usage - CLI mode

See help text - type radish -h

