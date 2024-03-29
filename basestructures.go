package config

type ibeamDeviceConfig interface {
	mustEmbedBaseDevice()
}
type BaseDeviceConfig struct {
	Active      bool   `ibOrder:"1" ibDispatch:"active"`
	Name        string `ibOrder:"5" ibDispatch:"name" ibDescription:""`
	DeviceID    uint32 `ibLabel:"Device Id" ibOrder:"10" ibDispatch:"deviceid" ibValidate:"unique_inc" ibDescription:"Unique number identifier for this device"`
	ModelID     uint32 `ibLabel:"Model Id" ibOrder:"15" ibDispatch:"modelid" ibDescription:"The model type of the device"`
	Description string `ibOrder:"20" ibDispatch:"description"`
}

func (b BaseDeviceConfig) mustEmbedBaseDevice() {}
