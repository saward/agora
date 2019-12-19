package stdlib

import (
	"bytes"
	"context"
	"regexp"
	"strings"

	"github.com/saward/agora/runtime"
)

// The strings module, as documented in
// https://github.com/saward/agora/wiki/Standard-library
type StringsMod struct {
	ktx *runtime.Kontext
	ob  runtime.Object
}

func (s *StringsMod) ID() string {
	return "strings"
}

func (s *StringsMod) Run(_ context.Context, _ ...runtime.Val) (v runtime.Val, err error) {
	defer runtime.PanicToError(&err)
	if s.ob == nil {
		// Prepare the object
		s.ob = runtime.NewObject()
		s.ob.Set(runtime.String("ToLower"), runtime.NewNativeFunc(s.ktx, "strings.ToLower", s.strings_ToLower))
		s.ob.Set(runtime.String("ToUpper"), runtime.NewNativeFunc(s.ktx, "strings.ToUpper", s.strings_ToUpper))
		s.ob.Set(runtime.String("HasPrefix"), runtime.NewNativeFunc(s.ktx, "strings.HasPrefix", s.strings_HasPrefix))
		s.ob.Set(runtime.String("HasSuffix"), runtime.NewNativeFunc(s.ktx, "strings.HasSuffix", s.strings_HasSuffix))
		s.ob.Set(runtime.String("Matches"), runtime.NewNativeFunc(s.ktx, "strings.Matches", s.strings_Matches))
		s.ob.Set(runtime.String("ByteAt"), runtime.NewNativeFunc(s.ktx, "strings.ByteAt", s.strings_ByteAt))
		s.ob.Set(runtime.String("Concat"), runtime.NewNativeFunc(s.ktx, "strings.Concat", s.strings_Concat))
		s.ob.Set(runtime.String("Contains"), runtime.NewNativeFunc(s.ktx, "strings.Contains", s.strings_Contains))
		s.ob.Set(runtime.String("Index"), runtime.NewNativeFunc(s.ktx, "strings.Index", s.strings_Index))
		s.ob.Set(runtime.String("LastIndex"), runtime.NewNativeFunc(s.ktx, "strings.LastIndex", s.strings_LastIndex))
		s.ob.Set(runtime.String("Slice"), runtime.NewNativeFunc(s.ktx, "strings.Slice", s.strings_Slice))
		s.ob.Set(runtime.String("Split"), runtime.NewNativeFunc(s.ktx, "strings.Split", s.strings_Split))
		s.ob.Set(runtime.String("Join"), runtime.NewNativeFunc(s.ktx, "strings.Join", s.strings_Join))
		s.ob.Set(runtime.String("Replace"), runtime.NewNativeFunc(s.ktx, "strings.Replace", s.strings_Replace))
		s.ob.Set(runtime.String("Repeat"), runtime.NewNativeFunc(s.ktx, "strings.Repeat", s.strings_Repeat))
		s.ob.Set(runtime.String("Trim"), runtime.NewNativeFunc(s.ktx, "strings.Trim", s.strings_Trim))
	}
	return s.ob, nil
}

func (s *StringsMod) SetKtx(c *runtime.Kontext) {
	s.ktx = c
}

// Converts strings to uppercase, concatenating all strings.
// Args:
// 0..n - The strings to convert to upper case and concatenate
// Returns:
// The uppercase string
func (s *StringsMod) strings_ToUpper(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(1, args)
	buf := bytes.NewBuffer(nil)
	for _, v := range args {
		_, err := buf.WriteString(strings.ToUpper(v.String(ctx)))
		if err != nil {
			panic(err)
		}
	}
	return runtime.String(buf.String())
}

// Converts strings to lowercase, concatenating all strings.
// Args:
// 0..n - The strings to convert to lower case and concatenate
// Returns:
// The lowercase string
func (s *StringsMod) strings_ToLower(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(1, args)
	buf := bytes.NewBuffer(nil)
	for _, v := range args {
		_, err := buf.WriteString(strings.ToLower(v.String(ctx)))
		if err != nil {
			panic(err)
		}
	}
	return runtime.String(buf.String())
}

// Returns true if the string at arg0 starts with any of the following strings.
// Args:
// 0 - The source string
// 1..n - The prefixes to test
// Returns:
// true if the source string starts with any of the specified prefixes
func (s *StringsMod) strings_HasPrefix(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(2, args)
	src := args[0].String(ctx)
	for _, v := range args[1:] {
		if strings.HasPrefix(src, v.String(ctx)) {
			return runtime.Bool(true)
		}
	}
	return runtime.Bool(false)
}

// Returns true if the string at arg0 ends with any of the following strings.
// Args:
// 0 - The source string
// 1..n - The suffixes to test
// Returns:
// true if the source string ends with any of the specified suffixes
func (s *StringsMod) strings_HasSuffix(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(2, args)
	src := args[0].String(ctx)
	for _, v := range args[1:] {
		if strings.HasSuffix(src, v.String(ctx)) {
			return runtime.Bool(true)
		}
	}
	return runtime.Bool(false)
}

func (s *StringsMod) strings_Repeat(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(2, args)
	src := args[0].String(ctx)
	n := int(args[1].Int(ctx))
	return runtime.String(strings.Repeat(src, n))
}

// Args:
// 0 - The string
// 1 - The regexp pattern
// 2 - (optional) a maximum number of matches to return
//
// Returns:
// An object holding all the matches, or nil if no match.
// Each match contains:
// n - The nth match group (when n=0, the full text of the match)
// Each match group contains:
// start - the index of the start of the match
// end - the end of the match
// text - the string of the match
func (s *StringsMod) strings_Matches(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(2, args)
	src := args[0].String(ctx)
	rx := regexp.MustCompile(args[1].String(ctx))
	n := -1 // By default, return all matches
	if len(args) > 2 {
		n = int(args[2].Int(ctx))
	}
	strmtch := rx.FindAllStringSubmatch(src, n)
	if strmtch == nil {
		return runtime.Nil
	}
	ixmtch := rx.FindAllStringSubmatchIndex(src, n)
	ob := runtime.NewObject()
	for i, mtches := range strmtch {
		obch := runtime.NewObject()
		for j, mtch := range mtches {
			leaf := runtime.NewObject()
			leaf.Set(runtime.String("Text"), runtime.String(mtch))
			leaf.Set(runtime.String("Start"), runtime.Number(ixmtch[i][2*j]))
			leaf.Set(runtime.String("End"), runtime.Number(ixmtch[i][2*j+1]))
			obch.Set(runtime.Number(j), leaf)
		}
		ob.Set(runtime.Number(i), obch)
	}
	return ob
}

// Args:
// 0 - The source string
// 1 - The 0-based index number
//
// Returns:
// The character at that position, as a string, or an empty string if
// the index is out of bounds.
func (s *StringsMod) strings_ByteAt(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(2, args)
	src := args[0].String(ctx)
	at := int(args[1].Int(ctx))
	if at < 0 || at >= len(src) {
		return runtime.String("")
	}
	return runtime.String(src[at])
}

// Args:
// 0..n - the strings to concatenate
// Returns:
// The concatenated string
func (s *StringsMod) strings_Concat(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(2, args)
	buf := bytes.NewBuffer(nil)
	for _, v := range args {
		_, err := buf.WriteString(v.String(ctx))
		if err != nil {
			panic(err)
		}
	}
	return runtime.String(buf.String())
}

// Args:
// 0 - the source string
// 1..n - the strings to test if they are found in the source string
// Returns:
// True if any of the strings are found in the source string, false otherwise.
func (s *StringsMod) strings_Contains(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(2, args)
	src := args[0].String(ctx)
	for _, v := range args[1:] {
		if strings.Contains(src, v.String(ctx)) {
			return runtime.Bool(true)
		}
	}
	return runtime.Bool(false)
}

// Args:
// 0 - The source string
// 1 - [Optional] the start index in the source string
// 2 (or 1) .. n - The substrings to search for in the source string.
// Returns:
// The index of the first found substring in source, if any is found, or -1
func (s *StringsMod) strings_Index(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(2, args)
	src := args[0].String(ctx)
	start := 0
	find := 1
	switch v := args[1].(type) {
	case runtime.Number:
		runtime.ExpectAtLeastNArgs(3, args)
		start = int(v.Int(ctx))
		find = 2
	}
	src = src[start:]
	for _, v := range args[find:] {
		if ix := strings.Index(src, v.String(ctx)); ix >= 0 {
			return runtime.Number(ix)
		}
	}
	return runtime.Number(-1)
}

// Args:
// 0 - The source string
// 1 - [Optional] the start index in the source string
// 2 (or 1) .. n - The substrings to search for in the source string.
// Returns:
// The last index of the first found substring in source, if any is found, or -1
func (s *StringsMod) strings_LastIndex(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(2, args)
	src := args[0].String(ctx)
	start := 0
	find := 1
	switch v := args[1].(type) {
	case runtime.Number:
		runtime.ExpectAtLeastNArgs(3, args)
		start = int(v.Int(ctx))
		find = 2
	}
	src = src[start:]
	for _, v := range args[find:] {
		if ix := strings.LastIndex(src, v.String(ctx)); ix >= 0 {
			return runtime.Number(ix)
		}
	}
	return runtime.Number(-1)
}

// Slice a string to get a substring. Basically the same as slicing in Go.
// Args:
// 0 - The source string
// 1 - The start index
// 2 [optional] - The high bound, such that the length of the resulting string is high-start
// Results:
// The sliced string.
func (s *StringsMod) strings_Slice(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(2, args)
	src := args[0].String(ctx)
	start := args[1].Int(ctx)
	end := len(src)
	if len(args) > 2 {
		end = int(args[2].Int(ctx))
	}
	return runtime.String(src[start:end])
}

// Args:
// 0 - the source string
// 1 - the separator
// 2 [optional] - the maximum number of splits, defaults to all
// Returns:
// An array-like object with splits as values and indices as keys.
func (s *StringsMod) strings_Split(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(2, args)
	src := args[0].String(ctx)
	sep := args[1].String(ctx)
	cnt := -1
	if len(args) > 2 {
		cnt = int(args[2].Int(ctx))
	}
	splits := strings.SplitN(src, sep, cnt)
	ob := runtime.NewObject()
	for i, v := range splits {
		ob.Set(runtime.Number(i), runtime.String(v))
	}
	return ob
}

// Args:
// 0 - The source object
// 1 - The separator, empty string by default
// Returns:
// The concatenated string of all the array-like indices of the source object.
func (s *StringsMod) strings_Join(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(1, args)
	ob := args[0].(runtime.Object)
	sep := ""
	if len(args) > 1 {
		sep = args[1].String(ctx)
	}
	l := int(ob.Len(ctx).Int(ctx))
	buf := bytes.NewBuffer(nil)
	for i := 0; i < l; i++ {
		val := ob.Get(runtime.Number(i))
		if _, err := buf.WriteString(val.String(ctx)); err != nil {
			panic(err)
		}
		if i < l-1 {
			if _, err := buf.WriteString(sep); err != nil {
				panic(err)
			}
		}
	}
	return runtime.String(buf.String())
}

// Args:
// 0 - The source string
// 1 - The old substring to replace
// 2 [optional] - the new substring to insert (none by default, delete only)
// 3 [optional] - the number of occurrences to replace. If 2 is a number, it is
// considered the value of 3 and 2 is empty.
// Returns:
// The string with n occurrences of old replaced by new.
func (s *StringsMod) strings_Replace(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(2, args)
	src := args[0].String(ctx)
	old := args[1].String(ctx)
	nw := ""
	cnt := -1
	if len(args) > 2 {
		switch v := args[2].(type) {
		case runtime.Number:
			cnt = int(v.Int(ctx))
		default:
			// args[2] is the new string, args[3], if present, is the count
			nw = v.String(ctx)
			if len(args) > 3 {
				cnt = int(args[3].Int(ctx))
			}
		}
	}
	return runtime.String(strings.Replace(src, old, nw, cnt))
}

// Args:
// 0 - the source string
// 1 [optional] - the cutset (all leading and trailing characters in this string will be
// removed). Defaults to whitespace (space, \n, \t, \v and \r).
// Returns:
// The trimmed string.
func (s *StringsMod) strings_Trim(ctx context.Context, args ...runtime.Val) runtime.Val {
	runtime.ExpectAtLeastNArgs(1, args)
	src := args[0].String(ctx)
	cut := " \n\t\v\r"
	if len(args) > 1 {
		cut = args[1].String(ctx)
	}
	return runtime.String(strings.Trim(src, cut))
}
