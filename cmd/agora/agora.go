// Package main provides the command-line tool `agora`.
//
// This tool offers the following commands:
// - agora run : run an agora source code file.
// - agora build : compile an agora source code file.
// - agora asm : compile an agora assembly code file.
// - agora dasm : disassemble an agora bytecode into assembly source.
// - agora ast : generate the abstract syntax tree for an agora source code file.
//
// See `agora -h` and `agora <cmd> -h` for available options.
package main

import (
	"bytes"
	"context"
	"fmt"
	"go/scanner"
	"io"
	"io/ioutil"
	"os"

	"github.com/saward/agora/bytecode"
	"github.com/saward/agora/compiler"
	"github.com/saward/agora/compiler/parser"
	"github.com/saward/agora/runtime"
	"github.com/saward/agora/runtime/stdlib"
	"github.com/jessevdk/go-flags"
)

var (
	// For test purpose
	stdout io.Writer = os.Stdout
)

// The assembler command struct.
type asm struct {
	Output string `short:"o" long:"output" description:"output file"`
	Hexa   bool   `short:"x" long:"hexadecimal" description:"print output as hexadecimal"`
}

// Execute the assembler command
func (a *asm) Execute(args []string) error {
	// Validate input
	if len(args) != 1 {
		return fmt.Errorf("expected an input file name")
	}
	// Open input file
	inf, err := os.Open(args[0])
	if err != nil {
		return err
	}
	defer inf.Close()
	// Compile to bytecode File
	f, err := new(compiler.Asm).Compile(args[0], inf)
	if err != nil {
		return err
	}
	// Write output
	var out io.Writer
	out = stdout
	if a.Output != "" {
		outF, err := os.Create(a.Output)
		if err != nil {
			return err
		}
		defer outF.Close()
		out = outF
	}
	// Encode to bytecode
	buf := bytes.NewBuffer(nil)
	err = bytecode.NewEncoder(buf).Encode(f)
	if err != nil {
		return err
	}
	if a.Hexa {
		_, err = io.WriteString(out, fmt.Sprintf("%x", buf.Bytes()))
	} else {
		_, err = out.Write(buf.Bytes())
	}
	if err != nil {
		return err
	}
	return nil
}

// The disassembler command struct
type dasm struct {
	Output string `short:"o" long:"output" description:"output file"`
}

// Execute the disassembler command
func (d *dasm) Execute(args []string) error {
	// Validate input
	if len(args) != 1 {
		return fmt.Errorf("expected an input file name")
	}
	// Open input file
	inf, err := os.Open(args[0])
	if err != nil {
		return err
	}
	defer inf.Close()
	// Open output file
	var out io.Writer
	out = stdout
	if d.Output != "" {
		outF, err := os.Create(d.Output)
		if err != nil {
			return err
		}
		defer outF.Close()
		out = outF
	}
	// Compile to assembly
	return new(compiler.Disasm).Uncompile(inf, out)
}

// The run command struct
type run struct {
	FromAsm  bool   `short:"a" long:"from-asm" description:"run an assembly input"`
	NoStdlib bool   `short:"S" long:"no-stdlib" description:"do not import the stdlib"`
	Debug    bool   `short:"d" long:"debug" description:"output debug information"`
	NoResult bool   `short:"R" long:"no-result" description:"do not print the result"`
	Output   string `short:"o" long:"output" description:"output file"`
}

func (r *run) Execute(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("expected an input file")
	}
	var c runtime.Compiler
	if r.FromAsm {
		c = new(compiler.Asm)
	} else {
		c = new(compiler.Compiler)
	}
	ktx := runtime.NewKtx(new(runtime.FileResolver), c)
	if !r.NoStdlib {
		// Register the standard lib's Fmt package
		ktx.RegisterNativeModule(new(stdlib.FmtMod))
		ktx.RegisterNativeModule(new(stdlib.FilepathMod))
		ktx.RegisterNativeModule(new(stdlib.StringsMod))
		ktx.RegisterNativeModule(new(stdlib.MathMod))
		ktx.RegisterNativeModule(new(stdlib.OsMod))
		ktx.RegisterNativeModule(new(stdlib.TimeMod))
	}
	ktx.Debug = r.Debug
	m, err := ktx.Load(args[0])
	if err != nil {
		return err
	}
	// Prepare extra parameters to send to module
	vals := make([]runtime.Val, len(args)-1)
	for i, s := range args[1:] {
		vals[i] = runtime.String(s)
	}
	// Open output stream
	outf := os.Stdout
	if r.Output != "" {
		outf, err = os.Open(r.Output)
		if err != nil {
			return err
		}
		defer outf.Close()
		ktx.Stdout = outf
	}
	res, err := m.Run(ctx, vals...)
	if err == nil && !r.NoResult {
		fmt.Fprintf(outf, "\n= %s (%T)\n", res, res)
	}
	return err
}

// The ast command struct
type ast struct {
	Output    string `short:"o" long:"output" description:"output file"`
	AllErrors bool   `short:"e" long:"all-errors" description:"print all errors"`
}

func (a *ast) Execute(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("expected an input file")
	}
	f, err := os.Open(args[0])
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	p := parser.New()
	syms, _, err := p.Parse(args[0], b)
	if err != nil {
		if a.AllErrors {
			scanner.PrintError(stdout, err)
		}
		return err
	}
	out := stdout
	if a.Output != "" {
		outf, err := os.Open(a.Output)
		if err != nil {
			return err
		}
		defer outf.Close()
		out = outf
	}
	for _, sym := range syms {
		fmt.Fprintln(out, sym)
	}
	return nil
}

// The build command struct
type build struct {
	Output string `short:"o" long:"output" description:"output file"`
	Asm    bool   `short:"a" long:"assembly" description:"build to assembly instead of bytecode"`
}

func (b *build) Execute(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("expected an input file")
	}
	inf, err := os.Open(args[0])
	if err != nil {
		return err
	}
	defer inf.Close()
	c := new(compiler.Compiler)
	f, err := c.Compile(args[0], inf)
	if err != nil {
		return err
	}
	out := stdout
	if b.Output != "" {
		outf, err := os.Create(b.Output)
		if err != nil {
			return err
		}
		defer outf.Close()
		out = outf
	}
	if b.Asm {
		dasm := new(compiler.Disasm)
		err = dasm.ToAsm(f, out)
	} else {
		enc := bytecode.NewEncoder(out)
		err = enc.Encode(f)
	}
	if err != nil {
		return err
	}
	return nil
}

type version struct{}

func (v *version) Execute(args []string) error {
	maj, min := bytecode.Version()
	fmt.Printf("agora version %d.%d (git revision %s)\n", maj, min, bytecode.AGORA_BUILD)
	return nil
}

func main() {
	a, d, r, s, b, v := new(asm), new(dasm), new(run), new(ast), new(build), new(version)
	p := flags.NewParser(nil, flags.Default)
	p.AddCommand("asm", "assembler", "compile assembly to bytecode", a)
	p.AddCommand("dasm", "disassembler", "disassemble bytecode to assembly", d)
	p.AddCommand("run", "run", "execute a source program", r)
	p.AddCommand("ast", "abstract syntax tree", "print the AST of a source program", s)
	p.AddCommand("build", "compiler", "compile a source program", b)
	p.AddCommand("version", "print the current version", "print the current version", v)
	// In case of errors, usage text is automatically displayed. In case of
	// success, the Execute() method of the matching command is called.
	p.Parse()
}
