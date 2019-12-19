package compiler

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/saward/agora/bytecode"
	. "github.com/saward/agora/bytecode/testing"
)

var (
	disasmcases = []struct {
		src []byte
		exp string
		err error
	}{
		0: {
			// No input
			err: bytecode.ErrInvalidData,
		},
		1: {
			// Empty func
			src: AppendAny(SigVer(bytecode.Version()), Int64ToByteSlice(4), 't', 'e', 's', 't',
				// StackSz - ExpArgs - ParentFnIx - LineStart - LineEnd
				ExpZeroInt64, ExpZeroInt64, ExpZeroInt64, ExpZeroInt64, ExpZeroInt64,
				// Ks - Ls - Is
				ExpZeroInt64, ExpZeroInt64, ExpZeroInt64),
			exp: disasmComment + `
[f]
test
0
0
0
0
0
[k]
[l]
[i]
`,
		},
		2: {
			// Full valid func
			src: AppendAny(SigVer(bytecode.Version()), Int64ToByteSlice(4), 't', 'e', 's', 't',
				// StackSz - ExpArgs - ParentFnIx - LineStart - LineEnd
				Int64ToByteSlice(1), ExpZeroInt64, ExpZeroInt64, ExpZeroInt64, Int64ToByteSlice(2),
				// Ks - Ls - Is
				Int64ToByteSlice(2), 's', Int64ToByteSlice(1), 'a', 'i', Int64ToByteSlice(5), ExpZeroInt64, Int64ToByteSlice(5),
				// 5 ops
				UInt64ToByteSlice(uint64(bytecode.NewInstr(bytecode.NewOpcode("PUSH"), bytecode.NewFlag("K"), 1))),
				UInt64ToByteSlice(uint64(bytecode.NewInstr(bytecode.NewOpcode("POP"), bytecode.NewFlag("V"), 0))),
				UInt64ToByteSlice(uint64(bytecode.NewInstr(bytecode.NewOpcode("PUSH"), bytecode.NewFlag("V"), 0))),
				UInt64ToByteSlice(uint64(bytecode.NewInstr(bytecode.NewOpcode("DUMP"), bytecode.NewFlag("Sn"), 1))),
				UInt64ToByteSlice(uint64(bytecode.NewInstr(bytecode.NewOpcode("RET"), bytecode.NewFlag("_"), 0))),
			),
			exp: disasmComment + `
[f]
test
1
0
0
0
2
[k]
sa
i5
[l]
[i]
PUSH K 1
POP V 0
PUSH V 0
DUMP Sn 1
RET _ 0
`,
		},
		3: {
			// Many functions, valid
			src: AppendAny(SigVer(bytecode.Version()), Int64ToByteSlice(4), 't', 'e', 's', 't',
				// StackSz - ExpArgs - ParentFnIx - LineStart - LineEnd
				Int64ToByteSlice(3), ExpZeroInt64, ExpZeroInt64, ExpZeroInt64, Int64ToByteSlice(3),
				// Ks - Ls - Is
				Int64ToByteSlice(3), 's', Int64ToByteSlice(3), 'A', 'd', 'd', 'i', Int64ToByteSlice(4), 's', Int64ToByteSlice(3), '1', '9', '8', ExpZeroInt64, Int64ToByteSlice(8),
				// 8 ops
				UInt64ToByteSlice(uint64(bytecode.NewInstr(bytecode.NewOpcode("PUSH"), bytecode.NewFlag("F"), 1))),
				UInt64ToByteSlice(uint64(bytecode.NewInstr(bytecode.NewOpcode("POP"), bytecode.NewFlag("V"), 0))),
				UInt64ToByteSlice(uint64(bytecode.NewInstr(bytecode.NewOpcode("PUSH"), bytecode.NewFlag("K"), 1))),
				UInt64ToByteSlice(uint64(bytecode.NewInstr(bytecode.NewOpcode("PUSH"), bytecode.NewFlag("K"), 2))),
				UInt64ToByteSlice(uint64(bytecode.NewInstr(bytecode.NewOpcode("PUSH"), bytecode.NewFlag("V"), 0))),
				UInt64ToByteSlice(uint64(bytecode.NewInstr(bytecode.NewOpcode("CALL"), bytecode.NewFlag("An"), 2))),
				UInt64ToByteSlice(uint64(bytecode.NewInstr(bytecode.NewOpcode("DUMP"), bytecode.NewFlag("Sn"), 1))),
				UInt64ToByteSlice(uint64(bytecode.NewInstr(bytecode.NewOpcode("RET"), bytecode.NewFlag("_"), 0))),
				// 2nd fn
				Int64ToByteSlice(3), 'A', 'd', 'd',
				// StackSz - ExpArgs - ParentFnIx - LineStart - LineEnd
				Int64ToByteSlice(2), Int64ToByteSlice(2), ExpZeroInt64, ExpZeroInt64, Int64ToByteSlice(2),
				// Ks - Ls - Is
				Int64ToByteSlice(2), 's', Int64ToByteSlice(1), 'x', 's', Int64ToByteSlice(1), 'y', ExpZeroInt64, Int64ToByteSlice(4),
				// 4 ops
				UInt64ToByteSlice(uint64(bytecode.NewInstr(bytecode.NewOpcode("PUSH"), bytecode.NewFlag("V"), 0))),
				UInt64ToByteSlice(uint64(bytecode.NewInstr(bytecode.NewOpcode("PUSH"), bytecode.NewFlag("V"), 1))),
				UInt64ToByteSlice(uint64(bytecode.NewInstr(bytecode.NewOpcode("ADD"), bytecode.NewFlag("_"), 0))),
				UInt64ToByteSlice(uint64(bytecode.NewInstr(bytecode.NewOpcode("RET"), bytecode.NewFlag("_"), 0))),
			),
			exp: disasmComment + `
[f]
test
3
0
0
0
3
[k]
sAdd
i4
s198
[l]
[i]
PUSH F 1
POP V 0
PUSH K 1
PUSH K 2
PUSH V 0
CALL An 2
DUMP Sn 1
RET _ 0
[f]
Add
2
2
0
0
2
[k]
sx
sy
[l]
[i]
PUSH V 0
PUSH V 1
ADD _ 0
RET _ 0
`,
		},
	}

	isolateDisasmCase = -1
)

func TestDisasm(t *testing.T) {
	d := new(Disasm)
	buf := bytes.NewBuffer(nil)
	for i, c := range disasmcases {
		if isolateDisasmCase >= 0 && isolateDisasmCase != i {
			continue
		}
		if testing.Verbose() {
			fmt.Printf("testing disassembler case %d...\n", i)
		}

		// Act
		buf.Reset()
		err := d.Uncompile(bytes.NewReader(c.src), buf)

		// Assert
		if err != c.err {
			if c.err == nil {
				t.Errorf("[%d] - expected no error, got `%s`", i, err)
			} else {
				t.Errorf("[%d] - expected error `%s`, got `%s`", i, c.err, err)
			}
		}
		if c.exp != "" {
			got := buf.String()
			if got != c.exp {
				t.Errorf("[%d] - expected \n%s\n, got \n%s\n", i, c.exp, got)
			}
		}
		if c.err == nil && c.exp == "" {
			t.Errorf("[%d] - no assertion", i)
		}
	}
}
