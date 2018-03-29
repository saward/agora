package runtime

import (
	"context"
	"fmt"
)

type builtinMod struct {
	ktx *Kontext
	ob  Object
}

func (b *builtinMod) ID() string {
	return "<builtin>"
}

func (b *builtinMod) Run(_ ...Val) (v Val, err error) {
	defer PanicToError(&err)
	if b.ob == nil {
		b.ob = NewObject()
		b.ob.Set(String("import"), NewNativeFunc(b.ktx, "import", b._import))
		b.ob.Set(String("panic"), NewNativeFunc(b.ktx, "panic", b._panic))
		b.ob.Set(String("recover"), NewNativeFunc(b.ktx, "recover", b._recover))
		b.ob.Set(String("len"), NewNativeFunc(b.ktx, "len", b._len))
		b.ob.Set(String("keys"), NewNativeFunc(b.ktx, "keys", b._keys))
		b.ob.Set(String("number"), NewNativeFunc(b.ktx, "number", b._number))
		b.ob.Set(String("string"), NewNativeFunc(b.ktx, "string", b._string))
		b.ob.Set(String("bool"), NewNativeFunc(b.ktx, "bool", b._bool))
		b.ob.Set(String("type"), NewNativeFunc(b.ktx, "type", b._type))
		b.ob.Set(String("status"), NewNativeFunc(b.ktx, "status", b._status))
		b.ob.Set(String("reset"), NewNativeFunc(b.ktx, "reset", b._reset))
	}
	return b.ob, nil
}

func (b *builtinMod) SetKtx(c *Kontext) {
	b.ktx = c
}

func (b *builtinMod) _import(ctx context.Context, args ...Val) Val {
	ExpectAtLeastNArgs(1, args)
	m, err := b.ktx.Load(args[0].String(ctx))
	if err != nil {
		panic(err)
	}
	v, err := m.Run(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

func (b *builtinMod) _panic(ctx context.Context, args ...Val) Val {
	ExpectAtLeastNArgs(1, args)
	if args[0].Bool(ctx) {
		panic(args[0])
	}
	return Nil
}

func (b *builtinMod) _recover(ctx context.Context, args ...Val) (ret Val) {
	// Do not catch panics if args are invalid
	ExpectAtLeastNArgs(1, args)
	// Catch panics in running the function. Cannot use PanicToError, because
	// it needs the true type of the panic'd value.
	ret = Nil
	defer func() {
		if err := recover(); err != nil {
			switch v := err.(type) {
			case Val:
				ret = v
			case error:
				ret = String(v.Error())
			default:
				ret = String(fmt.Sprintf("%v", v))
			}
		}
	}()
	// The value must be a function
	f, ok := args[0].(Func)
	if !ok {
		panic(NewTypeError(Type(args[0]), "", "recover"))
	}
	// Return value is discarded, because recover returns the error, if any, or Nil.
	// The function to run in recovery mode must be a closure or assign its return
	// value to an outer-scope variable.

	// TODO : This would lose the `this` keyword in case of recover being called
	// on an object's method.
	f.Call(ctx, Nil, args[1:]...)
	return ret
}

func (b *builtinMod) _len(ctx context.Context, args ...Val) Val {
	ExpectAtLeastNArgs(1, args)
	switch v := args[0].(type) {
	case Object:
		return v.Len(ctx)
	case null:
		return Number(0)
	default:
		return Number(len(v.String(ctx)))
	}
}

func (b *builtinMod) _keys(ctx context.Context, args ...Val) Val {
	ExpectAtLeastNArgs(1, args)
	ob := args[0].(Object)
	return ob.Keys(ctx)
}

func (b *builtinMod) _number(ctx context.Context, args ...Val) Val {
	ExpectAtLeastNArgs(1, args)
	return Number(args[0].Float(ctx))
}

func (b *builtinMod) _string(ctx context.Context, args ...Val) Val {
	ExpectAtLeastNArgs(1, args)
	return String(args[0].String(ctx))
}

func (b *builtinMod) _bool(ctx context.Context, args ...Val) Val {
	ExpectAtLeastNArgs(1, args)
	return Bool(args[0].Bool(ctx))
}

func (b *builtinMod) _type(_ context.Context, args ...Val) Val {
	ExpectAtLeastNArgs(1, args)
	return String(Type(args[0]))
}

func (b *builtinMod) _status(_ context.Context, args ...Val) Val {
	ExpectAtLeastNArgs(1, args)
	if v, ok := args[0].(*agoraFuncVal); ok {
		// If v is in the frame stack, return `running`
		// If v.coroState is not nil, return `suspended`
		// Else return empty string
		return String(v.status())
	} else if _, ok := args[0].(Func); !ok {
		// Can only be called on a Func
		panic(NewTypeError(Type(args[0]), "", "status"))
	}
	return String("")
}

func (b *builtinMod) _reset(_ context.Context, args ...Val) Val {
	ExpectAtLeastNArgs(1, args)
	if v, ok := args[0].(*agoraFuncVal); ok {
		v.reset()
	} else if _, ok := args[0].(Func); !ok {
		// Can only be called on a Func
		panic(NewTypeError(Type(args[0]), "", "reset"))
	}
	return Nil
}
