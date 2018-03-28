package runtime

import (
	"fmt"

	"github.com/bobg/agora/bytecode"
)

// FuncFn represents the Func signature for native functions.
type FuncFn func(...Val) Val

// A Func value in Agora is a Val that also implements the Func interface.
type Func interface {
	Val
	Call(this Val, args ...Val) Val
}

// An agoraFuncDef represents an agora function's prototype.
type agoraFuncDef struct {
	ktx *Kontext
	mod *agoraModule
	// Internal fields filled by the compiler
	name    string
	stackSz int64
	expArgs int64
	kTable  []Val
	lTable  []string
	code    []bytecode.Instr
}

func newAgoraFuncDef(mod *agoraModule, c *Kontext) *agoraFuncDef {
	return &agoraFuncDef{
		ktx: c,
		mod: mod,
	}
}

// NewNativeFunc returns a native function initialized with the specified context,
// name and function implementation.
func NewNativeFunc(ktx *Kontext, nm string, fn FuncFn) *NativeFunc {
	return &NativeFunc{
		&funcVal{
			ktx,
			nm,
		},
		fn,
	}
}

// A NativeFunc represents a Go function exposed to agora.
type NativeFunc struct {
	// Expose the default Func value's behaviour
	*funcVal
	// Internal fields
	fn FuncFn
}

// ExpectAtLeastNArgs is a utility function for native modules implementation
// to ensure that the minimum number of arguments required are provided. It panics
// otherwise, which is the correct way to raise errors in the agora runtime.
func ExpectAtLeastNArgs(n int, args []Val) {
	if len(args) < n {
		panic(fmt.Sprintf("expected at least %d argument(s), got %d", n, len(args)))
	}
}

// Native returns the Go native representation of the native function type.
func (n *NativeFunc) Native() interface{} {
	return n
}

// Call executes the native function and returns its return value.
func (n *NativeFunc) Call(_ Val, args ...Val) Val {
	n.ktx.pushFn(n, nil)
	defer n.ktx.popFn()
	return n.fn(args...)
}
