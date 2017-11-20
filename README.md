# A small PoC of a process wrapper that:

* Forwards signals
* Reap child processes (PID 1)
* Rewrites exit codes from JVM
* Handles crash reports


# Build:

Install go dep (https://github.com/golang/dep)

dep ensure
go install .


# Usage:

eve <javaargs> MainClass <args>

Example:

eve -XX:+CrashOnOutOfMemoryError -Xmx10m -cp . Main

