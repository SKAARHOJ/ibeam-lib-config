package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	cs "github.com/SKAARHOJ/ibeam-lib-config/configstructure"
	env "github.com/SKAARHOJ/ibeam-lib-env"

	log "github.com/s00500/env_logger"

	"github.com/BurntSushi/toml"
)

var devMode bool = false

const skaarOSpath string = "/var/ibeam/config"

var path string = skaarOSpath
var coreName string = ""

func init() {
	if env.IsDev() || env.IsProd() {
		devMode = true
		path = "" // In case we are not on skaarOS do not add the skaarOS path
	}
}

func GetConfigPath() string {
	return filepath.Join(path, coreName)
}

// Returns the current schema for a core
func GetSchema(structure interface{}) []byte {
	vptr := reflect.ValueOf(structure)
	v := vptr.Elem()

	csSchema := generateSchema(v.Type())

	jsonBytes, err := json.Marshal(&csSchema)
	log.Should(log.Wrap(err, "on getting schema"))
	return jsonBytes
}

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

	if !devMode {
		jsonBytes, err := json.Marshal(&csSchema)
		log.MustFatal(log.Wrap(err, "on encoding schema"))
		return ioutil.WriteFile(file, jsonBytes, 0644)
	}

	return nil
}

func generateSchema(v reflect.Type) *cs.ValueTypeDescriptor { // If fail: fatal
	return getTypeDescriptor(v, "", nil)
}

func getTypeDescriptor(typeName reflect.Type, fieldName string, parentTag *reflect.StructTag) *cs.ValueTypeDescriptor {
	var validateTag, descriptionTag, optionsTag, dispatchTag, orderTag, defaultTag, labelTag string
	if parentTag != nil {
		validateTag = parentTag.Get("ibValidate")
		descriptionTag = parentTag.Get("ibDescription")
		optionsTag = parentTag.Get("ibOptions")
		dispatchTag = parentTag.Get("ibDispatch")
		orderTag = parentTag.Get("ibOrder")
		defaultTag = parentTag.Get("ibDefault")
		labelTag = parentTag.Get("ibLabel")
	}

	vtd := new(cs.ValueTypeDescriptor)
	vtd.Description = descriptionTag
	vtd.Label = labelTag

	if dispatchTag != "" {
		vtd.DispatchOptions = strings.Split(dispatchTag, ",")
	}

	if orderTag != "" {
		orderNum, err := strconv.ParseInt(orderTag, 10, 32)
		log.ShouldWarn(log.Wrap(err, "on parsing order number for field %s", fieldName))
		vtd.Order = int(orderNum)
	} else {
		if containsString(vtd.DispatchOptions, "deviceip") {
			vtd.Order = 7 // Very special case, if IP is marked in core autofill the order parameter if not explicitly set
		}
	}

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
			if dispatchTag == "devices" || strings.ToLower(fieldName) == "devices" {
				var dcIface ibeamDeviceConfig
				if !sliceType.Implements(reflect.TypeOf(&dcIface).Elem()) {
					log.Fatal("Your deviceconfig array does not embedd config.BaseDeviceConfig, please add it and check for potential field duplications")
				}
			}

			vtd.Type = cs.ValueType_StructureArray
			vtd.StructureSubtypes = make(map[string]*cs.ValueTypeDescriptor)
			for i := 0; i < sliceType.NumField(); i++ { // Iterate through all fields of the struct
				tag := sliceType.Field(i).Tag
				if sliceType.Field(i).Type.Kind() == reflect.Struct && sliceType.Field(i).Anonymous {
					anoStructDescriptor := getTypeDescriptor(sliceType.Field(i).Type, sliceType.Field(i).Name, &tag)
					for name, typeDesc := range anoStructDescriptor.StructureSubtypes {
						if _, exists := vtd.StructureSubtypes[name]; exists {
							log.Fatalf("Potential struct Fieldname dupplication of field %s, ensure you have only one field with this name", name)
						}
						vtd.StructureSubtypes[name] = typeDesc
					}
					continue
				}

				if _, exists := vtd.StructureSubtypes[sliceType.Field(i).Name]; exists {
					log.Fatalf("Potential struct Fieldname dupplication of field %s, ensure you have only one field with this name", sliceType.Field(i).Name)
				}
				vtd.StructureSubtypes[sliceType.Field(i).Name] = getTypeDescriptor(sliceType.Field(i).Type, sliceType.Field(i).Name, &tag)
			}
		} else {
			//if dispatchTag != "" {
			//	log.Fatal("can not use dispatch tag on other fields than structured array")
			//}
			vtd.Type = cs.ValueType_Array
			vtd.ArraySubType = getTypeDescriptor(sliceType, fieldName, parentTag)
		}
		return vtd
	} else if typeName.Kind() == reflect.Struct {
		//if dispatchTag != "" {
		//	log.Fatal("can not use dispatch tag on other fields than structured array")
		//}
		vtd = structTypeDescriptor(typeName)
		vtd.Description = descriptionTag
		vtd.Label = labelTag
		return vtd
	}

	if optionsTag != "" { // could check for string here
		vtd.Options = strings.Split(optionsTag, ",")
	}
	vtd.Type, vtd.Default = getType(typeName.Name(), fieldName, validateTag, optionsTag, dispatchTag, defaultTag)
	return vtd
}

func structTypeDescriptor(field reflect.Type) *cs.ValueTypeDescriptor {
	vtd := new(cs.ValueTypeDescriptor)
	vtd.Type = cs.ValueType_Structure
	vtd.StructureSubtypes = make(map[string]*cs.ValueTypeDescriptor)
	for i := 0; i < field.NumField(); i++ { // Iterate through all fields of the struct
		subField := field.Field(i)

		tag := subField.Tag
		vtd.StructureSubtypes[subField.Name] = getTypeDescriptor(subField.Type, subField.Name, &tag)
	}

	return vtd
}

func getType(typeName, fieldName, validateTag, optionsTag, dispatchTag, defaultTag string) (vt cs.ValueType, defValue interface{}) {
	switch typeName {
	case "string":
		if defaultTag != "" {
			defValue = defaultTag
		}

		if optionsTag != "" {
			return cs.ValueType_Select, defValue
		}

		switch validateTag {
		case "":
			return cs.ValueType_String, defValue
		case "ip":
			return cs.ValueType_IP, defValue
		case "password":
			return cs.ValueType_Password, defValue
		default:
			log.Fatalf("Invalid validate '%s' tag on %s", validateTag, fieldName)
		}

	case "int", "int32", "int64", "uint32", "uint16", "uint64":
		if defaultTag != "" {
			defValue, _ = strconv.Atoi(defaultTag)
		}
		switch validateTag {
		case "":
			return cs.ValueType_Integer, defValue
		case "port":
			return cs.ValueType_Port, defValue
		case "unique_inc":
			return cs.ValueType_UniqueInc, nil // no use for a default here
		default:
			log.Fatalf("Invalid validate '%s' tag on %s", validateTag, fieldName)
		}
	case "bool":
		if defaultTag == "true" {
			defValue = true
		}
		return cs.ValueType_Checkbox, defValue
	case "float32", "float64":
		if defaultTag != "" {
			defValue, _ = strconv.ParseFloat(defaultTag, 32)
		}
		return cs.ValueType_Float, defValue
	default:
		log.Fatalf("Unknown type '%s' for config field  %s", typeName, fieldName)
	}
	log.Error("getType return 0")
	return 0, defValue
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
			err := os.MkdirAll(filepath.Join(path, coreName), 0700)
			log.Should(err)
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
		data, err = ioutil.ReadFile(baseFileName + ".toml")
		if err != nil {
			return log.Wrap(err, "on reading after saving")

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

	p := reflect.ValueOf(structure).Elem()
	p.Set(reflect.Zero(p.Type()))

	err = toml.Unmarshal(data, structure)
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

func containsString(all []string, one string) bool {
	for _, s := range all {
		if s == one {
			return true
		}
	}
	return false
}
