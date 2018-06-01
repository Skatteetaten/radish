# Radish
<img align="right" width=280px src="https://images.pexels.com/photos/244393/pexels-photo-244393.jpeg?cs=srgb&dl=close-up-colors-farm-produce-244393.jpg&fm=jpg">
For the dozers of Fraggle Rock, radishes are the beginning of everything they build, because dozer sticks are made from it.
For Aurora, Radish is the beginning of every running java application. It is a small process wrapper that:


* Forwards signals
* Reap child processes (PID 1)
* Rewrites exit codes from JVM
* Handles crash reports (currently passing to stdout)




# Build:

Install go dep (https://github.com/golang/dep)

* dep ensure
* go install


# Usage:

`radish <javaargs> MainClass <args>`

Example:

`radish -XX:+CrashOnOutOfMemoryError -Xmx10m -cp . Main`

