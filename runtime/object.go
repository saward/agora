package runtime

import (
	"bytes"
	"context"
	"fmt"
)

type (
	// This error is raised if a non-existing method is called.
	NoSuchMethodError string
)

// Error interface implementation.
func (e NoSuchMethodError) Error() string {
	return string(e)
}

// Create a new NoSuchMethodError.
func NewNoSuchMethodError(m string) NoSuchMethodError {
	return NoSuchMethodError(fmt.Sprintf("no such method: %s", m))
}

// The Object interface represents an agora object, which is an associative array.
// It can get and set keys, retrieve the length, the list of keys, and call methods
// and meta-methods.
type Object interface {
	Val
	Get(Val) Val
	Set(Val, Val)
	Len(context.Context) Val
	Keys(context.Context) Val
	callMethod(context.Context, Val, ...Val) Val
	callMetaMethod(context.Context, string, ...Val) (Val, bool)
}

// An object is a map of values, an associative array.
type object struct {
	m map[Val]Val
}

// NewObject returns a new instance of an object.
func NewObject() Object {
	return &object{
		make(map[Val]Val),
	}
}

// Dump pretty-prints the content of the object.
func (o *object) Dump() string {
	buf := bytes.NewBuffer(nil)
	for k, v := range o.m {
		buf.WriteString(fmt.Sprintf(" %s: %s, ", dumpVal(k), dumpVal(v)))
	}
	return fmt.Sprintf("{%s} (Object)", buf)
}

func (o *object) callMetaMethod(ctx context.Context, nm string, args ...Val) (Val, bool) {
	if mm, ok := o.m[String(nm)]; ok {
		if f, ok := mm.(Func); ok {
			return f.Call(ctx, o, args...), true
		}
	}
	return nil, false
}

// Int returns the integer value of the object. Such behaviour can be defined
// if a `__int` method is available on the object.
func (o *object) Int(ctx context.Context) int64 {
	if v, ok := o.callMetaMethod(ctx, "__int"); ok {
		return v.Int(ctx)
	}
	panic(NewTypeError(Type(o), "", "int"))
}

// Float returns the float value of the object. Such behaviour can be defined
// if a `__float` method is available on the object.
func (o *object) Float(ctx context.Context) float64 {
	if v, ok := o.callMetaMethod(ctx, "__float"); ok {
		return v.Float(ctx)
	}
	panic(NewTypeError(Type(o), "", "float"))
}

// String returns the string value of the object. Such behaviour can be overridden
// if a `__string` method is available on the object.
func (o *object) String(ctx context.Context) string {
	if v, ok := o.callMetaMethod(ctx, "__string"); ok {
		return v.String(ctx)
	}
	// Otherwise print the object's contents
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('{')
	keys := o.Keys(ctx).(Object)
	for i, l := int64(0), keys.Len(ctx).Int(ctx); i < l; i++ {
		if i > 0 {
			buf.WriteByte(',')
		}
		ival := Number(i)
		buf.WriteString(keys.Get(ival).String(ctx))
		buf.WriteByte(':')
		buf.WriteString(o.Get(keys.Get(ival)).String(ctx))
	}
	buf.WriteByte('}')
	return buf.String()
}

// Bool returns the boolean value of the object. Such behaviour can be defined
// if a `__bool` method is available on the object. Otherwise it returns true.
func (o *object) Bool(ctx context.Context) bool {
	if v, ok := o.callMetaMethod(ctx, "__bool"); ok {
		return v.Bool(ctx)
	}
	// If __bool is not defined, object returns true (since it is not nil)
	return true
}

// Native returns the Go native value of the object. Such behaviour can be defined
// if a `__native` method is available on the object.
func (o *object) Native(ctx context.Context) interface{} {
	if v, ok := o.callMetaMethod(ctx, "__native"); ok {
		return v.Native(ctx)
	}
	// Defaults to returning the internal map
	return o.m
}

// Get the length of the object. The behaviour can be overridden
// if a `__len` method is available on the object.
func (o *object) Len(ctx context.Context) Val {
	if v, ok := o.callMetaMethod(ctx, "__len"); ok {
		return v
	}
	return Number(len(o.m))
}

// Get the keys of the object in an array-like object value,
// indexed from 0 the the number of keys - 1. It is the responsibility
// of the object's implementation to return coherent values for Len()
// and Keys(). The list of keys is unordered.
func (o *object) Keys(ctx context.Context) Val {
	if v, ok := o.callMetaMethod(ctx, "__keys"); ok {
		return v
	}
	ob := NewObject()
	i := 0
	for k, _ := range o.m {
		ob.Set(Number(i), k)
		i++
	}
	return ob
}

// Get returns the value of the field identified by key. It returns Nil
// if the field does not exist.
func (o *object) Get(key Val) Val {
	if v, ok := o.m[key]; ok {
		return v
	}
	return Nil
}

// Set assigns the value v to the field identified by key. If the value
// is Nil, set instead removes the key from the object. If the key is nil,
// an error is raised.
func (o *object) Set(key Val, v Val) {
	if v == Nil {
		delete(o.m, key)
	} else if key == Nil {
		panic(NewTypeError(Type(key), "", "key"))
	} else {
		o.m[key] = v
	}
}

// callMethod calls the method identified by nm with the provided arguments.
// It panics if the field does not hold a function. If the field does not
// exist and a method named `__noSuchMethod` is defined, it is called instead.
func (o *object) callMethod(ctx context.Context, nm Val, args ...Val) Val {
	v, ok := o.m[nm]
	if ok {
		if f, ok := v.(Func); ok {
			return f.Call(ctx, o, args...)
		} else {
			panic(NewNoSuchMethodError(nm.String(ctx)))
		}
	} else if v, ok := o.callMetaMethod(ctx, "__noSuchMethod", append([]Val{nm}, args...)...); ok {
		// Method not found - call __noSuchMethod if it exists, otherwise panic
		return v
	} else {
		panic(NewNoSuchMethodError(nm.String(ctx)))
	}
}
