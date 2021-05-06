package configstructure

import (
	"net"
)

// ValueType of config value
type ValueType = int

// ValueTypeEnum for using ValueType
const (
	ValueType_Unknown ValueType = iota
	ValueType_Integer
	ValueType_Float
	ValueType_String
	ValueType_Port
	ValueType_IP
	ValueType_Checkbox
	ValueType_Structure
	ValueType_Array
	ValueType_StructureArray
	ValueType_Password
	ValueType_Select
	ValueType_UniqueInc
)

/*
Schema: Tree of ValueTypeDescriptors

Config Tree of Value
*/

type ValueTypeDescriptor struct {
	Type        ValueType
	Description string   `json:",omitempty"`
	Options     []string `json:",omitempty"`

	ArraySubType      *ValueTypeDescriptor            `json:",omitempty"`
	StructureSubtypes map[string]*ValueTypeDescriptor `json:",omitempty"`
}

// Value structure
type Value struct {
	Type           ValueTypeDescriptor
	StringValue    string          `json:",omitempty"`
	PasswordValue  string          `json:",omitempty"`
	IntValue       int64           `json:",omitempty"`
	FloatValue     float64         `json:",omitempty"`
	IPValue        net.IP          `json:",omitempty"`
	PortValue      uint16          `json:",omitempty"`
	Checkbox       bool            `json:",omitempty"`
	SelectValue    string          `json:",omitempty"`
	Array          *Array          `json:",omitempty"`
	Structure      *Structure      `json:",omitempty"`
	StructureArray *StructureArray `json:",omitempty"`
}

// Array has a type of values and a value field
type Array struct {
	Values []Value
}

// StructureArray has a types map for values and a value field
type StructureArray struct {
	Values []map[string]*Value
}

// Structure a structure is a named map of values
type Structure struct {
	Value map[string]*Value
}
