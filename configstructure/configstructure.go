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
	ValueType_Description
	ValueType_Select
	ValueType_UniqueInc
)

type ValueTypeDescriptor struct {
	Type        ValueType
	Description string   `json:",omitempty"`
	Options     []string `json:",omitempty"`

	ArraySubType      *ValueTypeDescriptor            `json:",omitempty"`
	StructureSubtypes map[string]*ValueTypeDescriptor `json:",omitempty"`
}

// Value structure
type Value struct {
	Type           ValueType
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

	Description string   `json:",omitempty"`
	Options     []string `json:",omitempty"`
}

type StringValue struct {
	Type   ValueTypeDescriptor
	Values []Value
}

// Array has a type of values and a value field
type Array struct {
	Type   *ValueTypeDescriptor
	Values []Value
}

// StructureArray has a types map for values and a value field
type StructureArray struct {
	Types  map[string]*ValueTypeDescriptor
	Values []Structure
}

// Structure a structure is a named map of values
type Structure map[string]*Value
