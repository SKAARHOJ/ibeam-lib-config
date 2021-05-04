package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	cs "github.com/SKAARHOJ/ibeam-lib-config/configstructure"
	"github.com/pkg/errors"

	log "github.com/s00500/env_logger"

	"github.com/BurntSushi/toml"
)

var devMode bool = false
var path string = "/var/ibeam/config"
var coreName string = ""

func storeSchema(file string, structure interface{}) error {
	vptr := reflect.ValueOf(structure)
	v := vptr.Elem()

	csSchema := generateSchema(v.Type())

	if !devMode || true { // FIXME:
		jsonBytes, err := json.Marshal(&csSchema)
		log.MustFatal(errors.Wrap(err, "on encoding schema"))
		return ioutil.WriteFile(file, jsonBytes, 0644)
	}

	return nil
}

func generateSchema(v reflect.Type) *cs.ValueTypeDescriptor { // If fail: fatal
	return getTypeDescriptor(v, "", "", "", "")
}

func getTypeDescriptor(typeName reflect.Type, fieldName, validateTag, descriptionTag, optionsTag string) *cs.ValueTypeDescriptor {
	vtd := new(cs.ValueTypeDescriptor)
	vtd.Description = descriptionTag

	if optionsTag != "" {
		vtd.Options = strings.Split(optionsTag, ",")
	}

	log.Info("Field: ", fieldName, " Type: ", typeName.Name())
	if typeName.Kind() == reflect.Slice {
		sliceType := typeName.Elem() // Get the type of a single slice element

		if sliceType.Kind() == reflect.Ptr { // Pointer?
			sliceType = sliceType.Elem() // Then dereference it
		}

		if sliceType.Kind() == reflect.Struct {
			vtd.Type = cs.ValueType_StructureArray
			vtd.StructureSubtypes = make(map[string]*cs.ValueTypeDescriptor)
			for i := 0; i < sliceType.NumField(); i++ { // Iterate through all fields of the struct
				tag := sliceType.Field(i).Tag
				vtd.StructureSubtypes[sliceType.Field(i).Name] = getTypeDescriptor(sliceType.Field(i).Type, sliceType.Field(i).Name, tag.Get("ibValidate"), tag.Get("ibDescription"), tag.Get("ibOptions"))
			}
		} else {
			vtd.Type = cs.ValueType_Array
			vtd.ArraySubType = getTypeDescriptor(sliceType, fieldName, validateTag, descriptionTag, optionsTag)
		}
		return vtd
	} else if typeName.Kind() == reflect.Struct {
		vtd = structTypeDescriptor(typeName)
		vtd.Description = descriptionTag
		return vtd
	}

	if optionsTag != "" { // could check for string here
		vtd.Options = strings.Split(optionsTag, ",")
	}
	vtd.Type = getType(typeName.Name(), fieldName, validateTag, optionsTag)
	return vtd
}

func structTypeDescriptor(field reflect.Type) *cs.ValueTypeDescriptor {
	vtd := new(cs.ValueTypeDescriptor)
	vtd.Type = cs.ValueType_Structure
	vtd.StructureSubtypes = make(map[string]*cs.ValueTypeDescriptor)
	for i := 0; i < field.NumField(); i++ { // Iterate through all fields of the struct
		subField := field.Field(i)
		tag := subField.Tag
		vtd.StructureSubtypes[subField.Name] = getTypeDescriptor(subField.Type, subField.Name, tag.Get("ibValidate"), tag.Get("ibDescription"), tag.Get("ibOptions"))
	}

	return vtd
}

func getType(typeName, fieldName, validateTag, optionsTag string) cs.ValueType {
	switch typeName {
	case "string":
		if optionsTag != "" {
			return cs.ValueType_Select
		}
		switch validateTag {
		case "":
			return cs.ValueType_String
		case "ip":
			return cs.ValueType_IP
		case "password":
			return cs.ValueType_Password
		default:
			log.Fatal("Invalid validate '%s' tag on %s", validateTag, fieldName)
		}

	case "int", "uint32", "uint16", "uint64": // TODO: more int types ?
		switch validateTag {
		case "":
			return cs.ValueType_Integer
		case "port":
			return cs.ValueType_Port
		case "unique_inc":
			return cs.ValueType_UniqueInc
		default:
			log.Fatal("Invalid validate '%s' tag on %s", validateTag, fieldName)
		}
	default:
		log.Fatalf("Unknown type '%s' for config field  %s", typeName, fieldName)
	}
	log.Error("return 0")
	return 0
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
	err = storeSchema(baseFileName+".schema.json", structure)
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

/*
func contains(all []string, value string) bool {
	for _, one := range all {
		if one == value {
			return true
		}
	}
	return false
}
*/
