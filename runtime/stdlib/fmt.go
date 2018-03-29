package stdlib

import (
	"bufio"
	"context"
	"fmt"

	"github.com/bobg/agora/runtime"
)

// The fmt module, as documented in
// https://github.com/bobg/agora/wiki/Standard-library
type FmtMod struct {
	ktx *runtime.Kontext
	ob  runtime.Object
}

func (f *FmtMod) ID() string {
	return "fmt"
}

func (f *FmtMod) Run(_ context.Context, _ ...runtime.Val) (v runtime.Val, err error) {
	defer runtime.PanicToError(&err)
	if f.ob == nil {
		// Prepare the object
		f.ob = runtime.NewObject()
		f.ob.Set(runtime.String("Print"), runtime.NewNativeFunc(f.ktx, "fmt.Print", f.fmt_Print))
		f.ob.Set(runtime.String("Println"), runtime.NewNativeFunc(f.ktx, "fmt.Println", f.fmt_Println))
		f.ob.Set(runtime.String("Scanln"), runtime.NewNativeFunc(f.ktx, "fmt.Scanln", f.fmt_Scanln))
		f.ob.Set(runtime.String("Scanint"), runtime.NewNativeFunc(f.ktx, "fmt.Scanint", f.fmt_Scanint))
	}
	return f.ob, nil
}

func (f *FmtMod) SetKtx(c *runtime.Kontext) {
	f.ktx = c
}

func toStringIface(ctx context.Context, args []runtime.Val) []interface{} {
	var ifs []interface{}

	if len(args) > 0 {
		ifs = make([]interface{}, len(args))
		for i, v := range args {
			ifs[i] = v.String(ctx)
		}
	}
	return ifs
}

func (f *FmtMod) fmt_Print(ctx context.Context, args ...runtime.Val) runtime.Val {
	ifs := toStringIface(ctx, args)
	n, err := fmt.Fprint(f.ktx.Stdout, ifs...)
	if err != nil {
		panic(err)
	}
	return runtime.Number(n)
}

func (f *FmtMod) fmt_Println(ctx context.Context, args ...runtime.Val) runtime.Val {
	ifs := toStringIface(ctx, args)
	n, err := fmt.Fprintln(f.ktx.Stdout, ifs...)
	if err != nil {
		panic(err)
	}
	return runtime.Number(n)
}

func (f *FmtMod) fmt_Scanln(ctx context.Context, args ...runtime.Val) runtime.Val {
	var (
		b, l []byte
		e    error
		pre  bool
	)
	r := bufio.NewReader(f.ktx.Stdin)
	for l, pre, e = r.ReadLine(); pre && e == nil; l, pre, e = r.ReadLine() {
		b = append(b, l...)
	}
	if e != nil {
		panic(e)
	}
	b = append(b, l...)
	return runtime.String(b)
}

func (f *FmtMod) fmt_Scanint(ctx context.Context, args ...runtime.Val) runtime.Val {
	var i int
	if _, e := fmt.Fscanf(f.ktx.Stdin, "%d", &i); e != nil {
		panic(e)
	}
	return runtime.Number(i)
}
