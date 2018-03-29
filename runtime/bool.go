package runtime

import (
	"context"
	"fmt"
)

// Bool is the representation of the Boolean type. It is equivalent
// to Go's bool type.
type Bool bool

// Dump pretty-prints the value for debugging purpose.
func (b Bool) Dump() string {
	return fmt.Sprintf("%v (Bool)", bool(b))
}

// Int returns 1 if true, 0 if false.
func (b Bool) Int(context.Context) int64 {
	if bool(b) {
		return 1
	}
	return 0
}

// Float returns 1 if true, 0 if false.
func (b Bool) Float(context.Context) float64 {
	if bool(b) {
		return 1.0
	}
	return 0.0
}

// String returns "true" if true, "false" otherwise.
func (b Bool) String(context.Context) string {
	if bool(b) {
		return "true"
	}
	return "false"
}

// Bool returns the boolean value itself.
func (b Bool) Bool(context.Context) bool {
	return bool(b)
}

// Native returns the bool native Go representation.
func (b Bool) Native(context.Context) interface{} {
	return bool(b)
}
