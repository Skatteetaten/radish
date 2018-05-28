# A small process wrapper that:

* Forwards signals
* Reap child processes (PID 1)
* Rewrites exit codes from JVM
* Handles crash reports




# Build:

Install go dep (https://github.com/golang/dep)

dep ensure
go install .


# Usage:

radish <javaargs> MainClass <args>

Example:

radish -XX:+CrashOnOutOfMemoryError -Xmx10m -cp . Main

