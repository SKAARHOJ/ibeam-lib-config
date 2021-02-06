package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
)

var devMode bool = false
var path string = "/var/ibeam/config"
var coreName string = ""

func storeSchema(file string, structure interface{}) error {
	vptr := reflect.ValueOf(structure)
	v := vptr.Elem()

	extendStruct(v)

	var schemaBuffer bytes.Buffer
	e := toml.NewEncoder(&schemaBuffer)
	err := e.Encode(structure)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, schemaBuffer.Bytes(), 0644)
}

func extendStruct(v reflect.Value) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem() // Then dereference it
	}
	for i := 0; i < v.NumField(); i++ {
		switch v.Field(i).Kind() {
		case reflect.Slice:
			extendSlice(v.Field(i))
		case reflect.Struct:
			extendStruct(v.Field(i))
		case reflect.Map:
			panic("skaarOS does not yet support maps in configs!")
			// Current Rules for config:
			// No pointers
			// no maps
		}
	}
}

func extendSlice(s reflect.Value) {
	// Problem: certain custom slice types like net.IP caue this to fail.... so we need to skip them
	if s.Type() == reflect.TypeOf(net.IP{}) {
		return
	}

	st := s.Type()
	sliceType := st.Elem() // Get the type of a single slice element

	if sliceType.Kind() == reflect.Ptr { // Pointer?
		sliceType = sliceType.Elem() // Then dereference it
	}

	newitem := reflect.New(sliceType)
	if sliceType.Kind() == reflect.Struct {
		extendStruct(newitem)
	} else if sliceType.Kind() == reflect.Slice {
		for i := 0; i < s.Len(); i++ {
			extendSlice(s.Index(i)) // extend inner slice values
		}
		extendSlice(newitem.Elem())
	}
	s.Set(reflect.Append(s, newitem.Elem()))
}

// Load a package config, also storing the default config and schema for ibeam-init to pick up
func Load(structure interface{}) error {
	if coreName == "" {
		return fmt.Errorf("No core name set")
	}
	// then it checks if the config exists, if not store default config
	// Then load config

	data, err := ioutil.ReadFile(filepath.Join(path, coreName) + ".toml")
	if err != nil {
		// There is a chance that file we are looking for
		// just doesn't exist. In this case we are supposed
		// to create an empty configuration file, based on v.
		// FIXME: LB: this inner error thing is ugly, should check file errors proeperly and handle
		if innerErr := save(structure); innerErr != nil {
			// Smth going on with the file system... returning error.
			return err
		}
	}

	// This function generates and stores a schema (= default config plus at least one of each type)
	if !devMode {
		err = storeSchema(strings.TrimSuffix(filepath.Join(path, coreName))+".schema.toml", structure)
		if err != nil {
			return fmt.Errorf("on storing schema: %w", err)
		}
	}

	_, err = toml.Decode(string(data), structure)
	if err != nil {
		return fmt.Errorf("on decoding toml: %w", err)
	}
	return nil
}

// save saves struct to toml
func save(structure interface{}) error {
	if coreName == "" {
		return fmt.Errorf("No core name set")
	}
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	err := enc.Encode(structure)
	if err != nil {
		return fmt.Errorf("on encoding toml: %w", err)
	}

	err = ioutil.WriteFile(filepath.Join(path, coreName)+".toml", buf.Bytes(), os.ModePerm)
	if err != nil {
		return fmt.Errorf("on encoding toml: %w", err)
	}

	return nil
}

func SetDevMode(devmode bool) {
	devMode = devmode
	if devMode {
		path = ""
	}
}

func SetCoreName(corename string) {
	coreName = corename
}
