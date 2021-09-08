package config

type ibeamDeviceConfig interface {
	mustEmbedBaseDevice()
}

// At some later point we might also use this structure as a template... but for atemproxy and others this migh not make sense
// type BaseCoreConfig struct {
//	Devices []ibeamDeviceConfig
// }

type BaseDeviceConfig struct {
	Active      bool   `ibOrder:"1" ibDispatch:"active" ibDescription:"disable connecting to the device"`
	DeviceID    uint32 `ibOrder:"2" ibDispatch:"deviceid" ibValidate:"unique_inc" ibDescription:"unique number identifier for this device"`
	ModelID     uint32 `ibOrder:"3" ibDispatch:"modelid" ibDescription:"the model type of the device"`
	Name        string `ibOrder:"4" ibDispatch:"name" ibDescription:"choose a name of your device"`
	Description string `ibOrder:"5" ibDispatch:"description"`
}

func (b BaseDeviceConfig) mustEmbedBaseDevice() {}
