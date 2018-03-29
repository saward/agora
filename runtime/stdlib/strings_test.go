package stdlib

import (
	"context"
	"testing"

	"github.com/bobg/agora/runtime"
)

func TestStringsMatches(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		args []runtime.Val
		exp  [][]string
	}{
		0: {
			args: []runtime.Val{
				runtime.String("this is a string"),
				runtime.String(`^.+$`),
			},
			exp: [][]string{
				0: []string{
					0: "this is a string",
				},
			},
		},
		1: {
			args: []runtime.Val{
				runtime.String("this is a string"),
				runtime.String(".*?(is)"),
			},
			exp: [][]string{
				0: []string{
					0: "this",
					1: "is",
				},
				1: []string{
					0: " is",
					1: "is",
				},
			},
		},
		2: {
			args: []runtime.Val{
				runtime.String("what whatever who where"),
				runtime.String(`(w.)\w+`),
				runtime.Number(2),
			},
			exp: [][]string{
				0: []string{
					0: "what",
					1: "wh",
				},
				1: []string{
					0: "whatever",
					1: "wh",
				},
			},
		},
	}
	ktx := runtime.NewKtx(nil, nil)
	sm := new(StringsMod)
	sm.SetKtx(ktx)
	for i, c := range cases {
		ret := sm.strings_Matches(ctx, c.args...)
		ob := ret.(runtime.Object)
		if int64(len(c.exp)) != ob.Len(ctx).Int(ctx) {
			t.Errorf("[%d] - expected %d matches, got %d", i, len(c.exp), ob.Len(ctx).Int(ctx))
		} else {
			for j := int64(0); j < ob.Len(ctx).Int(ctx); j++ {
				// For each match, there's 0..n number of matches (0 is the full match)
				mtch := ob.Get(runtime.Number(j))
				mo := mtch.(runtime.Object)
				if int64(len(c.exp[j])) != mo.Len(ctx).Int(ctx) {
					t.Errorf("[%d] - expected %d groups in match %d, got %d", i, len(c.exp[j]), j, mo.Len(ctx).Int(ctx))
				} else {
					for k := int64(0); k < mo.Len(ctx).Int(ctx); k++ {
						grp := mo.Get(runtime.Number(k))
						gro := grp.(runtime.Object)
						st := gro.Get(runtime.String("Start"))
						e := gro.Get(runtime.String("End"))
						if e.Int(ctx) != st.Int(ctx)+int64(len(c.exp[j][k])) {
							t.Errorf("[%d] - expected end %d for group %d of match %d, got %d", i, st.Int(ctx)+int64(len(c.exp[j][k])), k, j, e.Int(ctx))
						}
						s := gro.Get(runtime.String("Text"))
						if s.String(ctx) != c.exp[j][k] {
							t.Errorf("[%d] - expected text '%s' for group %d of match %d, got '%s'", i, c.exp[j][k], k, j, s)
						}
					}
				}
			}
		}
	}
}

func TestStringsToUpper(t *testing.T) {
	ctx := context.Background()
	ktx := runtime.NewKtx(nil, nil)
	sm := new(StringsMod)
	sm.SetKtx(ktx)
	ret := sm.strings_ToUpper(ctx, runtime.String("this"), runtime.String("Is"), runtime.String("A"), runtime.String("... strInG"))
	exp := "THISISA... STRING"
	if ret.String(ctx) != exp {
		t.Errorf("expected %s, got %s", exp, ret)
	}
}

func TestStringsToLower(t *testing.T) {
	ctx := context.Background()
	ktx := runtime.NewKtx(nil, nil)
	sm := new(StringsMod)
	sm.SetKtx(ktx)
	ret := sm.strings_ToLower(ctx, runtime.String("this"), runtime.String("Is"), runtime.String("A"), runtime.String("... strInG"))
	exp := "thisisa... string"
	if ret.String(ctx) != exp {
		t.Errorf("expected %s, got %s", exp, ret)
	}
}

func TestStringsHasPrefix(t *testing.T) {
	ctx := context.Background()
	ktx := runtime.NewKtx(nil, nil)
	sm := new(StringsMod)
	sm.SetKtx(ktx)
	ret := sm.strings_HasPrefix(ctx, runtime.String("what prefix?"), runtime.String("no"), runtime.Nil, runtime.Number(3), runtime.String("wh"))
	if !ret.Bool(ctx) {
		t.Errorf("expected true, got false")
	}
	ret = sm.strings_HasPrefix(ctx, runtime.String("what prefix?"), runtime.String("no"), runtime.Nil, runtime.Number(3), runtime.String("hw"))
	if ret.Bool(ctx) {
		t.Errorf("expected false, got true")
	}
}

func TestStringsHasSuffix(t *testing.T) {
	ctx := context.Background()
	ktx := runtime.NewKtx(nil, nil)
	sm := new(StringsMod)
	sm.SetKtx(ktx)
	ret := sm.strings_HasSuffix(ctx, runtime.String("suffix, you say"), runtime.String("ay"), runtime.Nil, runtime.Number(3), runtime.String("wh"))
	if !ret.Bool(ctx) {
		t.Errorf("expected true, got false")
	}
	ret = sm.strings_HasSuffix(ctx, runtime.String("suffix, you say"), runtime.String("no"), runtime.Nil, runtime.Number(3), runtime.String("hw"))
	if ret.Bool(ctx) {
		t.Errorf("expected false, got true")
	}
}

func TestStringsByteAt(t *testing.T) {
	ctx := context.Background()
	ktx := runtime.NewKtx(nil, nil)
	sm := new(StringsMod)
	sm.SetKtx(ktx)
	src := "some string"
	ix := 0
	ret := sm.strings_ByteAt(ctx, runtime.String(src), runtime.Number(ix))
	if ret.String(ctx) != string(src[ix]) {
		t.Errorf("expected byte %s at index %d, got %s", string(src[ix]), ix, ret)
	}
	ix = 3
	ret = sm.strings_ByteAt(ctx, runtime.String(src), runtime.Number(ix))
	if ret.String(ctx) != string(src[ix]) {
		t.Errorf("expected byte %s at index %d, got %s", string(src[ix]), ix, ret)
	}
	ix = 22
	ret = sm.strings_ByteAt(ctx, runtime.String(src), runtime.Number(ix))
	if ret.String(ctx) != "" {
		t.Errorf("expected byte %s at index %d, got %s", "", ix, ret)
	}
}

func TestStringsConcat(t *testing.T) {
	ctx := context.Background()
	ktx := runtime.NewKtx(nil, nil)
	sm := new(StringsMod)
	sm.SetKtx(ktx)
	ret := sm.strings_Concat(ctx, runtime.String("hello"), runtime.Number(12), runtime.Bool(true), runtime.String("end"))
	exp := "hello12trueend"
	if ret.String(ctx) != exp {
		t.Errorf("expected %s, got %s", exp, ret)
	}
}

func TestStringsContains(t *testing.T) {
	ctx := context.Background()
	ktx := runtime.NewKtx(nil, nil)
	sm := new(StringsMod)
	sm.SetKtx(ktx)
	ret := sm.strings_Contains(ctx, runtime.String("contains something"), runtime.String("what"), runtime.Nil, runtime.Number(3), runtime.String("some"))
	if !ret.Bool(ctx) {
		t.Errorf("expected true, got false")
	}
	ret = sm.strings_Contains(ctx, runtime.String("contains something"), runtime.String("no"), runtime.Nil, runtime.Number(3), runtime.String("hw"))
	if ret.Bool(ctx) {
		t.Errorf("expected false, got true")
	}
}

func TestStringsIndex(t *testing.T) {
	ctx := context.Background()
	ktx := runtime.NewKtx(nil, nil)
	sm := new(StringsMod)
	sm.SetKtx(ktx)
	ret := sm.strings_Index(ctx, runtime.String("agora"), runtime.String("arg"), runtime.Nil, runtime.Number(3), runtime.String("go"))
	exp := 1
	if ret.Int(ctx) != int64(exp) {
		t.Errorf("expected %d, got %d", exp, ret.Int(ctx))
	}
	ret = sm.strings_Index(ctx, runtime.String("agora"), runtime.Number(2), runtime.String("arg"), runtime.Nil, runtime.Number(3), runtime.String("go"))
	exp = -1
	if ret.Int(ctx) != int64(exp) {
		t.Errorf("expected %d, got %d", exp, ret.Int(ctx))
	}
}

func TestStringsLastIndex(t *testing.T) {
	ctx := context.Background()
	ktx := runtime.NewKtx(nil, nil)
	sm := new(StringsMod)
	sm.SetKtx(ktx)
	ret := sm.strings_LastIndex(ctx, runtime.String("agoragore"), runtime.String("arg"), runtime.Nil, runtime.Number(3), runtime.String("go"))
	exp := 5
	if ret.Int(ctx) != int64(exp) {
		t.Errorf("expected %d, got %d", exp, ret.Int(ctx))
	}
	ret = sm.strings_Index(ctx, runtime.String("agoragore"), runtime.Number(6), runtime.String("arg"), runtime.Nil, runtime.Number(3), runtime.String("go"))
	exp = -1
	if ret.Int(ctx) != int64(exp) {
		t.Errorf("expected %d, got %d", exp, ret.Int(ctx))
	}
}

func TestStringsSlice(t *testing.T) {
	ctx := context.Background()
	ktx := runtime.NewKtx(nil, nil)
	sm := new(StringsMod)
	sm.SetKtx(ktx)
	ret := sm.strings_Slice(ctx, runtime.String("agora"), runtime.Number(2))
	exp := "ora"
	if ret.String(ctx) != exp {
		t.Errorf("expected %s, got %s", exp, ret)
	}
	ret = sm.strings_Slice(ctx, runtime.String("agora"), runtime.Number(2), runtime.Number(4))
	exp = "or"
	if ret.String(ctx) != exp {
		t.Errorf("expected %s, got %s", exp, ret)
	}
}

func TestStringsSplit(t *testing.T) {
	ctx := context.Background()
	ktx := runtime.NewKtx(nil, nil)
	sm := new(StringsMod)
	sm.SetKtx(ktx)
	ret := sm.strings_Split(ctx, runtime.String("aa:bb::dd"), runtime.String(":"))
	ob := ret.(runtime.Object)
	exp := []string{"aa", "bb", "", "dd"}
	if l := ob.Len(ctx).Int(ctx); l != int64(len(exp)) {
		t.Errorf("expected split length of %d, got %d", len(exp), l)
	}
	for i, v := range exp {
		got := ob.Get(runtime.Number(i))
		if got.String(ctx) != v {
			t.Errorf("expected split index %d to be %s, got %s", i, v, got)
		}
	}
	ret = sm.strings_Split(ctx, runtime.String("aa:bb::dd:ee:"), runtime.String(":"), runtime.Number(2))
	ob = ret.(runtime.Object)
	exp = []string{"aa", "bb::dd:ee:"}
	if l := ob.Len(ctx).Int(ctx); l != int64(len(exp)) {
		t.Errorf("expected split length of %d, got %d", len(exp), l)
	}
	for i, v := range exp {
		got := ob.Get(runtime.Number(i))
		if got.String(ctx) != v {
			t.Errorf("expected split index %d to be %s, got %s", i, v, got)
		}
	}
}

func TestStringsJoin(t *testing.T) {
	ctx := context.Background()
	ktx := runtime.NewKtx(nil, nil)
	sm := new(StringsMod)
	sm.SetKtx(ktx)
	parts := []string{"this", "is", "", "it!"}
	ob := runtime.NewObject()
	for i, v := range parts {
		ob.Set(runtime.Number(i), runtime.String(v))
	}
	ret := sm.strings_Join(ctx, ob)
	exp := "thisisit!"
	if ret.String(ctx) != exp {
		t.Errorf("expected %s, got %s", exp, ret)
	}
	ret = sm.strings_Join(ctx, ob, runtime.String("--"))
	exp = "this--is----it!"
	if ret.String(ctx) != exp {
		t.Errorf("expected %s, got %s", exp, ret)
	}
}

func TestStringsReplace(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		args []runtime.Val
		exp  string
	}{
		0: {
			args: []runtime.Val{
				runtime.String("this is the source"),
				runtime.String("th"),
			},
			exp: "is is e source",
		},
		1: {
			args: []runtime.Val{
				runtime.String("this is the source"),
				runtime.String("th"),
				runtime.Number(1),
			},
			exp: "is is the source",
		},
		2: {
			args: []runtime.Val{
				runtime.String("this is the source"),
				runtime.String("t"),
				runtime.String("T"),
			},
			exp: "This is The source",
		},
		3: {
			args: []runtime.Val{
				runtime.String("this is the source"),
				runtime.String("t"),
				runtime.String("T"),
				runtime.Number(1),
			},
			exp: "This is the source",
		},
	}
	ktx := runtime.NewKtx(nil, nil)
	sm := new(StringsMod)
	sm.SetKtx(ktx)
	for i, c := range cases {
		ret := sm.strings_Replace(ctx, c.args...)
		if ret.String(ctx) != c.exp {
			t.Errorf("[%d] - expected %s, got %s", i, c.exp, ret)
		}
	}
}

func TestStringsTrim(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		args []runtime.Val
		exp  string
	}{
		0: {
			args: []runtime.Val{
				runtime.String(" "),
			},
			exp: "",
		},
		1: {
			args: []runtime.Val{
				runtime.String("\n  \t   hi \r"),
			},
			exp: "hi",
		},
		2: {
			args: []runtime.Val{
				runtime.String("xoxolovexox"),
				runtime.String("xo"),
			},
			exp: "love",
		},
	}
	ktx := runtime.NewKtx(nil, nil)
	sm := new(StringsMod)
	sm.SetKtx(ktx)
	for i, c := range cases {
		ret := sm.strings_Trim(ctx, c.args...)
		if ret.String(ctx) != c.exp {
			t.Errorf("[%d] - expected %s, got %s", i, c.exp, ret)
		}
	}
}
