package introspection

// IntrospectionData is the root of the introspection response.
type IntrospectionData struct {
	Schema Schema `json:"__schema"`
}

// Schema represents a GraphQL schema.
type Schema struct {
	QueryType        *TypeName   `json:"queryType"`
	MutationType     *TypeName   `json:"mutationType"`
	SubscriptionType *TypeName   `json:"subscriptionType"`
	Types            []FullType  `json:"types"`
	Directives       []Directive `json:"directives"`
}

// TypeName is a simple type reference with just a name.
type TypeName struct {
	Name string `json:"name"`
}

// FullType represents a complete type definition.
type FullType struct {
	Kind          string       `json:"kind"`
	Name          string       `json:"name"`
	Description   string       `json:"description,omitempty"`
	Fields        []Field      `json:"fields,omitempty"`
	InputFields   []InputValue `json:"inputFields,omitempty"`
	Interfaces    []TypeRef    `json:"interfaces,omitempty"`
	EnumValues    []EnumValue  `json:"enumValues,omitempty"`
	PossibleTypes []TypeRef    `json:"possibleTypes,omitempty"`
}

// Field represents a field on a type.
type Field struct {
	Name              string       `json:"name"`
	Description       string       `json:"description,omitempty"`
	Args              []InputValue `json:"args,omitempty"`
	Type              TypeRef      `json:"type"`
	IsDeprecated      bool         `json:"isDeprecated"`
	DeprecationReason string       `json:"deprecationReason,omitempty"`
}

// InputValue represents an input value (argument or input field).
type InputValue struct {
	Name         string  `json:"name"`
	Description  string  `json:"description,omitempty"`
	Type         TypeRef `json:"type"`
	DefaultValue *string `json:"defaultValue,omitempty"`
}

// TypeRef is a reference to a type, potentially nested for NON_NULL and LIST.
type TypeRef struct {
	Kind   string   `json:"kind"`
	Name   string   `json:"name,omitempty"`
	OfType *TypeRef `json:"ofType,omitempty"`
}

// EnumValue represents a value in an enum.
type EnumValue struct {
	Name              string `json:"name"`
	Description       string `json:"description,omitempty"`
	IsDeprecated      bool   `json:"isDeprecated"`
	DeprecationReason string `json:"deprecationReason,omitempty"`
}

// Directive represents a GraphQL directive.
type Directive struct {
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Locations   []string     `json:"locations"`
	Args        []InputValue `json:"args,omitempty"`
}

// Type kinds
const (
	KindScalar      = "SCALAR"
	KindObject      = "OBJECT"
	KindInterface   = "INTERFACE"
	KindUnion       = "UNION"
	KindEnum        = "ENUM"
	KindInputObject = "INPUT_OBJECT"
	KindList        = "LIST"
	KindNonNull     = "NON_NULL"
)
