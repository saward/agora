package bytecode

import (
	"bytes"
	"fmt"
	"testing"

	. "github.com/saward/agora/bytecode/testing"
	"github.com/davecgh/go-spew/spew"
)

var (
	defMaj = _MAJOR_VERSION
	defMin = _MINOR_VERSION

	deccases = []struct {
		maj int
		min int
		src []byte
		exp *File
		err error
	}{
		0: {
			// Simplest case, decodes the file header only
			maj: defMaj,
			min: defMin,
			src: SigVer(defMaj, defMin),
			exp: &File{
				MajorVersion: defMaj,
				MinorVersion: defMin,
			},
		},
		1: {
			// Decodes the file header and function header
			maj: defMaj,
			min: defMin,
			src: AppendAny(SigVer(defMaj, defMin), Int64ToByteSlice(4), 't', 'e', 's', 't',
				// StackSz - ExpArgs - ParentFnIx - LineStart - LineEnd
				Int64ToByteSlice(2), Int64ToByteSlice(3), ExpZeroInt64, Int64ToByteSlice(5), Int64ToByteSlice(6),
				// Ks - Ls - Is
				ExpZeroInt64, ExpZeroInt64, ExpZeroInt64),
			exp: &File{
				MajorVersion: defMaj,
				MinorVersion: defMin,
				Name:         "test", Fns: []*Fn{
					&Fn{
						Header: H{
							Name:       "test",
							StackSz:    2,
							ExpArgs:    3,
							ParentFnIx: 0,
							LineStart:  5,
							LineEnd:    6,
						},
					},
				}},
		},
		2: {
			// Invalid version
			maj: 1,
			min: 2,
			src: AppendAny(ExpSig, encodeVersionByte(2, 3), Int64ToByteSlice(4), 't', 'e', 's', 't',
				// StackSz - ExpArgs - ParentFnIx - LineStart - LineEnd
				Int64ToByteSlice(2), Int64ToByteSlice(3), ExpZeroInt64, Int64ToByteSlice(5), Int64ToByteSlice(6),
				// Ks - Ls - Is
				ExpZeroInt64, ExpZeroInt64, ExpZeroInt64),
			err: ErrVersionMismatch,
		},
		3: {
			// Top-level function gets the file name
			maj: defMaj,
			min: defMin,
			src: AppendAny(SigVer(defMaj, defMin), Int64ToByteSlice(4), 't', 'e', 's', 't',
				// StackSz - ExpArgs - ParentFnIx - LineStart - LineEnd
				ExpZeroInt64, ExpZeroInt64, ExpZeroInt64, ExpZeroInt64, ExpZeroInt64,
				// Ks - Ls - Is
				ExpZeroInt64, ExpZeroInt64, ExpZeroInt64),
			exp: &File{
				MajorVersion: defMaj,
				MinorVersion: defMin,
				Name:         "test", Fns: []*Fn{&Fn{Header: H{Name: "test"}}}},
		},
		4: {
			maj: defMaj,
			min: defMin,
			src: AppendAny(SigVer(defMaj, defMin), Int64ToByteSlice(4), 't', 'e', 's', 't',
				// StackSz - ExpArgs - ParentFnIx - LineStart - LineEnd
				Int64ToByteSlice(2), Int64ToByteSlice(3), ExpZeroInt64, Int64ToByteSlice(5), Int64ToByteSlice(6),
				// Ks - Ls - Is
				Int64ToByteSlice(1), byte(KtInteger), Int64ToByteSlice(7), ExpZeroInt64, ExpZeroInt64),
			exp: &File{
				MajorVersion: defMaj,
				MinorVersion: defMin,
				Name:         "test", Fns: []*Fn{
					&Fn{
						Header: H{
							Name:       "test",
							StackSz:    2,
							ExpArgs:    3,
							ParentFnIx: 0,
							LineStart:  5,
							LineEnd:    6,
						},
						Ks: []*K{
							&K{
								Type: KtInteger,
								Val:  int64(7),
							},
						},
					},
				}},
		},
		5: {
			// Invalid K Type
			maj: defMaj,
			min: defMin,
			src: AppendAny(SigVer(defMaj, defMin), Int64ToByteSlice(4), 't', 'e', 's', 't',
				// StackSz - ExpArgs - ParentFnIx - LineStart - LineEnd
				Int64ToByteSlice(2), Int64ToByteSlice(3), ExpZeroInt64, Int64ToByteSlice(5), Int64ToByteSlice(6),
				// Ks - Ls - Is
				Int64ToByteSlice(1), 'z', Int64ToByteSlice(7), ExpZeroInt64, ExpZeroInt64),
			err: ErrInvalidKType,
		},
		6: {
			// Impossible to reproduce same 6 as encode - cannot get invalid K value, it is
			// necessarily read as a type corresponding to its K type.
			maj: defMaj,
			min: defMin,
			err: ErrInvalidData,
		},
		7: {
			// Function with K and Is
			maj: defMaj,
			min: defMin,
			src: AppendAny(SigVer(defMaj, defMin), Int64ToByteSlice(4), 't', 'e', 's', 't',
				// StackSz - ExpArgs - ParentFnIx - LineStart - LineEnd
				Int64ToByteSlice(2), Int64ToByteSlice(3), ExpZeroInt64, Int64ToByteSlice(5), Int64ToByteSlice(6),
				// Ks - Ls - Is
				Int64ToByteSlice(1), byte(KtInteger), Int64ToByteSlice(7), ExpZeroInt64, Int64ToByteSlice(2),
				// 2 Ops
				0x0C, 0x00, 0x00, 0x00, 0x00, 0x00, byte(FLG_K), byte(OP_ADD), 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, byte(FLG_Sn), byte(OP_DUMP)),
			exp: &File{
				MajorVersion: defMaj,
				MinorVersion: defMin,
				Name:         "test", Fns: []*Fn{
					&Fn{
						Header: H{
							Name:       "test",
							StackSz:    2,
							ExpArgs:    3,
							ParentFnIx: 0,
							LineStart:  5,
							LineEnd:    6,
						},
						Ks: []*K{
							&K{
								Type: KtInteger,
								Val:  int64(7),
							},
						},
						Is: []Instr{
							NewInstr(OP_ADD, FLG_K, 12),
							NewInstr(OP_DUMP, FLG_Sn, 0),
						},
					},
				}},
		},
		8: {
			// Invalid opcode
			maj: defMaj,
			min: defMin,
			src: AppendAny(SigVer(defMaj, defMin), Int64ToByteSlice(4), 't', 'e', 's', 't',
				// StackSz - ExpArgs - ParentFnIx - LineStart - LineEnd
				Int64ToByteSlice(2), Int64ToByteSlice(3), ExpZeroInt64, Int64ToByteSlice(5), Int64ToByteSlice(6),
				// Ks - Ls - Is
				Int64ToByteSlice(1), byte(KtInteger), Int64ToByteSlice(7), ExpZeroInt64, Int64ToByteSlice(2),
				// 2 ops
				0x0C, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x09, byte(op_max)),
			err: ErrUnknownOpcode,
		},
		9: {
			// Multiple functions
			maj: defMaj,
			min: defMin,
			src: AppendAny(SigVer(defMaj, defMin), Int64ToByteSlice(4), 't', 'e', 's', 't',
				// StackSz - ExpArgs - ParentFnIx - LineStart - LineEnd
				Int64ToByteSlice(2), Int64ToByteSlice(3), ExpZeroInt64, Int64ToByteSlice(5), Int64ToByteSlice(6),
				// Ks - Ls - Is
				Int64ToByteSlice(1), byte(KtInteger), Int64ToByteSlice(7), ExpZeroInt64, Int64ToByteSlice(2),
				// 2 ops
				0x0C, 0x00, 0x00, 0x00, 0x00, 0x00, byte(FLG_K), byte(OP_ADD), 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, byte(FLG_Sn), byte(OP_DUMP),
				// 2nd Fn
				Int64ToByteSlice(2), 'f', '2',
				// StackSz - ExpArgs - ParentFnIx - LineStart - LineEnd
				Int64ToByteSlice(2), Int64ToByteSlice(3), Int64ToByteSlice(0), Int64ToByteSlice(5), Int64ToByteSlice(6),
				// Ks - Ls - Is
				Int64ToByteSlice(1), byte(KtString), Int64ToByteSlice(5), 'c', 'o', 'n', 's', 't', ExpZeroInt64, Int64ToByteSlice(1),
				// 1 op
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00),
			exp: &File{
				MajorVersion: defMaj,
				MinorVersion: defMin,
				Name:         "test", Fns: []*Fn{
					&Fn{
						Header: H{
							Name:       "test",
							StackSz:    2,
							ExpArgs:    3,
							ParentFnIx: 0,
							LineStart:  5,
							LineEnd:    6,
						},
						Ks: []*K{
							&K{
								Type: KtInteger,
								Val:  int64(7),
							},
						},
						Is: []Instr{
							NewInstr(OP_ADD, FLG_K, 12),
							NewInstr(OP_DUMP, FLG_Sn, 0),
						},
					},
					&Fn{
						Header: H{
							Name:       "f2",
							StackSz:    2,
							ExpArgs:    3,
							ParentFnIx: 0,
							LineStart:  5,
							LineEnd:    6,
						},
						Ks: []*K{
							&K{
								Type: KtString,
								Val:  "const",
							},
						},
						Is: []Instr{
							NewInstr(OP_RET, FLG__, 0),
						},
					},
				}},
		},
	}

	isolateDecCase = -1
)

func TestDecode(t *testing.T) {
	for i, c := range deccases {
		if isolateDecCase >= 0 && isolateDecCase != i {
			continue
		}
		if testing.Verbose() {
			fmt.Printf("testing decode case %d...\n", i)
		}

		// Arrange
		_MAJOR_VERSION = c.maj
		_MINOR_VERSION = c.min

		// Act
		f, err := NewDecoder(bytes.NewBuffer(c.src)).Decode()

		// Assert
		if err != c.err {
			if c.err == nil {
				t.Errorf("[%d] - expected no error, got `%s`", i, err)
			} else {
				t.Errorf("[%d] - expected error `%s`, got `%s`", i, c.err, err)
			}
		}
		if c.exp != nil {
			if !equal(f, c.exp) {
				t.Errorf("[%d] - expected\n", i)
				t.Error(spew.Sdump(c.exp))
				t.Error("got\n")
				t.Error(spew.Sdump(f))
			}
		}
		if c.err == nil && c.exp == nil {
			t.Errorf("[%d] - no assertion", i)
		}
	}
}

func equal(f1, f2 *File) bool {
	if f1 == nil && f2 == nil {
		return true
	}
	if f1 == nil || f2 == nil {
		return false
	}
	if f1.Name != f2.Name {
		return false
	}
	if f1.MajorVersion != f2.MajorVersion {
		return false
	}
	if f1.MinorVersion != f2.MinorVersion {
		return false
	}
	if len(f1.Fns) != len(f2.Fns) {
		return false
	}
	for i := 0; i < len(f1.Fns); i++ {
		fn1, fn2 := f1.Fns[i], f2.Fns[i]
		if fn1.Header.Name != fn2.Header.Name {
			return false
		}
		if fn1.Header.StackSz != fn2.Header.StackSz {
			return false
		}
		if fn1.Header.ExpArgs != fn2.Header.ExpArgs {
			return false
		}
		if fn1.Header.LineStart != fn2.Header.LineStart {
			return false
		}
		if fn1.Header.LineEnd != fn2.Header.LineEnd {
			return false
		}
		if len(fn1.Ks) != len(fn2.Ks) {
			return false
		}
		for j := 0; j < len(fn1.Ks); j++ {
			k1, k2 := fn1.Ks[j], fn2.Ks[j]
			if k1.Type != k2.Type {
				return false
			}
			if k1.Val != k2.Val {
				return false
			}
		}
		if len(fn1.Ls) != len(fn2.Ls) {
			return false
		}
		for j := 0; j < len(fn1.Ls); j++ {
			if fn1.Ls[j] != fn2.Ls[j] {
				return false
			}
		}
		if len(fn1.Is) != len(fn2.Is) {
			return false
		}
		for j := 0; j < len(fn1.Is); j++ {
			if fn1.Is[j] != fn2.Is[j] {
				return false
			}
		}
	}
	return true
}
