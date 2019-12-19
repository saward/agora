package stdlib

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/saward/agora/runtime"
)

func TestFmtPrint(t *testing.T) {
	ctx := context.Background()
	ktx := runtime.NewKtx(nil, nil)

	cases := []struct {
		src   []runtime.Val
		exp   string
		expln string
		start bool
	}{
		0: {
			src: []runtime.Val{runtime.Nil},
			exp: "nil",
		},
		1: {
			src:   []runtime.Val{runtime.Bool(true), runtime.Bool(false)},
			exp:   "truefalse",
			expln: "true false",
		},
		2: {
			// Ok, so print does *NOT* add spaces when the value is a native string
			src:   []runtime.Val{runtime.String("string"), runtime.Number(0), runtime.Number(-1), runtime.Number(17), runtime.String("pi"), runtime.Number(3.1415)},
			exp:   "string0-117pi3.1415",
			expln: "string 0 -1 17 pi 3.1415",
		},
		3: {
			src: []runtime.Val{runtime.String("func:"),
				runtime.NewNativeFunc(ktx, "", func(_ context.Context, args ...runtime.Val) runtime.Val { return runtime.Nil })},
			exp:   "func:<func  (",
			expln: "func: <func  (",
			start: true,
		},
		4: {
			src: []runtime.Val{runtime.NewObject()},
			exp: "{}",
		},
	}

	fm := new(FmtMod)
	fm.SetKtx(ktx)
	buf := bytes.NewBuffer(nil)
	ktx.Stdout = buf
	for i, c := range cases {
		for j := 0; j < 2; j++ {
			var res runtime.Val
			buf.Reset()
			if j == 1 {
				if c.expln != "" {
					c.exp = c.expln
				}
				if !c.start {
					c.exp += "\n"
				}
				res = fm.fmt_Println(ctx, c.src...)
			} else {
				res = fm.fmt_Print(ctx, c.src...)
			}
			if (c.start && !strings.HasPrefix(buf.String(), c.exp)) || (!c.start && c.exp != buf.String()) {
				t.Errorf("[%d] - expected %s, got %s", i, c.exp, buf.String())
			}
			if !c.start && res.Int(ctx) != int64(len(c.exp)) {
				t.Errorf("[%d] - expected return value of %d, got %d", i, len(c.exp), res.Int(ctx))
			}
		}
	}
}

func TestFmtScanln(t *testing.T) {
	ctx := context.Background()
	ktx := runtime.NewKtx(nil, nil)
	buf := bytes.NewBuffer([]byte(`This is
two lines
`))
	ktx.Stdin = buf
	fm := new(FmtMod)
	fm.SetKtx(ktx)
	ret := fm.fmt_Scanln(ctx)
	if ret.String(ctx) != "This is" {
		t.Errorf("expected line 1 to be 'This is', got '%s'", ret)
	}
}

func TestFmtScanint(t *testing.T) {
	ctx := context.Background()
	ktx := runtime.NewKtx(nil, nil)
	buf := bytes.NewBuffer([]byte("12\n"))
	ktx.Stdin = buf
	fm := new(FmtMod)
	fm.SetKtx(ktx)
	ret := fm.fmt_Scanint(ctx)
	if ret.Int(ctx) != 12 {
		t.Errorf("expected 12, got %d", ret.Int(ctx))
	}
}
