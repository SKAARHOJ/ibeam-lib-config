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
	"unicode"

	log "github.com/s00500/env_logger"

	"github.com/BurntSushi/toml"
)

var devMode bool = false
var path string = "/var/ibeam/config"
var coreName string = ""

func storeSchema(file string, structure interface{}) error {
	vptr := reflect.ValueOf(structure)
	v := vptr.Elem()

	extendStruct(v)

	if !devMode {
		var schemaBuffer bytes.Buffer
		e := toml.NewEncoder(&schemaBuffer)
		err := e.Encode(structure)
		if err != nil {
			return err
		}
		return ioutil.WriteFile(file, schemaBuffer.Bytes(), 0644)
	}

	return nil
}

func extendStruct(v reflect.Value) {

	if v.Kind() == reflect.Ptr {
		v = v.Elem() // Then dereference it
	}

	tomlTagList := make([]string, 0)

	for i := 0; i < v.NumField(); i++ {
		tomlTag := v.Type().Field(i).Tag.Get("toml")

		if tomlTag == "" {
			tomlTag = v.Type().Field(i).Name
		}

		tomlTagList = append(tomlTagList, tomlTag)

		if strings.HasSuffix(tomlTag, "_description") {
			if v.Type().Field(i).Type.Name() != "string" {
				log.Fatalf("descriptions must be strings! (failed on %s)", tomlTag)
			}
		}

		if strings.HasSuffix(tomlTag, "_options") {
			if v.Type().Field(i).Type.Name() != "[]string" {
				log.Fatalf("options must be string slices (failed on %s)", tomlTag)
			}
		}

		if strings.HasSuffix(tomlTag, "_select") {
			if v.Type().Field(i).Type.Name() != "string" {
				log.Fatalf("selects must be strings! (failed on %s)", tomlTag)
			}
		}

		name := []rune(v.Type().Field(i).Name)
		isExported := unicode.IsUpper(name[0])

		switch v.Field(i).Kind() {
		case reflect.Slice:
			if !isExported {
				log.Fatal("Make sure included slices are exported in the main struct (start with uppercase)")
			}
			extendSlice(v.Field(i))
		case reflect.Struct:
			extendStruct(v.Field(i))
		case reflect.Map:
			log.Panic("skaarOS config does not support maps!")
			// Current Rules for config:
			// No pointers
			// no maps
		}
	}

	validate(tomlTagList)

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
		log.Panic("no corename set")
	}
	// then it checks if the config exists, if not store default config
	// Then load config

	// check for the config dir, create if it does not exist
	if !devMode {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			os.Mkdir(filepath.Join(path, coreName, coreName), 0700)
		}
	}

	baseFileName := filepath.Join(path, coreName, coreName)
	if devMode {
		baseFileName = filepath.Join(path, coreName)
	}

	data, err := ioutil.ReadFile(baseFileName + ".toml")
	if err != nil {
		// There is a chance that file we are looking for
		// just doesn't exist. In this case we are supposed
		// to create an empty configuration file, based on v.
		if saveErr := save(structure); saveErr != nil {
			return saveErr
		}
	}

	// This function generates and stores a schema (= default config plus at least one of each type)
	err = storeSchema(baseFileName+".schema.toml", structure)
	if err != nil {
		return fmt.Errorf("on storing schema: %w", err)
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
		log.Panic("no corename set")
	}
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	err := enc.Encode(structure)
	if err != nil {
		return fmt.Errorf("on encoding toml: %w", err)
	}

	baseFileName := filepath.Join(path, coreName, coreName)
	if devMode {
		baseFileName = filepath.Join(path, coreName)
	}

	err = ioutil.WriteFile(baseFileName+".toml", buf.Bytes(), os.ModePerm)
	if err != nil {
		return fmt.Errorf("on encoding toml: %w", err)
	}

	return nil
}

// SetDevMode activates the development mode path configuration
func SetDevMode(devmode bool) {
	devMode = devmode
	if devMode {
		path = ""
	}
}

// SetCoreName sets the name of the core and therefore the files
func SetCoreName(corename string) {
	coreName = corename
}

func validate(tags []string) {
	// check that each description has a normal tag
	// check that each select has options

	for _, tag := range tags {
		if strings.HasSuffix(tag, "_description") {
			if !contains(tags, strings.TrimSuffix(tag, "_description")) {
				log.Fatal("Did not find field for description ", tag)
			}
		}

		if strings.HasSuffix(tag, "_select") {
			if !contains(tags, strings.TrimSuffix(tag, "_select")+"_options") {
				log.Fatal("Did not find options for select ", tag)
			}
		}
	}
}

func contains(all []string, value string) bool {
	for _, one := range all {
		if one == value {
			return true
		}
	}
	return false
}
