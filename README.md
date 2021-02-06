# ibeam-lib-config


This library is meant for providing configuration to a file in skaarOS in an easy way

It also solves the issues of storing type information for all special cases int he toml format where type information would be lost (eg in arrays)


## Usage 

There are 3 functions provided by this library:


**SetCoreName:** Call this first setting the name of your core, eg "core-example"

**SetDevMode:** Call this if you are running locally and not on skaarOS (check core-template how this can be done automatically)

Then create a config structure. Toml keys will become labels in the skaarOS webui. use the toml struct tags if you like your field names to be different!

Also there are some helpful conventions to improve the webuis understanding of your config:
* When using IPs: use a string and suffix the key with "_ip"
* When using Ports: use a uin16 and suffix the key with "_port"
* When using Passwsord: use a string and suffix the key with "_password"

Create a default instance of your config structure. It it needs to be used on multiple go routines use [atomic.Value](https://golang.org/pkg/sync/atomic/) or a **sync.Mutex** to properly protect it (Always check your core with the race detector `go run --race .`)

**Load** When the default config has been filled with default values pass a pointer to the structure to load. The library will automatically load the correct file, and also create it when necessary
