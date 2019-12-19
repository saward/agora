package runtime

import (
	"fmt"
	"io"
	"os"

	"github.com/saward/agora/bytecode"
)

type (
	// Error raised when a module ID is not found
	ModuleNotFoundError string
	// Error raised when a cyclic dependency is detected
	CyclicDependencyError string
)

// Error interface implementation.
func (e ModuleNotFoundError) Error() string {
	return string(e)
}

// Create a new ModuleNotFoundError.
func NewModuleNotFoundError(id string) ModuleNotFoundError {
	return ModuleNotFoundError(fmt.Sprintf("module not found: %s", id))
}

// Error interface implementation.
func (e CyclicDependencyError) Error() string {
	return string(e)
}

// Create a new CyclicDependencyError.
func NewCyclicDependencyError(id string) CyclicDependencyError {
	return CyclicDependencyError(fmt.Sprintf("cyclic dependency: %s already being loaded", id))
}

// The Compiler interface defines the required behaviour for a Compiler.
type Compiler interface {
	Compile(string, io.Reader) (*bytecode.File, error)
}

// A frame represents a currently executing function. A native function has no
// VM.
type frame struct {
	f   Func
	fvm *agoraFuncVM
}

// A Kontext represents the execution context. It is self-contained, share-nothing
// with other contexts. An execution context is *not* thread-safe, it should
// not be used concurrently. However, different instances of Kontext can be run
// concurrently, provided their components - Compiler, Resolver, etc. - are
// distinct instances too or do not rely on shared state or do so in a
// thread-safe way.
type Kontext struct {
	// Public fields
	Stdout     io.ReadWriter  // The standard streams
	Stdin      io.ReadWriter  // ...
	Stderr     io.ReadWriter  // ...
	Arithmetic Arithmetic     // The arithmetic processor
	Comparer   Comparer       // The comparison processor
	Resolver   ModuleResolver // The module loading resolver (match a module to a string literal)
	Compiler   Compiler       // The source code compiler
	Debug      bool           // Debug mode outputs helpful messages

	// Call stack
	frames []*frame
	frmsp  int

	// Modules management
	loadingMods map[string]bool // Modules currently being loaded
	loadedMods  map[string]Module
	builtin     Object
}

// NewKtx returns a new execution context, using the provided module resolver
// and compiler.
func NewKtx(resolver ModuleResolver, comp Compiler) *Kontext {
	c := &Kontext{
		Stdout:      os.Stdout,
		Stdin:       os.Stdin,
		Stderr:      os.Stderr,
		Arithmetic:  defaultArithmetic{},
		Comparer:    defaultComparer{},
		Resolver:    resolver,
		Compiler:    comp,
		loadingMods: make(map[string]bool),
		loadedMods:  make(map[string]Module),
	}
	// Automatically add the built-in functions
	b := new(builtinMod)
	b.SetKtx(c)
	if v, err := b.Run(); err != nil {
		panic("error loading agora builtin module: " + err.Error())
	} else {
		c.builtin = v.(Object)
	}
	return c
}

// Load resolves the module identified by the provided identifier, and loads
// it into memory. It returns a ready-to-run module, or an error.
//
// The sequence for loading, compiling, and bootstrapping execution is the
// following:
//
// * If id is empty string, return error.
// * If module is cached (ktx.loadedMods), return the Module, done.
// * If module is not cached, call ModuleResolver.Resolve(id string) (io.Reader, error)
// * If Resolve returns an error, return nil, error, done.
// * If file is already bytecode, just load it into memory using a decoder
// * If decoder returns an error, return nil, error, done.
// * Otherwise (if not bytecode) call Compiler.Compile(id string, r io.Reader) (*bytecode.File, error)
// * If Compile returns an error, return nil, error, done.
// * Create module from *bytecode.File
// * Cache module and return, do NOT execute the module.
//
func (c *Kontext) Load(id string) (Module, error) {
	if id == "" {
		return nil, NewModuleNotFoundError(id)
	}
	// If already loaded, return from cache
	if m, ok := c.loadedMods[id]; ok {
		return m, nil
	}
	// Else, resolve the matching file from the module id
	r, err := c.Resolver.Resolve(id)
	if err != nil {
		return nil, err
	}
	defer func() {
		if rc, ok := r.(io.ReadCloser); ok {
			rc.Close()
		}
	}()
	// If already bytecode, just decode
	var f *bytecode.File
	if rs, ok := r.(io.ReadSeeker); ok && bytecode.IsBytecode(rs) {
		dec := bytecode.NewDecoder(r)
		f, err = dec.Decode()
	} else {
		// Compile to bytecode
		f, err = c.Compiler.Compile(id, r)
	}
	if err != nil {
		return nil, err
	}
	mod := newAgoraModule(f, c)
	// cache and return
	c.loadedMods[id] = mod
	return mod, nil
}

// RegisterNativeModule adds the provided native module to the list of loaded and cached
// modules in this execution context (replacing any other module with the same ID).
func (c *Kontext) RegisterNativeModule(m NativeModule) {
	m.SetKtx(c)
	c.loadedMods[m.ID()] = m
}

// Mark the specified module as currently executing
func (c *Kontext) pushModule(id string) {
	if c.loadingMods[id] {
		panic(NewCyclicDependencyError(id))
	}
	c.loadingMods[id] = true
}

// Mark the specified module as no longer executing
func (c *Kontext) popModule(id string) {
	delete(c.loadingMods, id)
}

// Push a function onto the frame stack.
func (c *Kontext) pushFn(f Func, fvm *agoraFuncVM) {
	// Stack has to grow as needed
	if c.frmsp == len(c.frames) {
		if c.Debug && c.frmsp == cap(c.frames) {
			fmt.Fprintf(c.Stdout, "DEBUG expanding frames of ktx, current size: %d\n", len(c.frames))
		}
		c.frames = append(c.frames, &frame{f, fvm})
	} else {
		c.frames[c.frmsp] = &frame{f, fvm}
	}
	c.frmsp++
}

// Pop the top function from the frame stack.
func (c *Kontext) popFn() {
	c.frmsp--
	c.frames[c.frmsp] = nil // free this reference for gc
}

// IsRunning returns true if the specified function is currently executing.
func (c *Kontext) IsRunning(f Func) bool {
	for i := c.frmsp - 1; i >= 0; i-- {
		if c.frames[i].f == f {
			return true
		}
	}
	return false
}

// Get the variable identified by name, looking up the lexical scope stack and ultimately the
// built-ins.
func (c *Kontext) getVar(nm string, fvm *agoraFuncVM) (Val, bool) {
	// First look in locals
	if v, ok := fvm.vars[nm]; ok {
		return v, true
	}
	// Then recursively in parent environments
	for parent := fvm.val.env; parent != nil; parent = parent.parent {
		if v, ok := parent.upvals[nm]; ok {
			return v, true
		}
	}
	// Finally, look if the identifier refers to a built-in function.
	// This will return Nil if it doesn't match any built-in.
	b := c.builtin.Get(String(nm))
	return b, b != Nil
}

// Set the value of the variable identified by the provided name, looking up the
// frame stack if necessary. Returns true if the variable was found.
func (c *Kontext) setVar(nm string, v Val, fvm *agoraFuncVM) bool {
	// First attempt to set as local var
	if _, ok := fvm.vars[nm]; ok {
		fvm.vars[nm] = v
		return true
	}
	// Then recursively in parent environments
	for parent := fvm.val.env; parent != nil; parent = parent.parent {
		if _, ok := parent.upvals[nm]; ok {
			parent.upvals[nm] = v
			return true
		}
	}
	return false
}

// Pretty-print the execution context, up to n number of frames.
func (c *Kontext) dump(n int) {
	if n < 0 {
		return
	}
	for i, cnt := c.frmsp, c.frmsp-n; i > 0 && i > cnt; i-- {
		fmt.Fprintf(c.Stdout, "\n[Frame %3d]\n===========", i-1)
		if frm := c.frames[i-1]; frm.fvm != nil {
			fmt.Fprintln(c.Stdout, frm.fvm.dump())
		} else {
			fmt.Fprintln(c.Stdout, dumpVal(frm.f))
		}
	}
}
