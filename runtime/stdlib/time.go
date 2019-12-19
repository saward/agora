package stdlib

import (
	"context"
	"time"

	"github.com/saward/agora/runtime"
)

// The time module, as documented in
// https://github.com/saward/agora/wiki/Standard-library
type TimeMod struct {
	ktx *runtime.Kontext
	ob  runtime.Object
}

func (t *TimeMod) ID() string {
	return "time"
}

func (t *TimeMod) Run(_ context.Context, _ ...runtime.Val) (v runtime.Val, err error) {
	defer runtime.PanicToError(&err)
	if t.ob == nil {
		// Prepare the object
		t.ob = runtime.NewObject()
		t.ob.Set(runtime.String("Date"), runtime.NewNativeFunc(t.ktx, "time.Date", t.time_Date))
		t.ob.Set(runtime.String("Now"), runtime.NewNativeFunc(t.ktx, "time.Now", t.time_Now))
		t.ob.Set(runtime.String("Sleep"), runtime.NewNativeFunc(t.ktx, "time.Sleep", t.time_Sleep))
	}
	return t.ob, nil
}

func (t *TimeMod) SetKtx(c *runtime.Kontext) {
	t.ktx = c
}

func (t *TimeMod) time_Sleep(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(1, args)
	time.Sleep(time.Duration(args[0].Int(ctx)) * time.Millisecond)
	return runtime.Nil
}

type _time struct {
	runtime.Object
	t time.Time
}

func (t *TimeMod) newTime(tm time.Time) runtime.Val {
	ob := &_time{
		runtime.NewObject(),
		tm,
	}
	ob.Set(runtime.String("__int"), runtime.NewNativeFunc(t.ktx, "time._time.__int", func(_ context.Context, args ...runtime.Val) runtime.Val {
		return runtime.Number(ob.t.Unix())
	}))
	ob.Set(runtime.String("__string"), runtime.NewNativeFunc(t.ktx, "time._time.__string", func(_ context.Context, args ...runtime.Val) runtime.Val {
		return runtime.String(ob.t.Format(time.RFC3339))
	}))
	ob.Set(runtime.String("Year"), runtime.Number(tm.Year()))
	ob.Set(runtime.String("Month"), runtime.Number(tm.Month()))
	ob.Set(runtime.String("Day"), runtime.Number(tm.Day()))
	ob.Set(runtime.String("Hour"), runtime.Number(tm.Hour()))
	ob.Set(runtime.String("Minute"), runtime.Number(tm.Minute()))
	ob.Set(runtime.String("Second"), runtime.Number(tm.Second()))
	ob.Set(runtime.String("Nanosecond"), runtime.Number(tm.Nanosecond()))
	return ob
}

func (t *TimeMod) time_Now(ctx context.Context, args ...runtime.Val) runtime.Val {
	return t.newTime(time.Now())
}

func (t *TimeMod) time_Date(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(1, args)
	yr := int(args[0].Int(ctx))
	mth := 1
	if len(args) > 1 {
		mth = int(args[1].Int(ctx))
	}
	dy := 1
	if len(args) > 2 {
		dy = int(args[2].Int(ctx))
	}
	hr := 0
	if len(args) > 3 {
		hr = int(args[3].Int(ctx))
	}
	min := 0
	if len(args) > 4 {
		min = int(args[4].Int(ctx))
	}
	sec := 0
	if len(args) > 5 {
		sec = int(args[5].Int(ctx))
	}
	nsec := 0
	if len(args) > 6 {
		nsec = int(args[6].Int(ctx))
	}
	return t.newTime(time.Date(yr, time.Month(mth), dy, hr, min, sec, nsec, time.UTC))
}
