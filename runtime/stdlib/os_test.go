package stdlib

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/saward/agora/runtime"
)

func TestOsTryOpen(t *testing.T) {
	ctx := context.Background()
	ktx := runtime.NewKtx(nil, nil)
	om := new(OsMod)
	om.SetKtx(ktx)
	// With an unknown file
	fn := "./testdata/unknown.txt"
	ret := om.os_TryOpen(ctx, runtime.String(fn))
	if ret != runtime.Nil {
		t.Errorf("expected unknown file to return nil, got '%v'", ret)
	}
	// With an existing file
	fn = "./testdata/readfile.txt"
	ret = om.os_TryOpen(ctx, runtime.String(fn))
	if fl, ok := ret.(*file); !ok {
		t.Errorf("expected existing file to return *file, got '%T'", ret)
	} else {
		fl.closeFile(ctx)
	}
}

func TestOsOpen(t *testing.T) {
	ctx := context.Background()
	ktx := runtime.NewKtx(nil, nil)
	om := new(OsMod)
	om.SetKtx(ktx)
	fn := "./testdata/readfile.txt"
	f := om.os_Open(ctx, runtime.String(fn))
	fl := f.(*file)
	ret := fl.Get(runtime.String("Name"))
	if ret.String(ctx) != fn {
		t.Errorf("expected Name to be '%s', got '%s'", fn, ret)
	}
	exp := "ok"
	ret = fl.readLine(ctx)
	if ret.String(ctx) != exp {
		t.Errorf("expected read line 1 to be '%s', got '%s'", exp, ret)
	}
	exp = ""
	ret = fl.readLine(ctx)
	if ret.String(ctx) != exp {
		t.Errorf("expected read line 2 to be '%s', got '%s'", exp, ret)
	}
	ret = fl.readLine(ctx)
	if ret != runtime.Nil {
		t.Errorf("expected read line 3 to be nil, got '%v'", ret)
	}
	ret = fl.closeFile(ctx)
	if ret != runtime.Nil {
		t.Errorf("expected close file to be nil, got '%v'", ret)
	}
}

func TestOsWrite(t *testing.T) {
	ctx := context.Background()
	ktx := runtime.NewKtx(nil, nil)
	om := new(OsMod)
	om.SetKtx(ktx)
	fn := "./testdata/write.txt"
	f := om.os_Open(ctx, runtime.String(fn), runtime.String("w+"))
	fl := f.(*file)
	defer fl.closeFile(ctx)
	// Write the first value
	ret := fl.writeLine(ctx, runtime.Number(1))
	if ret.Int(ctx) != 2 {
		t.Errorf("expected 1st written length to be 2, got %d", ret.Int(ctx))
	}
	// Move back to start
	ret = fl.seek(ctx)
	if ret.Int(ctx) != 0 {
		t.Errorf("expected seek to return to start, got offset %d", ret.Int(ctx))
	}
	// Write the second value
	ret = fl.writeLine(ctx, runtime.Number(2))
	if ret.Int(ctx) != 2 {
		t.Errorf("expected 2nd written length to be 2, got %d", ret.Int(ctx))
	}
}

func TestOsFields(t *testing.T) {
	ctx := context.Background()
	ktx := runtime.NewKtx(nil, nil)
	om := new(OsMod)
	om.SetKtx(ktx)
	ob, err := om.Run(ctx)
	if err != nil {
		panic(err)
	}
	{
		ob := ob.(runtime.Object)
		ret := ob.Get(runtime.String("PathSeparator"))
		exp := string(os.PathSeparator)
		if ret.String(ctx) != exp {
			t.Errorf("expected path separator %s, got %s", exp, ret.String(ctx))
		}
		ret = ob.Get(runtime.String("PathListSeparator"))
		exp = string(os.PathListSeparator)
		if ret.String(ctx) != exp {
			t.Errorf("expected path list separator %s, got %s", exp, ret.String(ctx))
		}
		ret = ob.Get(runtime.String("DevNull"))
		exp = os.DevNull
		if ret.String(ctx) != exp {
			t.Errorf("expected dev/null %s, got %s", exp, ret)
		}
		ret = ob.Get(runtime.String("TempDir"))
		exp = os.TempDir()
		if ret.String(ctx) != exp {
			t.Errorf("expected temp dir %s, got %s", exp, ret)
		}
	}
}

func TestOsExec(t *testing.T) {
	ctx := context.Background()
	ktx := runtime.NewKtx(nil, nil)
	om := new(OsMod)
	om.SetKtx(ktx)
	exp := "hello"
	ret := om.os_Exec(ctx, runtime.String("echo"), runtime.String(exp))
	// Shell adds a \n after output
	if ret.String(ctx) != exp+"\n" {
		t.Errorf("expected '%s', got '%s'", exp, ret)
	}
}

func TestOsGetenv(t *testing.T) {
	ctx := context.Background()
	ktx := runtime.NewKtx(nil, nil)
	om := new(OsMod)
	om.SetKtx(ktx)
	exp := "ok"
	e := os.Setenv("TEST", exp)
	if e != nil {
		panic(e)
	}
	ret := om.os_Getenv(ctx, runtime.String("TEST"))
	if ret.String(ctx) != exp {
		t.Errorf("expected '%s', got '%s'", exp, ret.String(ctx))
	}
}

func TestOsGetwd(t *testing.T) {
	ctx := context.Background()
	ktx := runtime.NewKtx(nil, nil)
	om := new(OsMod)
	om.SetKtx(ktx)
	exp, e := os.Getwd()
	if e != nil {
		panic(e)
	}
	ret := om.os_Getwd(ctx)
	if ret.String(ctx) != exp {
		t.Errorf("expected '%s', got '%s'", exp, ret.String(ctx))
	}
}

func TestOsReadFile(t *testing.T) {
	ctx := context.Background()
	ktx := runtime.NewKtx(nil, nil)
	om := new(OsMod)
	om.SetKtx(ktx)
	b, e := ioutil.ReadFile("./testdata/readfile.txt")
	if e != nil {
		panic(e)
	}
	exp := string(b)
	ret := om.os_ReadFile(ctx, runtime.String("./testdata/readfile.txt"))
	if ret.String(ctx) != exp {
		t.Errorf("expected '%s', got '%s'", exp, ret.String(ctx))
	}
}

func TestOsWriteFile(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		src []runtime.Val
		exp string
	}{
		0: {
			exp: "",
		},
		1: {
			src: []runtime.Val{runtime.String("hello")},
			exp: "hello",
		},
		2: {
			src: []runtime.Val{runtime.String("string"), runtime.Number(3), runtime.Bool(true),
				runtime.Nil, runtime.Number(1.23)},
			exp: "string3truenil1.23",
		},
	}
	fn := "./testdata/writefile.txt"
	ktx := runtime.NewKtx(nil, nil)
	om := new(OsMod)
	om.SetKtx(ktx)
	for i, c := range cases {
		args := append([]runtime.Val{runtime.String(fn)}, c.src...)
		ret := om.os_WriteFile(ctx, args...)
		b, e := ioutil.ReadFile(fn)
		if e != nil {
			panic(e)
		}
		got := string(b)
		if ret.Int(ctx) != int64(len(c.exp)) {
			t.Errorf("[%d] - expected %d, got %d", i, len(c.exp), ret.Int(ctx))
		}
		if got != c.exp {
			t.Errorf("[%d] - expected '%s', got '%s'", i, c.exp, got)
		}
	}
}

func TestOsMkRemRenReadDir(t *testing.T) {
	ctx := context.Background()
	ktx := runtime.NewKtx(nil, nil)
	om := new(OsMod)
	om.SetKtx(ktx)
	// First create directories
	d1, d2 := "./testdata/d1", "./testdata/d2/d3"
	om.os_Mkdir(ctx, runtime.String(d1), runtime.String(d2))
	// Check that they exist
	if _, e := os.Stat(d1); os.IsNotExist(e) {
		t.Errorf("expected d1 to be created, got %s", e)
	} else if e != nil {
		panic(e)
	}
	if _, e := os.Stat(d2); os.IsNotExist(e) {
		t.Errorf("expected d2 to be created, got %s", e)
	} else if e != nil {
		panic(e)
	}
	// Create a file
	fn := filepath.Join(d2, "test.txt")
	om.os_WriteFile(ctx, runtime.String(fn), runtime.String("hi"))
	// Read the dir
	ret := om.os_ReadDir(ctx, runtime.String(d2))
	ob := ret.(runtime.Object)
	if ob.Len(ctx).Int(ctx) != 1 {
		t.Errorf("expected read dir to return 1 file, got %d", ob.Len(ctx).Int(ctx))
	}
	v := ob.Get(runtime.Number(0))
	ob = v.(runtime.Object)
	if s := ob.Get(runtime.String("Name")); s.String(ctx) != "test.txt" {
		t.Errorf("expected read file to be 'test.txt', got %s", s)
	}
	// Remove the first dir
	om.os_Remove(ctx, runtime.String(d1))
	if _, e := os.Stat(d1); !os.IsNotExist(e) {
		t.Errorf("expected d1 to be deleted, got %s", e)
	}
	// Remove all for the second dir and file
	d2 = filepath.Join(d2, "..")
	om.os_RemoveAll(ctx, runtime.String(d2))
	if _, e := os.Stat(d2); !os.IsNotExist(e) {
		t.Errorf("expected d2 to be deleted, got %s", e)
	}
}
