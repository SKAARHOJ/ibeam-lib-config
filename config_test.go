package config_test

import (
	"fmt"
	"net"
	"testing"

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
	conf.SetDevMode(true) // only use this in development
	conf.SetCoreName("core-template")
	err := conf.Load(&config)
	if err != nil {
		fmt.Println(err)
	}
}

func TestSave(t *testing.T) {
	type DeviceConfig struct {
		conf.BaseDeviceConfig
		IPAddress      string
		Port           uint16
		SomeConfigName string
		Stringoray     []string
	}

	type OtherStruct struct {
		Name    string
		Address int
	}

	type Config struct {
		Devices    []DeviceConfig
		OtherStuff []OtherStruct
	}

	// Define config filled with all defaults
	var config = Config{
		Devices: []DeviceConfig{
			{Port: 20},
			{Port: 20},
			{Port: 20},
		},
	}
	conf.SetDevMode(true) // only use this in development
	conf.SetCoreName("core-template")
	err := conf.Load(&config)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(config)
}

func TestLoad(t *testing.T) {
	// Needs to run after test save
	type DeviceConfig struct {
		conf.BaseDeviceConfig
		IPAddress      string
		Port           uint16
		SomeConfigName string
		Stringoray     []string
	}

	type OtherStruct struct {
		Name    string
		Address int
	}

	type Config struct {
		Devices    []DeviceConfig
		OtherStuff []OtherStruct
	}

	// Define config filled with all defaults
	var config = Config{
		Devices: []DeviceConfig{
			{Port: 20},
			{Port: 20},
			{Port: 20},
		},
	}
	conf.SetDevMode(true) // only use this in development
	conf.SetCoreName("core-template")
	err := conf.Load(&config)
	if err != nil {
		t.Error(err)
	}
}
