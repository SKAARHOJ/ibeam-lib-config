# ibeam-lib-config


This library is meant for providing configuration to a file in skaarOS in an easy way
It dumps the config, a schema in json and a default config into the correct directory in skaarOS or next to the binary in devmode

## Usage 

There are 3 functions provided by this library:

- **SetCoreName:** Call this first setting the name of your core, eg "core-example"
- **SetDevMode:** Call this if you are running locally and not on skaarOS (check core-template how this can be done automatically)
- **Load** When the default config has been filled with default values pass a pointer to the structure to load. The library will automatically load the correct file, and also create it when necessary

Then create a config structure. Toml keys will become labels in the skaarOS webui. use the toml struct tags if you like your field names to be different!

Also there are some helpful struct tags to improve the webuis understanding of your config:
* Validation: use `ibValidator:"ip"` to provide validations in the webUI, all possible validators are: "ip", "port", "password", "devices"
* Descriptions: use a `ibDescription:"Some description"` to provide a description for the current files
* When using a string use: `ibOptions:"Option1,Option2,Option3"` to provide options in a select
* To indicate certain files for dispatch use: `ibDispatch:"devices"`, all possible options are currently: "devices"

Create a default instance of your config structure. It it needs to be used on multiple go routines use a **sync.Mutex** to properly protect it (Always check your core with the race detector `go run --race .`)

