package configstructure

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

type ValueTypeDescriptor struct {
	Type            ValueType
	Description     string   `json:",omitempty"`
	Options         []string `json:",omitempty"`
	Order           int      `json:",omitempty"`
	DispatchOptions []string `json:",omitempty"`

	ArraySubType      *ValueTypeDescriptor            `json:",omitempty"`
	StructureSubtypes map[string]*ValueTypeDescriptor `json:",omitempty"`
}
