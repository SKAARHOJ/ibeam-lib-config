package config

type ibeamDeviceConfig interface {
	mustEmbedBaseDevice()
}

// At some later point we might also use this structure as a template... but for atemproxy and others this migh not make sense
// type BaseCoreConfig struct {
//	Devices []ibeamDeviceConfig
// }

type BaseDeviceConfig struct {
	Active      bool   `ibOrder:"1" ibDispatch:"active"`
	Name        string `ibOrder:"5" ibDispatch:"name" ibDescription:""`
	DeviceID    uint32 `ibOrder:"10" ibDispatch:"deviceid" ibValidate:"unique_inc" ibDescription:"Unique number identifier for this device"`
	ModelID     uint32 `ibOrder:"15" ibDispatch:"modelid" ibDescription:"The model type of the device"`
	Description string `ibOrder:"20" ibDispatch:"description"`
}

func (b BaseDeviceConfig) mustEmbedBaseDevice() {}
