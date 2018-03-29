package stdlib

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/bobg/agora/runtime"
)

func TestFilepathAbs(t *testing.T) {
	ctx := context.Background()
	ktx := runtime.NewKtx(nil, nil)
	fm := new(FilepathMod)
	fm.SetKtx(ktx)
	p := "./testdata"
	// Abs
	exp, e := filepath.Abs(p)
	if e != nil {
		panic(e)
	}
	ret := fm.filepath_Abs(ctx, runtime.String(p))
	if ret.String(ctx) != exp {
		t.Errorf("expected '%s', got '%s'", exp, ret.String(ctx))
	}
	// IsAbs
	{
		exp := filepath.IsAbs(p)
		ret := fm.filepath_IsAbs(ctx, runtime.String(p))
		if ret.Bool(ctx) != exp {
			t.Errorf("expected '%v', got '%v'", exp, ret.Bool(ctx))
		}
	}
}

func TestFilepathBaseDirExt(t *testing.T) {
	ctx := context.Background()
	ktx := runtime.NewKtx(nil, nil)
	fm := new(FilepathMod)
	fm.SetKtx(ktx)
	p, e := filepath.Abs("./testdata/readfile.txt")
	if e != nil {
		panic(e)
	}
	// Base
	exp := filepath.Base(p)
	ret := fm.filepath_Base(ctx, runtime.String(p))
	if ret.String(ctx) != exp {
		t.Errorf("expected base '%s', got '%s'", exp, ret.String(ctx))
	}
	// Dir
	exp = filepath.Dir(p)
	ret = fm.filepath_Dir(ctx, runtime.String(p))
	if ret.String(ctx) != exp {
		t.Errorf("expected dir '%s', got '%s'", exp, ret.String(ctx))
	}
	// Ext
	exp = filepath.Ext(p)
	ret = fm.filepath_Ext(ctx, runtime.String(p))
	if ret.String(ctx) != exp {
		t.Errorf("expected extension '%s', got '%s'", exp, ret.String(ctx))
	}
}

func TestFilepathJoin(t *testing.T) {
	ctx := context.Background()
	ktx := runtime.NewKtx(nil, nil)
	fm := new(FilepathMod)
	fm.SetKtx(ktx)
	parts := []string{"./testdata", "..", "../../compiler", "test"}
	exp := filepath.Join(parts...)
	vals := make([]runtime.Val, len(parts))
	for i, s := range parts {
		vals[i] = runtime.String(s)
	}
	ret := fm.filepath_Join(ctx, vals...)
	if ret.String(ctx) != exp {
		t.Errorf("expected '%s', got '%s'", exp, ret.String(ctx))
	}
}
