package config

type ibeamDeviceConfig interface {
	mustEmbedBaseDevice()
}

//type BaseCoreConfig struct {
//	Devices []ibeamDeviceConfig
//}

type BaseDeviceConfig struct {
	DeviceID    uint32 `ibDispatch:"deviceid" ibValidate:"unique_inc" ibDescription:"unique number identifier for this device"`
	ModelID     uint32 `ibDescription:"the model type of the device"`
	Active      bool   `ibDispatch:"active" ibDescription:"disable connecting to the device"`
	Name        string `ibDispatch:"name" ibDescription:"choose a name of your device"`
	Description string
}

func (b BaseDeviceConfig) mustEmbedBaseDevice() {}
