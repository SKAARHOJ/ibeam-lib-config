package config_test

import (
	"fmt"
	"net"

	conf "github.com/SKAARHOJ/ibeam-lib-config"
)

func ExampleLoad() {
	type GlobalConfig struct {
		FunnyString           string
		FunnyStringRay        []string
		FunnyStringRay_Double [][]string
		FunnyIntRayDOUBLE     [][]uint16
		FunnyBool             bool
	}

	type DeviceConfig struct {
		IPAddress      net.IP
		Port           uint16
		SomeConfigName string
		DeviceID       int
		Stringoray     []string
		Active         bool
	}

	type Config struct {
		Global  GlobalConfig
		Devices []DeviceConfig
	}

	// Define config filled with all defaults
	var config = Config{
		Global: GlobalConfig{
			FunnyStringRay: []string{"Hi people"},
			FunnyStringRay_Double: [][]string{
				{"Hi people"},
			},
		},
		Devices: []DeviceConfig{
			{Port: 20},
			{Port: 20},
			{Port: 20},
		},
	}

	conf.SetCoreName("core-template")
	err := conf.Load(&config)
	if err != nil {
		fmt.Println(err)
	}
}
