package stdlib

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/bobg/agora/runtime"
)

func TestPi(t *testing.T) {
	ctx := context.Background()
	ktx := runtime.NewKtx(nil, nil)
	mm := new(MathMod)
	mm.SetKtx(ktx)
	ob, err := mm.Run(ctx)
	if err != nil {
		panic(err)
	}
	ret := ob.(runtime.Object).Get(runtime.String("Pi"))
	exp := math.Pi
	if ret.Float(ctx) != exp {
		t.Errorf("expected %f, got %f", exp, ret.Float(ctx))
	}
}

func TestMax(t *testing.T) {
	ctx := context.Background()
	ktx := runtime.NewKtx(nil, nil)
	mm := new(MathMod)
	mm.SetKtx(ktx)

	cases := []struct {
		src []runtime.Val
		exp runtime.Val
	}{
		0: {
			src: []runtime.Val{runtime.Number(3), runtime.Number(0), runtime.Number(-12.74), runtime.Number(1)},
			exp: runtime.Number(3),
		},
		1: {
			src: []runtime.Val{runtime.String("24"), runtime.Bool(true), runtime.Number(12.74)},
			exp: runtime.Number(24),
		},
		2: {
			src: []runtime.Val{runtime.Number(0), runtime.String("0")},
			exp: runtime.Number(0),
		},
	}

	for i, c := range cases {
		ret := mm.math_Max(ctx, c.src...)
		if ret != c.exp {
			t.Errorf("[%d] - expected %f, got %f", i, c.exp.Float(ctx), ret.Float(ctx))
		}
	}
}

func TestMin(t *testing.T) {
	ctx := context.Background()
	ktx := runtime.NewKtx(nil, nil)
	mm := new(MathMod)
	mm.SetKtx(ktx)

	cases := []struct {
		src []runtime.Val
		exp runtime.Val
	}{
		0: {
			src: []runtime.Val{runtime.Number(3), runtime.Number(0), runtime.Number(-12.74), runtime.Number(1)},
			exp: runtime.Number(-12.74),
		},
		1: {
			src: []runtime.Val{runtime.String("24"), runtime.Bool(true), runtime.Number(12.74)},
			exp: runtime.Number(1),
		},
		2: {
			src: []runtime.Val{runtime.Number(0), runtime.String("0")},
			exp: runtime.Number(0),
		},
	}

	for i, c := range cases {
		ret := mm.math_Min(ctx, c.src...)
		if ret != c.exp {
			t.Errorf("[%d] - expected %f, got %f", i, c.exp.Float(ctx), ret.Float(ctx))
		}
	}
}

func TestRand(t *testing.T) {
	ctx := context.Background()
	ktx := runtime.NewKtx(nil, nil)
	mm := new(MathMod)
	mm.SetKtx(ktx)

	mm.math_RandSeed(ctx, runtime.Number(time.Now().UnixNano()))
	// no-arg form
	ret := mm.math_Rand(ctx)
	if ret.Int(ctx) < 0 {
		t.Errorf("expected no-arg to procude non-negative value, got %d", ret.Int(ctx))
	}
	// one-arg form
	ret = mm.math_Rand(ctx, runtime.Number(10))
	if ret.Int(ctx) < 0 || ret.Int(ctx) >= 10 {
		t.Errorf("expected one-arg to produce non-negative value lower than 10, got %d", ret.Int(ctx))
	}
	// two-args form
	ret = mm.math_Rand(ctx, runtime.Number(3), runtime.Number(9))
	if ret.Int(ctx) < 3 || ret.Int(ctx) >= 9 {
		t.Errorf("expected two-args to produce value >= 3 and < 9, got %d", ret.Int(ctx))
	}
}

func TestMathAbs(t *testing.T) {
	ctx := context.Background()
	// This is just an interface to Go's function, so just a quick simple test
	ktx := runtime.NewKtx(nil, nil)
	mm := new(MathMod)
	mm.SetKtx(ktx)
	val := -3.5
	ret := mm.math_Abs(ctx, runtime.Number(val))
	exp := math.Abs(val)
	if ret.Float(ctx) != exp {
		t.Errorf("expected %f, got %f", exp, ret.Float(ctx))
	}
}

func TestMathAcos(t *testing.T) {
	ctx := context.Background()
	// This is just an interface to Go's function, so just a quick simple test
	ktx := runtime.NewKtx(nil, nil)
	mm := new(MathMod)
	mm.SetKtx(ktx)
	val := 0.12
	ret := mm.math_Acos(ctx, runtime.Number(val))
	exp := math.Acos(val)
	if ret.Float(ctx) != exp {
		t.Errorf("expected %f, got %f", exp, ret.Float(ctx))
	}
}

func TestMathAcosh(t *testing.T) {
	ctx := context.Background()
	// This is just an interface to Go's function, so just a quick simple test
	ktx := runtime.NewKtx(nil, nil)
	mm := new(MathMod)
	mm.SetKtx(ktx)
	val := 1.12
	ret := mm.math_Acosh(ctx, runtime.Number(val))
	exp := math.Acosh(val)
	if ret.Float(ctx) != exp {
		t.Errorf("expected %f, got %f", exp, ret.Float(ctx))
	}
}

func TestMathAsin(t *testing.T) {
	ctx := context.Background()
	// This is just an interface to Go's function, so just a quick simple test
	ktx := runtime.NewKtx(nil, nil)
	mm := new(MathMod)
	mm.SetKtx(ktx)
	val := 0.12
	ret := mm.math_Asin(ctx, runtime.Number(val))
	exp := math.Asin(val)
	if ret.Float(ctx) != exp {
		t.Errorf("expected %f, got %f", exp, ret.Float(ctx))
	}
}

func TestMathAsinh(t *testing.T) {
	ctx := context.Background()
	// This is just an interface to Go's function, so just a quick simple test
	ktx := runtime.NewKtx(nil, nil)
	mm := new(MathMod)
	mm.SetKtx(ktx)
	val := 0.12
	ret := mm.math_Asinh(ctx, runtime.Number(val))
	exp := math.Asinh(val)
	if ret.Float(ctx) != exp {
		t.Errorf("expected %f, got %f", exp, ret.Float(ctx))
	}
}

func TestMathAtan(t *testing.T) {
	ctx := context.Background()
	// This is just an interface to Go's function, so just a quick simple test
	ktx := runtime.NewKtx(nil, nil)
	mm := new(MathMod)
	mm.SetKtx(ktx)
	val := 0.12
	ret := mm.math_Atan(ctx, runtime.Number(val))
	exp := math.Atan(val)
	if ret.Float(ctx) != exp {
		t.Errorf("expected %f, got %f", exp, ret.Float(ctx))
	}
}

func TestMathAtan2(t *testing.T) {
	ctx := context.Background()
	// This is just an interface to Go's function, so just a quick simple test
	ktx := runtime.NewKtx(nil, nil)
	mm := new(MathMod)
	mm.SetKtx(ktx)
	val := 0.12
	val2 := 1.12
	ret := mm.math_Atan2(ctx, runtime.Number(val), runtime.Number(val2))
	exp := math.Atan2(val, val2)
	if ret.Float(ctx) != exp {
		t.Errorf("expected %f, got %f", exp, ret.Float(ctx))
	}
}

func TestMathAtanh(t *testing.T) {
	ctx := context.Background()
	// This is just an interface to Go's function, so just a quick simple test
	ktx := runtime.NewKtx(nil, nil)
	mm := new(MathMod)
	mm.SetKtx(ktx)
	val := 0.12
	ret := mm.math_Atanh(ctx, runtime.Number(val))
	exp := math.Atanh(val)
	if ret.Float(ctx) != exp {
		t.Errorf("expected %f, got %f", exp, ret.Float(ctx))
	}
}

func TestMathCeil(t *testing.T) {
	ctx := context.Background()
	// This is just an interface to Go's function, so just a quick simple test
	ktx := runtime.NewKtx(nil, nil)
	mm := new(MathMod)
	mm.SetKtx(ktx)
	val := 6.12
	ret := mm.math_Ceil(ctx, runtime.Number(val))
	exp := math.Ceil(val)
	if ret.Float(ctx) != exp {
		t.Errorf("expected %f, got %f", exp, ret.Float(ctx))
	}
}

func TestMathCos(t *testing.T) {
	ctx := context.Background()
	// This is just an interface to Go's function, so just a quick simple test
	ktx := runtime.NewKtx(nil, nil)
	mm := new(MathMod)
	mm.SetKtx(ktx)
	val := 0.12
	ret := mm.math_Cos(ctx, runtime.Number(val))
	exp := math.Cos(val)
	if ret.Float(ctx) != exp {
		t.Errorf("expected %f, got %f", exp, ret.Float(ctx))
	}
}

func TestMathCosh(t *testing.T) {
	ctx := context.Background()
	// This is just an interface to Go's function, so just a quick simple test
	ktx := runtime.NewKtx(nil, nil)
	mm := new(MathMod)
	mm.SetKtx(ktx)
	val := 0.12
	ret := mm.math_Cosh(ctx, runtime.Number(val))
	exp := math.Cosh(val)
	if ret.Float(ctx) != exp {
		t.Errorf("expected %f, got %f", exp, ret.Float(ctx))
	}
}

func TestMathExp(t *testing.T) {
	ctx := context.Background()
	// This is just an interface to Go's function, so just a quick simple test
	ktx := runtime.NewKtx(nil, nil)
	mm := new(MathMod)
	mm.SetKtx(ktx)
	val := 0.12
	ret := mm.math_Exp(ctx, runtime.Number(val))
	exp := math.Exp(val)
	if ret.Float(ctx) != exp {
		t.Errorf("expected %f, got %f", exp, ret.Float(ctx))
	}
}

func TestMathFloor(t *testing.T) {
	ctx := context.Background()
	// This is just an interface to Go's function, so just a quick simple test
	ktx := runtime.NewKtx(nil, nil)
	mm := new(MathMod)
	mm.SetKtx(ktx)
	val := 4.12
	ret := mm.math_Floor(ctx, runtime.Number(val))
	exp := math.Floor(val)
	if ret.Float(ctx) != exp {
		t.Errorf("expected %f, got %f", exp, ret.Float(ctx))
	}
}

func TestMathInf(t *testing.T) {
	ctx := context.Background()
	// This is just an interface to Go's function, so just a quick simple test
	ktx := runtime.NewKtx(nil, nil)
	mm := new(MathMod)
	mm.SetKtx(ktx)
	val := 1
	ret := mm.math_Inf(ctx, runtime.Number(val))
	exp := math.Inf(val)
	if ret.Float(ctx) != exp {
		t.Errorf("expected %f, got %f", exp, ret.Float(ctx))
	}
}

func TestMathIsInf(t *testing.T) {
	ctx := context.Background()
	// This is just an interface to Go's function, so just a quick simple test
	ktx := runtime.NewKtx(nil, nil)
	mm := new(MathMod)
	mm.SetKtx(ktx)
	val := 3.12
	val2 := 1
	ret := mm.math_IsInf(ctx, runtime.Number(val), runtime.Number(val2))
	exp := math.IsInf(val, val2)
	if ret.Bool(ctx) != exp {
		t.Errorf("expected %v, got %v", exp, ret.Bool(ctx))
	}
}

func TestMathIsNaN(t *testing.T) {
	ctx := context.Background()
	// This is just an interface to Go's function, so just a quick simple test
	ktx := runtime.NewKtx(nil, nil)
	mm := new(MathMod)
	mm.SetKtx(ktx)
	val := 0.12
	ret := mm.math_IsNaN(ctx, runtime.Number(val))
	exp := math.IsNaN(val)
	if ret.Bool(ctx) != exp {
		t.Errorf("expected %v, got %v", exp, ret.Bool(ctx))
	}
}

func TestMathNaN(t *testing.T) {
	ctx := context.Background()
	// This is just an interface to Go's function, so just a quick simple test
	ktx := runtime.NewKtx(nil, nil)
	mm := new(MathMod)
	mm.SetKtx(ktx)
	ret := mm.math_NaN(ctx)
	exp := math.NaN()
	if math.IsNaN(ret.Float(ctx)) != math.IsNaN(exp) {
		t.Errorf("expected NaN, got %f", ret.Float(ctx))
	}
}

func TestMathPow(t *testing.T) {
	ctx := context.Background()
	// This is just an interface to Go's function, so just a quick simple test
	ktx := runtime.NewKtx(nil, nil)
	mm := new(MathMod)
	mm.SetKtx(ktx)
	val := 1.12
	val2 := 3.45
	ret := mm.math_Pow(ctx, runtime.Number(val), runtime.Number(val2))
	exp := math.Pow(val, val2)
	if ret.Float(ctx) != exp {
		t.Errorf("expected %f, got %f", exp, ret.Float(ctx))
	}
}

func TestMathSin(t *testing.T) {
	ctx := context.Background()
	// This is just an interface to Go's function, so just a quick simple test
	ktx := runtime.NewKtx(nil, nil)
	mm := new(MathMod)
	mm.SetKtx(ktx)
	val := 1.12
	ret := mm.math_Sin(ctx, runtime.Number(val))
	exp := math.Sin(val)
	if ret.Float(ctx) != exp {
		t.Errorf("expected %f, got %f", exp, ret.Float(ctx))
	}
}

func TestMathSinh(t *testing.T) {
	ctx := context.Background()
	// This is just an interface to Go's function, so just a quick simple test
	ktx := runtime.NewKtx(nil, nil)
	mm := new(MathMod)
	mm.SetKtx(ktx)
	val := 1.12
	ret := mm.math_Sinh(ctx, runtime.Number(val))
	exp := math.Sinh(val)
	if ret.Float(ctx) != exp {
		t.Errorf("expected %f, got %f", exp, ret.Float(ctx))
	}
}

func TestMathSqrt(t *testing.T) {
	ctx := context.Background()
	// This is just an interface to Go's function, so just a quick simple test
	ktx := runtime.NewKtx(nil, nil)
	mm := new(MathMod)
	mm.SetKtx(ktx)
	val := 1.12
	ret := mm.math_Sqrt(ctx, runtime.Number(val))
	exp := math.Sqrt(val)
	if ret.Float(ctx) != exp {
		t.Errorf("expected %f, got %f", exp, ret.Float(ctx))
	}
}

func TestMathTan(t *testing.T) {
	ctx := context.Background()
	// This is just an interface to Go's function, so just a quick simple test
	ktx := runtime.NewKtx(nil, nil)
	mm := new(MathMod)
	mm.SetKtx(ktx)
	val := 1.12
	ret := mm.math_Tan(ctx, runtime.Number(val))
	exp := math.Tan(val)
	if ret.Float(ctx) != exp {
		t.Errorf("expected %f, got %f", exp, ret.Float(ctx))
	}
}

func TestMathTanh(t *testing.T) {
	ctx := context.Background()
	// This is just an interface to Go's function, so just a quick simple test
	ktx := runtime.NewKtx(nil, nil)
	mm := new(MathMod)
	mm.SetKtx(ktx)
	val := 1.12
	ret := mm.math_Tanh(ctx, runtime.Number(val))
	exp := math.Tanh(val)
	if ret.Float(ctx) != exp {
		t.Errorf("expected %f, got %f", exp, ret.Float(ctx))
	}
}
