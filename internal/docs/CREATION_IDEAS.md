# Background of Creation

[中文说明](CREATION_IDEAS.zh.md)

## Design Intent

This package enables code to provide panic/warning options when errors happen.

## Core Functions

The core of this package provides three functions:

```go
// Must - Hard, panics and crashes on error, halting the process
func Must(err error)

// Soft - Soft, prints a warning log on error, continues the process
func Soft(err error)

// Omit - Quiet, does nothing on error, continues the process
func Omit(err error)
```

## Simple Usage

You can use it like this in the code logic:

```go
data, err := json.Marshal(example)
sure.Must(err) // Does nothing if there is no error, crashes with panic if there is an error
```

You can also use:

```go
sure.Soft(err) // Does nothing if there is no error, prints a warning log if there is an error
```

Of course, this brute-force approach of throwing exceptions might not be appropriate in production scenarios and needs cautious consideration.

## Code Generation

Use code generation to auto give certain functions in a class the capabilities to `must`, also certain functions in a package the capabilities to `must`.

## More Background:

During development, I encountered a confusing situation - I saw some defining the `Must` function like this: `a := Must(defaultValue)`, where if a calculation issue happens, it returns a default value, meaning even without a result, it must return something.

Example function:

```go
func (j *Json) MustInt(args ...int) int {
	var def int

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustInt() received excess arguments %d", len(args))
	}

	i, err := j.Int()
	if err == nil {
		return i
	}

	return def
}
```

I spent considerable time understanding the meaning of `must` here. Those authors even made the default value not required, showing the understanding of `must` was different from mine.

In contrast, when needing to warn on issues, those developers used the `Require` function. It seems people not proficient in English might struggle with what `must` means, which is unfortunate.

In this project, `must` means you must accomplish something to get a result, otherwise accept the consequence (i.e., `panic`), while the opposite understanding is that no challenges should prevent giving a result, returning the default value when the result can't be obtained. The opposite focus is on guaranteeing a result.

Since different understandings exist about `must`, let me explain: this project's `must` means you must accomplish something, and when no issue happens, it proceeds; otherwise, it would `panic`. The entire project is based on this understanding.

Example code:

```go
func (T *SimpleMust) Strings(key string) (res []string) {
	res, err1 := T.T.Strings(key)
	sure.Must(err1)
	return res
}
```

Usage:

```go
tags := sim.Must().Strings("tags")
```

When it can't be obtained, it `panics` and crashes, acting as an assertion when fetching the value.

The entire project is based on this understanding.

## Related Packages

I spent much time choosing between package names like `mustsoft` (soft and hard), `mustgo` (hard), `flexible` (adaptable), and `mustdone` (must achieve), and decided it could work fine with `github.com/yylego/done`, so I named it `sure`, which is quite short.

The `github.com/yylego/done` package can also solve the problem, but it requires an extra `nice` invocation, and each time you wrap it outside, the code becomes more complex to read when wrapping exceeds two layers.

Example usage:

```go
defer func() { // This is an operation when closing db *gorm.DB, assume it is in a test case, which creates a scratch DB, and closing it once the case ends looks like this
    done.Done(done.VCE(db.DB()).Nice().Close())
}()
```

The code becomes much more confusing.

## Ultimate Effect

In practice, usage scenarios are rare, and in 90% of cases, regenerating code to enhance methods is not as good as using `github.com/yylego/done`. In the remaining 10% of cases, using `if err != nil { panic(err) }` works fine.

If you want to use it, you can, as it can save three lines of code in important moments.

---
