# Radish descriptor
Example:
```json
{
  "Type": "JavaDescriptor",
  "Version": "1",
  "Data": {
    "BaseDir": "testdata",
    "PathsToClassLibraries": [
      "lib/lib2/lib4.jar",
      "lib",
      "lib/lib3/**"
    ],
    "JavaOptions": "-Dfoo=bar",
    "MainClass": "foo.bar.Main",
    "ApplicationArgs": "--logging.config=logback.xml"
  }
}
```

## PathsToClassLibraries
Sub directories of BaseDir that will be included on classpath
There are three ways of defining subdirectories

### Fully qualified declaration
```json
...
    "PathsToClassLibraries": [
      "lib/lib2/lib4.jar"
    ]
...
```
Path and name of jar that will be included on classpath.

### All libraries in a directory
```json
...
    "PathsToClassLibraries": [
      "lib"
    ]
...
```
All jar files in folder lib (not subdirectories) will be included on classpath

### All libraries in directory and subdirectories
```json
...
    "PathsToClassLibraries": [
      "lib/lib3/**"
    ]
...
```
All jar files expect jar-files ending with "-exec.jar" in folder lib/lib3/ and all subdirectories will be included on classpath
