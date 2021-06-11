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

	log "github.com/s00500/env_logger"

	"github.com/BurntSushi/toml"
)

var devMode bool = false

const skaarOSpath string = "/var/ibeam/config"

var path string = skaarOSpath
var coreName string = ""

func storeSchema(file string, structure interface{}) error {
	vptr := reflect.ValueOf(structure)
	v := vptr.Elem()

	csSchema := generateSchema(v.Type())

	schemaPath := os.Getenv("IBEAM_CONFIG_SCHEMA")
	if schemaPath != "" {
		jsonBytes, err := json.Marshal(&csSchema)
		log.MustFatal(log.Wrap(err, "on encoding schema"))
		return ioutil.WriteFile(filepath.Join(schemaPath, coreName+".schema.json"), jsonBytes, 0644)
	}

	if devMode {
		jsonBytes, err := json.Marshal(&csSchema)
		log.MustFatal(log.Wrap(err, "on encoding schema"))
		return ioutil.WriteFile(file, jsonBytes, 0644)
	}

	return nil
}

func generateSchema(v reflect.Type) *cs.ValueTypeDescriptor { // If fail: fatal
	return getTypeDescriptor(v, "", "", "", "", "")
}

func getTypeDescriptor(typeName reflect.Type, fieldName, validateTag, descriptionTag, optionsTag, dispatchTag string) *cs.ValueTypeDescriptor {
	vtd := new(cs.ValueTypeDescriptor)
	vtd.Description = descriptionTag

	if optionsTag != "" {
		vtd.Options = strings.Split(optionsTag, ",")
	}

	if typeName.Kind() == reflect.Slice {
		sliceType := typeName.Elem() // Get the type of a single slice element

		if sliceType.Kind() == reflect.Ptr { // Pointer?
			sliceType = sliceType.Elem() // Then dereference it
		}

		if sliceType.Kind() == reflect.Struct {
			//if dispatchTag != "" && dispatchTag != "devices" {
			//	log.Fatal("can not use dispatch tag other than devices currently")
			//}
			vtd.DispatchOptions = strings.Split(dispatchTag, ",")
			vtd.Type = cs.ValueType_StructureArray
			vtd.StructureSubtypes = make(map[string]*cs.ValueTypeDescriptor)
			for i := 0; i < sliceType.NumField(); i++ { // Iterate through all fields of the struct
				tag := sliceType.Field(i).Tag
				vtd.StructureSubtypes[sliceType.Field(i).Name] = getTypeDescriptor(sliceType.Field(i).Type, sliceType.Field(i).Name, tag.Get("ibValidate"), tag.Get("ibDescription"), tag.Get("ibOptions"), tag.Get("ibDispatch"))
			}
		} else {
			//if dispatchTag != "" {
			//	log.Fatal("can not use dispatch tag on other fields than structured array")
			//}
			vtd.Type = cs.ValueType_Array
			vtd.ArraySubType = getTypeDescriptor(sliceType, fieldName, validateTag, descriptionTag, optionsTag, dispatchTag)
		}
		return vtd
	} else if typeName.Kind() == reflect.Struct {
		//if dispatchTag != "" {
		//	log.Fatal("can not use dispatch tag on other fields than structured array")
		//}
		vtd = structTypeDescriptor(typeName)
		vtd.Description = descriptionTag
		return vtd
	}

	if optionsTag != "" { // could check for string here
		vtd.Options = strings.Split(optionsTag, ",")
	}
	vtd.Type = getType(typeName.Name(), fieldName, validateTag, optionsTag, dispatchTag)
	return vtd
}

func structTypeDescriptor(field reflect.Type) *cs.ValueTypeDescriptor {
	vtd := new(cs.ValueTypeDescriptor)
	vtd.Type = cs.ValueType_Structure
	vtd.StructureSubtypes = make(map[string]*cs.ValueTypeDescriptor)
	for i := 0; i < field.NumField(); i++ { // Iterate through all fields of the struct
		subField := field.Field(i)
		tag := subField.Tag
		vtd.StructureSubtypes[subField.Name] = getTypeDescriptor(subField.Type, subField.Name, tag.Get("ibValidate"), tag.Get("ibDescription"), tag.Get("ibOptions"), tag.Get("ibDispatch"))
	}

	return vtd
}

func getType(typeName, fieldName, validateTag, optionsTag, dispatchTag string) cs.ValueType {
	if dispatchTag != "" {
		log.Fatal("can not use dispatch tag on other fields than structured array")
	}
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

	case "int", "int32", "int64", "uint32", "uint16", "uint64":
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
	case "bool":
		return cs.ValueType_Checkbox
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
		if saveErr := Save(structure); saveErr != nil {
			return saveErr
		}
	}

	err = storeSchema(baseFileName+".schema.json", structure)
	if err != nil {
		return fmt.Errorf("on storing schema: %w", err)
	}

	err = save(structure, coreName+".default")
	if err != nil {
		return fmt.Errorf("on storing : %w", err)
	}

	_, err = toml.Decode(string(data), structure)
	if err != nil {
		return fmt.Errorf("on decoding toml: %w", err)
	}
	return nil
}

// Save saves struct to toml
func Save(structure interface{}) error {
	if coreName == "" {
		log.Panic("no corename set")
	}
	return save(structure, coreName)
}

// save saves struct to toml
func save(structure interface{}, filename string) error {
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	err := enc.Encode(structure)
	if err != nil {
		return fmt.Errorf("on encoding toml: %w", err)
	}

	baseFileName := filepath.Join(path, coreName, filename)
	if devMode {
		baseFileName = filepath.Join(path, filename)
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
	} else {
		path = skaarOSpath
	}
}

// SetCoreName sets the name of the core and therefore the files
func SetCoreName(corename string) {
	coreName = corename
}
