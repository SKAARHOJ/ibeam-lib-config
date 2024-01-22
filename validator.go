package config

import (
	"fmt"

	cs "github.com/SKAARHOJ/ibeam-lib-config/configstructure"
	"github.com/oxequa/grace"
	log "github.com/s00500/env_logger"
)

// ValidateConfig validates a config structure against a schema, osed in different places to validate the correctness of configs.
// It also fixes the int types that get lost by json using float64 for everything. The result is returned as cleanedValues
// Pass strict mode to return errors when additional values are found in the config
func ValidateConfig(schema *cs.ValueTypeDescriptor, values interface{}, strictMode bool, nameForWarnings string) (cleanedValue interface{}, e error) {
	defer grace.Recover(&e)
	log := log.WithField("package", nameForWarnings)

	if schema == nil {
		if strictMode {
			return nil, fmt.Errorf("schema is not defined")
		}
		return values, nil
	}

	switch schema.Type {
	case cs.ValueType_Unknown:
		log.Warn("found unknown type in config!")

	case cs.ValueType_Integer:
		intVal, ok := intType(values)
		if !ok {
			return nil, fmt.Errorf("integer is no integertype, but %T", values)
		}
		values = intVal

	case cs.ValueType_Float:
		if _, ok := values.(float64); !ok {
			return nil, fmt.Errorf("float is no float, but %T", values)
		}

	case cs.ValueType_String:
		if _, ok := values.(string); !ok {
			return nil, fmt.Errorf("string is no string, but %T", values)
		}

	case cs.ValueType_Port:
		intVal, ok := intType(values)
		if !ok {
			return nil, fmt.Errorf("port is no integertype, but %T", values)
		}
		values = intVal

		if intVal < 0 || intVal > 65535 {
			return nil, fmt.Errorf("port out of range, is %d", intVal)
		}
	case cs.ValueType_IP:
		if _, ok := values.(string); !ok {
			return nil, fmt.Errorf("ip is no string, but %T", values)
		}

		// LB: We do not parse IPs here, as it could fail for hostnames!
		//if net.ParseIP(values.(string)) == nil {
		//	return nil, fmt.Errorf("could not parse ip %s", values.(string))
		//}

	case cs.ValueType_Checkbox:
		if _, ok := values.(bool); !ok {
			return nil, fmt.Errorf("bool is no bool, but %T", values)
		}

	case cs.ValueType_Structure:
		valueMap, ok := values.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("structure is no object")
		}

		for name, v := range valueMap {
			schemaValue, ok := schema.StructureSubtypes[name]
			if !ok {
				if strictMode {
					return nil, fmt.Errorf("value %s does not exist in schema", name)
				}
				log.Warnf("config validator: value %s does not exist in schema", name)
				continue
			}
			cleaned, err := ValidateConfig(schemaValue, v, strictMode, nameForWarnings)
			if err != nil {
				return nil, fmt.Errorf("on validating %s: %w", name, err)
			}
			valueMap[name] = cleaned
		}

	case cs.ValueType_Array:
		if values == nil {
			return values, nil
		}

		valueMap, ok := values.([]interface{})
		if !ok {
			return nil, fmt.Errorf("array is no array")
		}
		schemaValue := schema.ArraySubType

		for id, v := range valueMap {
			cleaned, err := ValidateConfig(schemaValue, v, strictMode, nameForWarnings)
			if err != nil {
				return nil, fmt.Errorf("on validating arrayvalue index %d: %w", id, err)
			}
			valueMap[id] = cleaned
		}

	case cs.ValueType_StructureArray:
		valueMap, ok := values.([]map[string]interface{})
		if !ok {
			if values == nil {
				return values, nil
			}
			// try to test again
			valueMap, ok := values.([]interface{})
			if !ok {
				return nil, fmt.Errorf("structured array is no array, but %T", values)
			}

			for id, structureValue := range valueMap {
				sv, ok := structureValue.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("(structure index %d) is no object", id)
				}
				for name, v := range sv {
					schemaValue, ok := schema.StructureSubtypes[name]
					if !ok {
						if strictMode {
							return nil, fmt.Errorf("(structure index %d) value %s does not exist in schema", id, name)
						}
						log.Warnf("(structure index %d) value %s does not exist in schema", id, name)
						continue
					}
					cleaned, err := ValidateConfig(schemaValue, v, strictMode, nameForWarnings)
					if err != nil {
						return nil, fmt.Errorf("on validating %s (structure index %d): %w", name, id, err)
					}
					valueMap[id].(map[string]interface{})[name] = cleaned
				}
			}

			return values, nil
		}

		for id, structureValue := range valueMap {
			for name, v := range structureValue {
				schemaValue, ok := schema.StructureSubtypes[name]
				if !ok {
					if strictMode {
						return nil, fmt.Errorf("(structure index %d) value %s does not exist in schema", id, name)
					}
					log.Warnf("(structure index %d) value %s does not exist in schema", id, name)
					continue
				}
				cleaned, err := ValidateConfig(schemaValue, v, strictMode, nameForWarnings)
				if err != nil {
					return nil, fmt.Errorf("on validating %s (structure index %d): %w", name, id, err)
				}
				valueMap[id][name] = cleaned
			}
		}

	case cs.ValueType_Password:
		if _, ok := values.(string); !ok {
			return nil, fmt.Errorf("password is no string, but %T", values)
		}

	case cs.ValueType_Select:
		if _, ok := values.(string); !ok {
			return nil, fmt.Errorf("select is no string, but %T", values)
		}

		if values.(string) != "" && !containsString(schema.Options, values.(string)) {
			return nil, fmt.Errorf("invalid select option %q", values.(string))
		}

	case cs.ValueType_UniqueInc:
		// TODO: it might make sense to actually check uniqueness, here at some point

		intVal, ok := intType(values)
		if !ok {
			return nil, fmt.Errorf("unique_integer is no integertype, but %T", values)
		}
		values = intVal
	}

	return values, nil
}

func intType(values interface{}) (int, bool) {
	if _, ok := values.(float64); ok {
		return int(values.(float64)), true
	}
	if _, ok := values.(int64); ok {
		return int(values.(int64)), true
	}
	if _, ok := values.(int); ok {
		return values.(int), true
	}
	return 0, false
}
