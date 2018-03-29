package stdlib

import (
	"context"
	"math"
	"math/rand"

	"github.com/bobg/agora/runtime"
)

// The math module, as documented in
// https://github.com/bobg/agora/wiki/Standard-library
type MathMod struct {
	ktx *runtime.Kontext
	ob  runtime.Object
}

func (m *MathMod) ID() string {
	return "math"
}

func (m *MathMod) Run(_ context.Context, _ ...runtime.Val) (v runtime.Val, err error) {
	defer runtime.PanicToError(&err)
	if m.ob == nil {
		// Prepare the object
		m.ob = runtime.NewObject()
		m.ob.Set(runtime.String("Pi"), runtime.Number(math.Pi))
		m.ob.Set(runtime.String("Abs"), runtime.NewNativeFunc(m.ktx, "math.Abs", m.math_Abs))
		m.ob.Set(runtime.String("Acos"), runtime.NewNativeFunc(m.ktx, "math.Acos", m.math_Acos))
		m.ob.Set(runtime.String("Acosh"), runtime.NewNativeFunc(m.ktx, "math.Acosh", m.math_Acosh))
		m.ob.Set(runtime.String("Asin"), runtime.NewNativeFunc(m.ktx, "math.Asin", m.math_Asin))
		m.ob.Set(runtime.String("Asinh"), runtime.NewNativeFunc(m.ktx, "math.Asinh", m.math_Asinh))
		m.ob.Set(runtime.String("Atan"), runtime.NewNativeFunc(m.ktx, "math.Atan", m.math_Atan))
		m.ob.Set(runtime.String("Atan2"), runtime.NewNativeFunc(m.ktx, "math.Atan2", m.math_Atan2))
		m.ob.Set(runtime.String("Atanh"), runtime.NewNativeFunc(m.ktx, "math.Atanh", m.math_Atanh))
		m.ob.Set(runtime.String("Ceil"), runtime.NewNativeFunc(m.ktx, "math.Ceil", m.math_Ceil))
		m.ob.Set(runtime.String("Cos"), runtime.NewNativeFunc(m.ktx, "math.Cos", m.math_Cos))
		m.ob.Set(runtime.String("Cosh"), runtime.NewNativeFunc(m.ktx, "math.Cosh", m.math_Cosh))
		m.ob.Set(runtime.String("Exp"), runtime.NewNativeFunc(m.ktx, "math.Exp", m.math_Exp))
		m.ob.Set(runtime.String("Floor"), runtime.NewNativeFunc(m.ktx, "math.Floor", m.math_Floor))
		m.ob.Set(runtime.String("Inf"), runtime.NewNativeFunc(m.ktx, "math.Inf", m.math_Inf))
		m.ob.Set(runtime.String("IsInf"), runtime.NewNativeFunc(m.ktx, "math.IsInf", m.math_IsInf))
		m.ob.Set(runtime.String("IsNaN"), runtime.NewNativeFunc(m.ktx, "math.IsNaN", m.math_IsNaN))
		m.ob.Set(runtime.String("Max"), runtime.NewNativeFunc(m.ktx, "math.Max", m.math_Max))
		m.ob.Set(runtime.String("Min"), runtime.NewNativeFunc(m.ktx, "math.Min", m.math_Min))
		m.ob.Set(runtime.String("NaN"), runtime.NewNativeFunc(m.ktx, "math.NaN", m.math_NaN))
		m.ob.Set(runtime.String("Pow"), runtime.NewNativeFunc(m.ktx, "math.Pow", m.math_Pow))
		m.ob.Set(runtime.String("Sin"), runtime.NewNativeFunc(m.ktx, "math.Sin", m.math_Sin))
		m.ob.Set(runtime.String("Sinh"), runtime.NewNativeFunc(m.ktx, "math.Sinh", m.math_Sinh))
		m.ob.Set(runtime.String("Sqrt"), runtime.NewNativeFunc(m.ktx, "math.Sqrt", m.math_Sqrt))
		m.ob.Set(runtime.String("Tan"), runtime.NewNativeFunc(m.ktx, "math.Tan", m.math_Tan))
		m.ob.Set(runtime.String("Tanh"), runtime.NewNativeFunc(m.ktx, "math.Tanh", m.math_Tanh))
		m.ob.Set(runtime.String("RandSeed"), runtime.NewNativeFunc(m.ktx, "math.RandSeed", m.math_RandSeed))
		m.ob.Set(runtime.String("Rand"), runtime.NewNativeFunc(m.ktx, "math.Rand", m.math_Rand))
	}
	return m.ob, nil
}

func (m *MathMod) SetKtx(ktx *runtime.Kontext) {
	m.ktx = ktx
}

func (m *MathMod) math_Abs(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(1, args)
	return runtime.Number(math.Abs(args[0].Float(ctx)))
}

func (m *MathMod) math_Acos(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(1, args)
	return runtime.Number(math.Acos(args[0].Float(ctx)))
}

func (m *MathMod) math_Acosh(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(1, args)
	return runtime.Number(math.Acosh(args[0].Float(ctx)))
}

func (m *MathMod) math_Asin(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(1, args)
	return runtime.Number(math.Asin(args[0].Float(ctx)))
}

func (m *MathMod) math_Asinh(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(1, args)
	return runtime.Number(math.Asinh(args[0].Float(ctx)))
}

func (m *MathMod) math_Atan(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(1, args)
	return runtime.Number(math.Atan(args[0].Float(ctx)))
}

func (m *MathMod) math_Atan2(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(2, args)
	return runtime.Number(math.Atan2(args[0].Float(ctx), args[1].Float(ctx)))
}

func (m *MathMod) math_Atanh(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(1, args)
	return runtime.Number(math.Atanh(args[0].Float(ctx)))
}

func (m *MathMod) math_Ceil(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(1, args)
	return runtime.Number(math.Ceil(args[0].Float(ctx)))
}

func (m *MathMod) math_Cos(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(1, args)
	return runtime.Number(math.Cos(args[0].Float(ctx)))
}

func (m *MathMod) math_Cosh(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(1, args)
	return runtime.Number(math.Cosh(args[0].Float(ctx)))
}

func (m *MathMod) math_Exp(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(1, args)
	return runtime.Number(math.Exp(args[0].Float(ctx)))
}

func (m *MathMod) math_Floor(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(1, args)
	return runtime.Number(math.Floor(args[0].Float(ctx)))
}

func (m *MathMod) math_Inf(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(1, args)
	return runtime.Number(math.Inf(int(args[0].Int(ctx))))
}

func (m *MathMod) math_IsInf(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(2, args)
	return runtime.Bool(math.IsInf(args[0].Float(ctx), int(args[1].Int(ctx))))
}

func (m *MathMod) math_IsNaN(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(1, args)
	return runtime.Bool(math.IsNaN(args[0].Float(ctx)))
}

func (m *MathMod) math_Max(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(2, args)
	max := args[len(args)-1].Float(ctx)
	for i := len(args) - 2; i >= 0; i-- {
		max = math.Max(max, args[i].Float(ctx))
	}
	return runtime.Number(max)
}

func (m *MathMod) math_Min(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(2, args)
	min := args[len(args)-1].Float(ctx)
	for i := len(args) - 2; i >= 0; i-- {
		min = math.Min(min, args[i].Float(ctx))
	}
	return runtime.Number(min)
}

func (m *MathMod) math_NaN(ctx context.Context, _ ...runtime.Val) runtime.Val {
	return runtime.Number(math.NaN())
}

func (m *MathMod) math_Pow(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(2, args)
	return runtime.Number(math.Pow(args[0].Float(ctx), args[1].Float(ctx)))
}

func (m *MathMod) math_Sin(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(1, args)
	return runtime.Number(math.Sin(args[0].Float(ctx)))
}

func (m *MathMod) math_Sinh(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(1, args)
	return runtime.Number(math.Sinh(args[0].Float(ctx)))
}

func (m *MathMod) math_Sqrt(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(1, args)
	return runtime.Number(math.Sqrt(args[0].Float(ctx)))
}

func (m *MathMod) math_Tan(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(1, args)
	return runtime.Number(math.Tan(args[0].Float(ctx)))
}

func (m *MathMod) math_Tanh(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(1, args)
	return runtime.Number(math.Tanh(args[0].Float(ctx)))
}

func (m *MathMod) math_RandSeed(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(1, args)
	rand.Seed(args[0].Int(ctx))
	return runtime.Nil
}

func (m *MathMod) math_Rand(ctx context.Context, args ...runtime.Val) runtime.Val {
	switch len(args) {
	case 0:
		return runtime.Number(rand.Int())
	case 1:
		return runtime.Number(rand.Intn(int(args[0].Int(ctx))))
	default:
		low := args[0].Int(ctx)
		high := args[1].Int(ctx)
		n := rand.Intn(int(high - low))
		return runtime.Number(int64(n) + low)
	}
}
