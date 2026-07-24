package object
// Package object defines runtime values used by the interpreter.
//
// Participants will add object types for integers, booleans, and errors.
package object

// Object is the common interface for runtime values.
type Object interface {
	Type() ObjectType
	Inspect() string
}

// ObjectType identifies the kind of runtime value.
type ObjectType string

const (
	// IntegerObj represents integer values.
	IntegerObj ObjectType = "INTEGER"
)

// Integer stores an integer runtime value.
type Integer struct {
	Value int64
}

// Type returns the object type.
func (i *Integer) Type() ObjectType {
	panic("not implemented")
}

// Inspect returns a string description of the object.
func (i *Integer) Inspect() string {
	panic("not implemented")
}

// TODO: Add support for booleans, nil, and error values.
