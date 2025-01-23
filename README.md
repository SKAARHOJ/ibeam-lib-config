# ibeam-lib-config


This library is meant for providing configuration to a file in skaarOS in an easy way
It dumps the config, a schema in json and a default config into the correct directory in skaarOS or next to the binary in other environments

## Usage 

There are 3 functions provided by this library:

(**NOTE:** IBeam-corelib-go provides the convenient CreateServerWithConfig helper, that automatically handles these things for you! Please use it in devicecores)


- **SetCoreName:** Call this first setting the name of your core, eg "core-example"
- **SetDevMode:** Call this if you are running locally and not on skaarOS (check core-template how this can be done automatically)
- **Load** When the default config has been filled with default values pass a pointer to the structure to load. The library will automatically load the correct file, and also create it when necessary

PLEASE DO NOT MAKE OF THE **SAVE** function at the moment, talk to Lukas Bachschwell @s00500

Then create a config structure. Fieldnames become labels in the skaarOS webui. Use the **struct tags** if you like your field names to be different!

## Available struct Tags

* **Validation:** use `ibValidate:"ip"` to provide validations in the webUI, all possible validators are: `ip`, `port`, `password`, `devices`, `unique_inc` (unique autoincrent id, int)
* **Labels:** use `ibLabel:"My Field"` to provide a human readable label
* **Descriptions:** use `ibDescription:"Some description"` to provide a description for the current field
* **Options**: use `ibOptions:"Option1,Option2,Option3"` to provide a dropdown select with options, the field type needs to be **string**
* **Field Ordering**: use `ibOrder:"1"` to provide a integer value indicating a ordering of your fields used to sort the input form in the UI
* **Default Values in structured Arrays**: use `ibOrder:"myDefaultValue"` to provide a default value on fields inside of structure arrays. These values will be choosen when new elements are added to the structured array
* **Required Field** use `ibRequired:"Please specify the password generated by the camera"` to mark fields as required. The text in the tag will be shown as a red warning when the field stays empty. Keep in mind that you should NOT use this in cases where a default value can be assumed by the core.
* **Special Flags for Reactor**  To indicate certain files for reactor use: `ibDispatch:"devices"`, all possible options are currently: `devices`, `active`, `modelid` (creates model selector on cores), `deviceid`, `description`, `ip` `port`
* **Filter for Models**: use `ibOnlyOnModel` and `ibNotOnModel` with a comma seperated list of model ids to hide these fields in the UI. Keep in mind that due to older config entries there could still be values in these fields also for models that do not have them. This might cause confusion, so avoid parsing them on models that are not valid
* **Headline**: use `ibHeadline` to set a text that will be displayed above the field it's attached to. this also includes a separator line
* **Hidden Configuration**: use `ibHidden:"true"` to completely hide an element. this can be usefull to store data in the config structure and therefore in reactors project without directly showing it.

Create a default instance of your config structure. If it needs to be used on multiple go routines use a **sync.Mutex** to properly protect it (Always check your core with the race detector `go run --race .`)

## Other notes

To make the schema print in environments outside of skaarOS set `IBEAM_CONFIG_SCHEMA=.`

