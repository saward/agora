// Package emitter generates the bytecode instructions by traversing the abstract syntax tree
// generated by the parser.
package emitter

import (
	"errors"
	"strconv"
	"strings"

	"github.com/saward/agora/bytecode"
	"github.com/saward/agora/compiler/parser"
)

var (
	binSym2op = map[string]bytecode.Opcode{
		"+":  bytecode.OP_ADD,
		"-":  bytecode.OP_SUB,
		"*":  bytecode.OP_MUL,
		"/":  bytecode.OP_DIV,
		"%":  bytecode.OP_MOD,
		"<":  bytecode.OP_LT,
		"<=": bytecode.OP_LTE,
		">":  bytecode.OP_GT,
		">=": bytecode.OP_GTE,
		"==": bytecode.OP_EQ,
		"!=": bytecode.OP_NEQ,
	}
	binAsgSym2op = map[string]bytecode.Opcode{
		"+=": bytecode.OP_ADD,
		"-=": bytecode.OP_SUB,
		"*=": bytecode.OP_MUL,
		"/=": bytecode.OP_DIV,
		"%=": bytecode.OP_MOD,
	}
	unrSym2op = map[string]bytecode.Opcode{
		"++": bytecode.OP_ADD,
		"--": bytecode.OP_SUB,
		"!":  bytecode.OP_NOT,
		"-":  bytecode.OP_UNM,
	}
)

type forData struct {
	breaks []int
	conts  []int
}

type kId struct {
	v string
	t bytecode.KType
}

type asgType int

const (
	atTrue asgType = iota
	atFalse
	atDefine
)

// An Emitter is responsible for generating the instructions for an agora program.
type Emitter struct {
	err     error
	kMap    map[*bytecode.Fn]map[kId]int
	stackSz map[*bytecode.Fn]int64
	forNest map[*bytecode.Fn][]*forData
	fnIx    []int64
}

// Emit takes a module identifier, the symbols generated by the parser (the headless *AST*),
// and the root scope, and emits the instructions required to execute the program.
// It returns the in-memory bytecode representation of the program. If an error is
// encountered, it is returned as second value, otherwise it returns nil.
func (e *Emitter) Emit(id string, syms []*parser.Symbol, scps *parser.Scope) (*bytecode.File, error) {
	// Reset the internal fields
	e.err = nil
	e.kMap = make(map[*bytecode.Fn]map[kId]int)
	e.stackSz = make(map[*bytecode.Fn]int64)
	e.forNest = make(map[*bytecode.Fn][]*forData)

	// Create the bytecode representation structure
	f := bytecode.NewFile(id)
	fn := new(bytecode.Fn)
	fn.Header.Name = f.Name // Expected args and parent func are always 0 for top-level func
	// TODO : Line start and end
	f.Fns = append(f.Fns, fn)
	e.fnIx = []int64{0}
	e.emitBlock(f, fn, syms)
	return f, e.err
}

func (e *Emitter) emitFn(f *bytecode.File, sym *parser.Symbol) {
	if e.err != nil {
		return
	}
	e.assert(sym.Ar == parser.ArFunction, errors.New("expected `"+sym.Id+"` to have function arity"))
	fn := new(bytecode.Fn)
	fn.Header.Name = sym.Name
	args := sym.First.([]*parser.Symbol)
	fn.Header.ExpArgs = int64(len(args))
	fn.Header.ParentFnIx = e.fnIx[len(e.fnIx)-1]
	// TODO : Line Start, Line End
	f.Fns = append(f.Fns, fn)
	e.fnIx = append(e.fnIx, int64(len(f.Fns)-1))
	// Define the expected args in the K table - *MUST* be defined in spots 0..ExpArgs - 1
	for _, arg := range args {
		e.assert(arg.Ar == parser.ArName, errors.New("expected argument to have name arity"))
		e.registerK(fn, arg.Val, true, true)
	}
	stmts := sym.Second.([]*parser.Symbol)
	e.emitBlock(f, fn, stmts)
	// Cleanup map keys of this fn
	e.fnIx = e.fnIx[:len(e.fnIx)-1]
	delete(e.kMap, fn)
	delete(e.stackSz, fn)
	delete(e.forNest, fn)
}

func (e *Emitter) emitAny(f *bytecode.File, fn *bytecode.Fn, sym *parser.Symbol, any interface{}) {
	switch v := any.(type) {
	case *parser.Symbol:
		e.emitSymbol(f, fn, v, atFalse)
	case []*parser.Symbol:
		e.emitBlock(f, fn, v)
	default:
		e.assert(false, errors.New("expected branch of `"+sym.Id+"` to be a symbol or a slice of symbols"))
	}
}

func (e *Emitter) emitBlock(f *bytecode.File, fn *bytecode.Fn, syms []*parser.Symbol) {
	for _, sym := range syms {
		e.emitSymbol(f, fn, sym, atFalse)
	}
}

func (e *Emitter) emitShortcutIf(f *bytecode.File, fn *bytecode.Fn, parent *parser.Symbol, cond, truePart, falsePart interface{}) {
	// Emit the condition
	e.emitAny(f, fn, parent, cond)
	// Next comes the TEST
	tstIx := e.addTempInstr(fn)
	// Then the true expression
	e.emitAny(f, fn, parent, truePart)
	// Then a jump over the false expression
	jmpIx := e.addTempInstr(fn)
	// Update the test instruction, here starts the false part
	e.updateTestInstr(fn, tstIx)
	// Emit the false expression
	e.emitAny(f, fn, parent, falsePart)
	// Update the jump instruction, to after the false part
	e.updateJumpfInstr(fn, jmpIx)
}

func (e *Emitter) emitSymbol(f *bytecode.File, fn *bytecode.Fn, sym *parser.Symbol, asg asgType) {
	if e.err != nil {
		return
	}
	switch sym.Id {
	case "nil":
		e.assert(asg == atFalse, errors.New("invalid assignment to nil"))
		e.addInstr(fn, bytecode.OP_PUSH, bytecode.FLG_N, 0)
	case "(name)", "import", "panic", "recover", "len", "keys", "string", "number",
		"bool", "type", "status", "reset": // TODO : Cleaner way to handle all builtins
		// Register the symbol, may or may not be a local
		e.assert(sym.Ar == parser.ArName || sym.Ar == parser.ArLiteral, errors.New("expected `"+sym.Id+"` to have name or literal arity"))
		kix := e.registerK(fn, sym.Val, true, asg == atDefine)
		if asg != atFalse {
			e.addInstr(fn, bytecode.OP_POP, bytecode.FLG_V, kix)
		} else if sym.Ar == parser.ArLiteral {
			e.addInstr(fn, bytecode.OP_PUSH, bytecode.FLG_K, kix)
		} else {
			e.addInstr(fn, bytecode.OP_PUSH, bytecode.FLG_V, kix)
		}
	case "(literal)", "true", "false":
		// Register the symbol
		e.assert(asg == atFalse, errors.New("invalid assignment to a literal"))
		e.assert(sym.Ar == parser.ArLiteral, errors.New("expected `"+sym.Id+"` to have literal arity"))
		kix := e.registerK(fn, sym.Val, false, false)
		e.addInstr(fn, bytecode.OP_PUSH, bytecode.FLG_K, kix)
	case "this":
		e.assert(asg == atFalse, errors.New("invalid assignment to the `this` keyword"))
		e.addInstr(fn, bytecode.OP_PUSH, bytecode.FLG_T, 0)
	case "args":
		e.assert(asg == atFalse, errors.New("invalid assignment to the `args` keyword"))
		e.addInstr(fn, bytecode.OP_PUSH, bytecode.FLG_A, 0)
	case ".", "[":
		e.assert(sym.Ar == parser.ArBinary, errors.New("expected `"+sym.Id+"` to have binary arity"))
		e.emitSymbol(f, fn, sym.Second.(*parser.Symbol), atFalse)
		e.emitSymbol(f, fn, sym.First.(*parser.Symbol), atFalse)
		if asg != atFalse {
			e.addInstr(fn, bytecode.OP_SFLD, bytecode.FLG__, 0)
		} else {
			e.addInstr(fn, bytecode.OP_GFLD, bytecode.FLG__, 0)
		}
	case ":=":
		e.assert(sym.Ar == parser.ArBinary, errors.New("expected `:=` to have binary arity"))
		e.emitSymbol(f, fn, sym.Second.(*parser.Symbol), atFalse)
		e.emitSymbol(f, fn, sym.First.(*parser.Symbol), atDefine)
	case "!":
		e.assert(sym.Ar == parser.ArUnary, errors.New("expected `!` to have unary arity"))
		e.emitSymbol(f, fn, sym.First.(*parser.Symbol), atFalse)
		e.addInstr(fn, unrSym2op[sym.Id], bytecode.FLG__, 0)
	case "-":
		if sym.Ar == parser.ArUnary {
			e.emitSymbol(f, fn, sym.First.(*parser.Symbol), atFalse)
			e.addInstr(fn, unrSym2op[sym.Id], bytecode.FLG__, 0)
			break
		}
		fallthrough
	case "+", "*", "/", "%", "<", ">", "<=", ">=", "==", "!=":
		e.assert(sym.Ar == parser.ArBinary, errors.New("expected `"+sym.Id+"` to have binary arity"))
		e.emitSymbol(f, fn, sym.First.(*parser.Symbol), atFalse)
		e.emitSymbol(f, fn, sym.Second.(*parser.Symbol), atFalse)
		e.addInstr(fn, binSym2op[sym.Id], bytecode.FLG__, 0)
	case "&&", "||":
		e.assert(sym.Ar == parser.ArBinary, errors.New("expected `"+sym.Id+"` to have binary arity"))
		if sym.Id == "&&" {
			// Equivalent to if <first> then <second> else <first>
			e.emitShortcutIf(f, fn, sym, sym.First, sym.Second, sym.First)
		} else {
			// Equivalent to if <first> then <first> else <second>
			e.emitShortcutIf(f, fn, sym, sym.First, sym.First, sym.Second)
		}
	case "=":
		e.assert(sym.Ar == parser.ArBinary, errors.New("expected `=` to have binary arity"))
		e.emitSymbol(f, fn, sym.Second.(*parser.Symbol), atFalse)
		left := sym.First.(*parser.Symbol)
		if left.Id == "." {
			// Emit left, which will generate a SFLD
			e.emitSymbol(f, fn, left, atTrue)
		} else {
			// Emit a standard POP instruction
			e.emitSymbol(f, fn, left, atTrue)
		}
	case "+=", "-=", "*=", "/=", "%=":
		e.assert(sym.Ar == parser.ArBinary, errors.New("expected `"+sym.Id+"` to have binary arity"))
		e.emitSymbol(f, fn, sym.First.(*parser.Symbol), atFalse)
		e.emitSymbol(f, fn, sym.Second.(*parser.Symbol), atFalse)
		e.addInstr(fn, binAsgSym2op[sym.Id], bytecode.FLG__, 0)
		e.emitSymbol(f, fn, sym.First.(*parser.Symbol), atTrue)
	case "++", "--":
		e.assert(sym.Ar == parser.ArStatement, errors.New("expected `"+sym.Id+"` to have statement arity"))
		e.emitSymbol(f, fn, sym.First.(*parser.Symbol), atFalse)
		// Implicit `1` constant
		ix := e.registerK(fn, "1", false, false)
		e.addInstr(fn, bytecode.OP_PUSH, bytecode.FLG_K, ix)
		e.addInstr(fn, unrSym2op[sym.Id], bytecode.FLG__, 0)
		e.emitSymbol(f, fn, sym.First.(*parser.Symbol), atTrue)
	case "func":
		funcIx := len(f.Fns) // New Fn will be added at this index
		if sym.Name != "" {
			// Function defined as a statement, register the name as a K,
			// and push the function's value into this variable.
			kix := e.registerK(fn, sym.Name, true, true)
			e.addInstr(fn, bytecode.OP_PUSH, bytecode.FLG_F, uint64(funcIx))
			e.addInstr(fn, bytecode.OP_POP, bytecode.FLG_V, kix)
		}
		e.emitFn(f, sym)
		if sym.Name == "" {
			// Func defined as an expression, must be pushed on the stack
			e.addInstr(fn, bytecode.OP_PUSH, bytecode.FLG_F, uint64(funcIx))
		}
	case "(":
		e.assert(sym.Ar == parser.ArBinary || sym.Ar == parser.ArTernary, errors.New("expected `(` to have binary or ternary arity"))
		// Push parameters
		var parms []*parser.Symbol
		var op bytecode.Opcode
		if sym.Ar == parser.ArBinary {
			parms = sym.Second.([]*parser.Symbol)
			op = bytecode.OP_CALL
		} else {
			parms = sym.Third.([]*parser.Symbol)
			op = bytecode.OP_CFLD
		}
		for _, parm := range parms {
			e.emitSymbol(f, fn, parm, atFalse)
		}
		// If ternary, push field (Second)
		if sym.Ar == parser.ArTernary {
			e.emitSymbol(f, fn, sym.Second.(*parser.Symbol), atFalse)
		}
		// Push function name (or parent object of the field if ternary)
		e.emitSymbol(f, fn, sym.First.(*parser.Symbol), atFalse)
		// Call
		e.addInstr(fn, op, bytecode.FLG_An, uint64(len(parms)))
	case "{":
		e.assert(sym.Ar == parser.ArUnary, errors.New("expected `{` to have unary arity"))
		ln := 0
		if !e.isEmpty(sym.First) {
			e.emitAny(f, fn, sym, sym.First)
			if ar, ok := sym.First.([]*parser.Symbol); ok {
				ln = len(ar)
			}
		}
		e.addInstr(fn, bytecode.OP_NEW, bytecode.FLG__, uint64(ln))
	case "?":
		// Similar to if, but yields a value
		e.assert(sym.Ar == parser.ArTernary, errors.New("expected `?` to have ternary arity"))
		e.emitShortcutIf(f, fn, sym, sym.First, sym.Second, sym.Third)
	case "if":
		e.assert(sym.Ar == parser.ArStatement, errors.New("expected `if` to have statement arity"))
		// First is the condition, always a *Symbol
		e.emitSymbol(f, fn, sym.First.(*parser.Symbol), atFalse)
		// Next comes the TEST, but we don't know yet how many instructions to jump
		// insert a placeholder (invalid op) so that it fails explicitly should it ever make it to
		// the VM.
		tstIx := e.addTempInstr(fn)
		// Then comes the body
		e.emitBlock(f, fn, sym.Second.([]*parser.Symbol))
		// Update the test instruction, now that we know where to jump to
		e.updateTestInstr(fn, tstIx)
		// Then comes the ELSE/ELSE IF, maybe
		if sym.Third != nil {
			// If so, insert a jump over the else part
			jmpIx := e.addTempInstr(fn)
			// And re-update the test instruction, since an instr was added
			e.updateTestInstr(fn, tstIx)
			// Emit the else or else-if part
			e.emitAny(f, fn, sym, sym.Third)
			// Update the jump instruction now that we know how many instrs to jump over
			e.updateJumpfInstr(fn, jmpIx)
		}
	case "forr":
		// For-range notation, distinct from regular for
		// First must be either a `=` or a `:=`
		assign := sym.First.(*parser.Symbol)
		e.assert(assign.Id == "=" || assign.Id == ":=", errors.New("left hand side of `for...range` must be `=` or `:=`"))
		rng := assign.Second.(*parser.Symbol)
		e.assert(rng.Id == "range", errors.New("right hand side of `for...range` must be the `range` keyword"))
		// Push `range` args onto the stack
		args := rng.First.([]*parser.Symbol)
		e.emitBlock(f, fn, args)
		// Start the `range` coroutine
		e.addInstr(fn, bytecode.OP_RNGS, bytecode.FLG_An, uint64(len(args)))
		// For loop officially starts here
		start := len(fn.Is)
		// Push one value from the coro (until multiple vals are supported), + condition
		e.addInstr(fn, bytecode.OP_RNGP, bytecode.FLG_An, 1)
		// Test the end of loop
		tstIx := e.addTempInstr(fn)
		// Pop the top value from the stack into the iteration var
		if assign.Id == "=" {
			e.emitSymbol(f, fn, assign.First.(*parser.Symbol), atTrue)
		} else {
			e.emitSymbol(f, fn, assign.First.(*parser.Symbol), atDefine)
		}
		// Emit the body
		e.startFor(fn)
		e.emitAny(f, fn, sym, sym.Second)
		// Update the continue statements (must jump to the next statement)
		e.updateForJmp(fn, false)
		// Add the jump back to RNGP instruction
		e.addInstr(fn, bytecode.OP_JMP, bytecode.FLG_Jb, uint64(len(fn.Is)-start))
		// Break statements must jump to the next statement (RNGE)
		e.updateForJmp(fn, true)
		// Update the test instruction to jump to the next statement (RNGE)
		e.updateTestInstr(fn, tstIx)
		e.endFor(fn)
		// Emit the range end (clear coroutine) statement
		e.addInstr(fn, bytecode.OP_RNGE, bytecode.FLG__, 0)

	case "for":
		var tstIx int
		var parts []interface{}
		var ok bool
		start := len(fn.Is)
		empty := e.isEmpty(sym.First)
		longForm := false
		if !empty {
			var cond interface{}
			if parts, ok = sym.First.([]interface{}); ok {
				// 3-part form, render the init part
				e.assert(len(parts) == 3, errors.New("expected 3-part `for` loop to have 3 parts, got "+strconv.Itoa(len(parts))))
				longForm = true
				e.emitAny(f, fn, sym, parts[0])
				// The start of the loop, for the jumpback instruction, is now the next instr
				start = len(fn.Is)
				cond = parts[1]
			} else {
				cond = sym.First
			}
			// Emit the condition
			e.emitAny(f, fn, sym, cond)
			// Add a test instruction placeholder
			tstIx = e.addTempInstr(fn)
		}
		// Emit the body
		e.startFor(fn)
		e.emitAny(f, fn, sym, sym.Second)
		// Update the continue statements (must jump to the next statement)
		e.updateForJmp(fn, false)
		if !empty && longForm {
			// Emit the post statement
			e.emitAny(f, fn, sym, parts[2])
		}
		// Add the jump-back to for condition instruction (or for body start if no condition)
		e.addInstr(fn, bytecode.OP_JMP, bytecode.FLG_Jb, uint64(len(fn.Is)-start))
		if !empty {
			// Update the test instruction
			e.updateTestInstr(fn, tstIx)
		}
		// The break statements must jump to the next statement (after the whole for loop)
		e.updateForJmp(fn, true)
		e.endFor(fn)
	case "debug":
		var err error
		var ix int64 = 1 // Default to 1 stack to dump
		if sym.First != nil {
			// If present, it must be a literal number
			ix, err = strconv.ParseInt(sym.First.(*parser.Symbol).Val.(string), 10, 64)
			e.assert(err == nil, errors.New("invalid number literal"))
		}
		e.addInstr(fn, bytecode.OP_DUMP, bytecode.FLG_Sn, uint64(ix))
	case "break":
		e.assert(len(e.forNest[fn]) > 0, errors.New("invalid break statement outside any `for` loop"))
		e.addForData(fn, true, e.addTempInstr(fn))
	case "continue":
		e.assert(len(e.forNest[fn]) > 0, errors.New("invalid continue statement outside any `for` loop"))
		e.addForData(fn, false, e.addTempInstr(fn))
	case "yield":
		e.assert(len(e.fnIx) > 1, errors.New("cannot yield from the top-level module function"))
		// Push the value to yield
		e.emitSymbol(f, fn, sym.First.(*parser.Symbol), atFalse)
		// Yield
		e.addInstr(fn, bytecode.OP_YLD, bytecode.FLG__, 0)
	case "return":
		e.emitSymbol(f, fn, sym.First.(*parser.Symbol), atFalse)
		e.addInstr(fn, bytecode.OP_RET, bytecode.FLG__, 0)
	default:
		e.err = errors.New("unexpected symbol id: " + sym.Id)
	}
	// After treating the symbol, if it had a Key value, push the Key name
	if sym.Key != nil {
		// Can be on name, literal, func call, any operator, hard to assert...
		kix := e.registerK(fn, sym.Key, true, false)
		e.addInstr(fn, bytecode.OP_PUSH, bytecode.FLG_K, kix)
	}
}

func (e *Emitter) startFor(fn *bytecode.Fn) {
	e.forNest[fn] = append(e.forNest[fn], &forData{})
}

func (e *Emitter) endFor(fn *bytecode.Fn) {
	fors := e.forNest[fn]
	e.forNest[fn] = fors[:len(fors)-1]
}

func (e *Emitter) updateForJmp(fn *bytecode.Fn, br bool) {
	fors := e.forNest[fn]
	f := fors[len(fors)-1]
	sl := f.breaks
	if !br {
		sl = f.conts
	}
	for _, ix := range sl {
		e.updateJumpfInstr(fn, ix)
	}
}

func (e *Emitter) addForData(fn *bytecode.Fn, br bool, ix int) {
	fors := e.forNest[fn]
	f := fors[len(fors)-1]
	if br {
		f.breaks = append(f.breaks, ix)
	} else {
		f.conts = append(f.conts, ix)
	}
}

func (e *Emitter) isEmpty(v interface{}) bool {
	if v == nil {
		return true
	}
	switch s := v.(type) {
	case *parser.Symbol:
		return s == nil
	case []*parser.Symbol:
		return len(s) == 0
	case []interface{}:
		return len(s) == 0
	}
	return true
}

func (e *Emitter) addTempInstr(fn *bytecode.Fn) int {
	e.addInstr(fn, bytecode.OP_INVL, bytecode.FLG_INVL, 0)
	return len(fn.Is) - 1
}

func (e *Emitter) updateTestInstr(fn *bytecode.Fn, ix int) {
	fn.Is[ix] = bytecode.NewInstr(bytecode.OP_TEST, bytecode.FLG_Jf, uint64(len(fn.Is)-ix-1))
}

func (e *Emitter) updateJumpfInstr(fn *bytecode.Fn, ix int) {
	fn.Is[ix] = bytecode.NewInstr(bytecode.OP_JMP, bytecode.FLG_Jf, uint64(len(fn.Is)-ix-1))
}

func (e *Emitter) addInstr(fn *bytecode.Fn, op bytecode.Opcode, flg bytecode.Flag, ix uint64) {
	if e.err != nil {
		return
	}
	switch op {
	case bytecode.OP_PUSH:
		e.stackSz[fn] += 1
	case bytecode.OP_NEW:
		e.stackSz[fn] += (1 - (2 * int64(ix)))
	case bytecode.OP_POP, bytecode.OP_RET, bytecode.OP_UNM, bytecode.OP_NOT, bytecode.OP_TEST,
		bytecode.OP_LT, bytecode.OP_LTE, bytecode.OP_GT, bytecode.OP_GTE, bytecode.OP_EQ,
		bytecode.OP_ADD, bytecode.OP_SUB, bytecode.OP_MUL,
		bytecode.OP_DIV, bytecode.OP_MOD, bytecode.OP_GFLD, bytecode.OP_NEQ:
		e.stackSz[fn] -= 1
	case bytecode.OP_SFLD:
		e.stackSz[fn] -= 3
	case bytecode.OP_CALL:
		e.stackSz[fn] -= (int64(ix) + 1)
	case bytecode.OP_CFLD:
		e.stackSz[fn] -= (int64(ix) + 2)
	}
	if e.stackSz[fn] > fn.Header.StackSz {
		fn.Header.StackSz = e.stackSz[fn]
	}
	fn.Is = append(fn.Is, bytecode.NewInstr(op, flg, ix))
}

func (e *Emitter) registerK(fn *bytecode.Fn, val interface{}, isName bool, local bool) uint64 {
	var kt bytecode.KType
	s, ok := val.(string)
	if ok {
		if isName {
			val = s
			kt = bytecode.KtString
		} else if s[0] == '"' || s[0] == '`' {
			// Unquote the string, keeping escaped characters
			var err error
			s, err = strconv.Unquote(s)
			e.assert(err == nil, err)
			val = s
			kt = bytecode.KtString
		} else if strings.Index(s, ".") >= 0 {
			val, e.err = strconv.ParseFloat(s, 64)
			kt = bytecode.KtFloat
		} else {
			val, e.err = strconv.ParseInt(s, 10, 64)
			kt = bytecode.KtInteger
		}
	} else {
		kt = bytecode.KtBoolean
		if v := val.(bool); v {
			s = "true"
			val = int64(1)
		} else {
			s = "false"
			val = int64(0)
		}
	}
	// Create the map slot for this function if this is its first symbol
	m, ok := e.kMap[fn]
	if !ok {
		m = make(map[kId]int)
		e.kMap[fn] = m
	}
	// Check if this symbol already exists for this fn with this same type, otherwise add it to its Ks
	i, ok := m[kId{s, kt}]
	if !ok {
		i = len(m)
		m[kId{s, kt}] = i
		fn.Ks = append(fn.Ks, &bytecode.K{Type: kt, Val: val})
	}
	// If this is a local definition, add the K index to this function's L table.
	// It can't have duplicates by definition, because it would have been caught as an
	// error in the parser stage.
	if local {
		fn.Ls = append(fn.Ls, int64(i))
	}
	return uint64(i)
}

func (e *Emitter) assert(cond bool, err error) {
	if !cond {
		e.err = err
	}
}
