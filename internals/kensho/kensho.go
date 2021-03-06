// Package kensho provides a series of tests case which can be used to validate that a giving
// generated shogun package meet it's design and expected operation.
package kensho

import (
	"bytes"
	gctx "context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"context"

	"github.com/influx6/faux/tests"
	"github.com/influx6/shogun/internals"
)

// TestWriterFunction validates the behaviour of a function that expects a writer argument.
func TestWriterFunction(fun internals.ShogunFunc) {
	var err error

	defer func() {
		if rec := recover(); rec != nil {
			switch drec := rec.(type) {
			case error:
				err = drec
			default:
				err = fmt.Errorf("Recover Error: %+q", rec)
			}
		}
	}()

	var outgoing bytes.Buffer

	realFunc := fun.Function.(func(io.WriteCloser))
	realGCtxFunc := fun.Function.(func(context.Context, io.WriteCloser))

	realFuncWithReturn := fun.Function.(func(io.WriteCloser) error)
	realGCtxFuncWithReturn := fun.Function.(func(context.Context, io.WriteCloser) error)

	switch fun.Context {
	case internals.NoContext:
		if fun.Return == internals.NoReturn {
			realFunc(wopCloser{Writer: &outgoing})
		}

		if fun.Return == internals.ErrorReturn {
			err = realFuncWithReturn(wopCloser{Writer: &outgoing})
		}
	case internals.UseGoogleContext:
		err = execWithContext(func(ctx context.Context) error {
			if fun.Return == internals.NoReturn {
				realGCtxFunc(ctx, wopCloser{Writer: &outgoing})
			}

			if fun.Return == internals.ErrorReturn {
				err = realGCtxFuncWithReturn(ctx, wopCloser{Writer: &outgoing})
			}

			return nil
		}, 0)
	}

	if err != nil {
		tests.Failed("Function %q with alias %q failed StringOnlyFunction criterias: %+q", fun.Name, fun.NS, err)
		return
	}

	tests.Passed("Function %q with alias %q passes StringOnlyFunction criterias", fun.Name, fun.NS)
}

// TestReaderFunction validates the behaviour of a function that expects a reader argument.
func TestReaderFunction(fun internals.ShogunFunc) {
	var err error

	defer func() {
		if rec := recover(); rec != nil {
			switch drec := rec.(type) {
			case error:
				err = drec
			default:
				err = fmt.Errorf("Recover Error: %+q", rec)
			}
		}
	}()

	var incoming bytes.Buffer
	incoming.WriteString(`{"name":"Rock"}`)

	realFunc := fun.Function.(func(io.Reader))
	realGCtxFunc := fun.Function.(func(context.Context, io.Reader))

	realFuncWithReturn := fun.Function.(func(io.Reader) error)
	realGCtxFuncWithReturn := fun.Function.(func(context.Context, io.Reader) error)

	switch fun.Context {
	case internals.NoContext:
		if fun.Return == internals.NoReturn {
			realFunc(&incoming)
		}

		if fun.Return == internals.ErrorReturn {
			err = realFuncWithReturn(&incoming)
		}
	case internals.UseGoogleContext:
		err = execWithContext(func(ctx context.Context) error {
			if fun.Return == internals.NoReturn {
				realGCtxFunc(ctx, &incoming)
			}

			if fun.Return == internals.ErrorReturn {
				err = realGCtxFuncWithReturn(ctx, &incoming)
			}

			return nil
		}, 0)
	}

	if err != nil {
		tests.Failed("Function %q with alias %q failed StringOnlyFunction criterias: %+q", fun.Name, fun.NS, err)
		return
	}

	tests.Passed("Function %q with alias %q passes StringOnlyFunction criterias", fun.Name, fun.NS)
	return
}

// TestReaderWithWriterFunction validates the behaviour of a function that expects a reader and WriteCloser argument.
func TestReaderWithWriterFunction(fun internals.ShogunFunc) {
	var err error

	defer func() {
		if rec := recover(); rec != nil {
			switch drec := rec.(type) {
			case error:
				err = drec
			default:
				err = fmt.Errorf("Recover Error: %+q", rec)
			}
		}
	}()

	var incoming, outgoing bytes.Buffer
	incoming.WriteString(`{"name":"Rock"}`)

	realFunc := fun.Function.(func(io.Reader, io.WriteCloser))
	realGCtxFunc := fun.Function.(func(context.Context, io.Reader, io.WriteCloser))

	realFuncWithReturn := fun.Function.(func(io.Reader, io.WriteCloser) error)
	realGCtxFuncWithReturn := fun.Function.(func(context.Context, io.Reader, io.WriteCloser) error)

	switch fun.Context {
	case internals.NoContext:
		if fun.Return == internals.NoReturn {
			realFunc(&incoming, wopCloser{Writer: &outgoing})
		}

		if fun.Return == internals.ErrorReturn {
			err = realFuncWithReturn(&incoming, wopCloser{Writer: &outgoing})
		}
	case internals.UseGoogleContext:
		err = execWithContext(func(ctx context.Context) error {
			if fun.Return == internals.NoReturn {
				realGCtxFunc(ctx, &incoming, wopCloser{Writer: &outgoing})
			}

			if fun.Return == internals.ErrorReturn {
				err = realGCtxFuncWithReturn(ctx, &incoming, wopCloser{Writer: &outgoing})
			}

			return nil
		}, 0)
	}

	if err != nil {
		tests.Failed("Function %q with alias %q failed StringOnlyFunction criterias: %+q", fun.Name, fun.NS, err)
		return
	}

	if outgoing.Len() == 0 {
		tests.Failed("Function %q with alias %q should have responded with output", fun.Name, fun.NS)
	}

	tests.Passed("Function %q with alias %q passes StringOnlyFunction criterias", fun.Name, fun.NS)
	return
}

// TestMapFunction validates the behaviour of a function that expects a map argument.
func TestMapFunction(fun internals.ShogunFunc) {
	var err error

	defer func() {
		if rec := recover(); rec != nil {
			switch drec := rec.(type) {
			case error:
				err = drec
			default:
				err = fmt.Errorf("Recover Error: %+q", rec)
			}
		}
	}()

	var incoming bytes.Buffer
	incoming.WriteString(`{"name":"Rock"}`)

	realFunc := fun.Function.(func(map[string]interface{}))
	realGCtxFunc := fun.Function.(func(context.Context, map[string]interface{}))

	realFuncWithReturn := fun.Function.(func(map[string]interface{}) error)
	realGCtxFuncWithReturn := fun.Function.(func(context.Context, map[string]interface{}) error)

	data := make(map[string]interface{})
	if jserr := json.NewDecoder(&incoming).Decode(&data); jserr != nil {
		tests.Failed("Function %q with alias %q failed StringOnlyFunction criterias: %+q", fun.Name, fun.NS, jserr)
		return
	}

	switch fun.Context {
	case internals.NoContext:
		if fun.Return == internals.NoReturn {
			realFunc(data)
		}

		if fun.Return == internals.ErrorReturn {
			err = realFuncWithReturn(data)
		}
	case internals.UseGoogleContext:
		err = execWithContext(func(ctx context.Context) error {
			if fun.Return == internals.NoReturn {
				realGCtxFunc(ctx, data)
			}

			if fun.Return == internals.ErrorReturn {
				err = realGCtxFuncWithReturn(ctx, data)
			}

			return nil
		}, 0)
	}

	if err != nil {
		tests.Failed("Function %q with alias %q failed StringOnlyFunction criterias: %+q", fun.Name, fun.NS, err)
		return
	}

	tests.Passed("Function %q with alias %q passes StringOnlyFunction criterias", fun.Name, fun.NS)
	return
}

// TestMapWithWriterFunction validates the behaviour of a function that expects a string argument.
func TestMapWithWriterFunction(fun internals.ShogunFunc) {
	var err error

	defer func() {
		if rec := recover(); rec != nil {
			switch drec := rec.(type) {
			case error:
				err = drec
			default:
				err = fmt.Errorf("Recover Error: %+q", rec)
			}
		}
	}()

	var incoming, outgoing bytes.Buffer
	incoming.WriteString(`{"name":"Rock"}`)

	realFunc := fun.Function.(func(map[string]interface{}, io.WriteCloser))
	realGCtxFunc := fun.Function.(func(context.Context, map[string]interface{}, io.WriteCloser))

	realFuncWithReturn := fun.Function.(func(map[string]interface{}, io.WriteCloser) error)
	realGCtxFuncWithReturn := fun.Function.(func(context.Context, map[string]interface{}, io.WriteCloser) error)

	data := make(map[string]interface{})
	if jserr := json.NewDecoder(&incoming).Decode(&data); jserr != nil {
		tests.Failed("Function %q with alias %q failed StringOnlyFunction criterias: %+q", fun.Name, fun.NS, jserr)
		return
	}

	switch fun.Context {
	case internals.NoContext:
		if fun.Return == internals.NoReturn {
			realFunc(data, wopCloser{Writer: &outgoing})
		}

		if fun.Return == internals.ErrorReturn {
			err = realFuncWithReturn(data, wopCloser{Writer: &outgoing})
		}
	case internals.UseGoogleContext:
		err = execWithContext(func(ctx context.Context) error {
			if fun.Return == internals.NoReturn {
				realGCtxFunc(ctx, data, wopCloser{Writer: &outgoing})
			}

			if fun.Return == internals.ErrorReturn {
				err = realGCtxFuncWithReturn(ctx, data, wopCloser{Writer: &outgoing})
			}

			return nil
		}, 0)
	}

	if err != nil {
		tests.Failed("Function %q with alias %q failed StringOnlyFunction criterias: %+q", fun.Name, fun.NS, err)
		return
	}

	if outgoing.Len() == 0 {
		tests.Failed("Function %q with alias %q should have responded with output", fun.Name, fun.NS)
	}

	tests.Passed("Function %q with alias %q passes StringOnlyFunction criterias", fun.Name, fun.NS)
	return
}

// TestNoArgumentFunction validates the behaviour of a function that expects no argument.
func TestNoArgumentFunction(fun internals.ShogunFunc) {
	var err error

	defer func() {
		if rec := recover(); rec != nil {
			switch drec := rec.(type) {
			case error:
				err = drec
			default:
				err = fmt.Errorf("Recover Error: %+q", rec)
			}
		}
	}()

	realFunc := fun.Function.(func())
	realGCtxFunc := fun.Function.(func(context.Context))
	realFCtxFunc := fun.Function.(func(context.Context))
	realCnFCtxFunc := fun.Function.(func(context.Context))

	realFuncWithReturn := fun.Function.(func() error)
	realGCtxFuncWithReturn := fun.Function.(func(context.Context) error)
	realFCtxFuncWithReturn := fun.Function.(func(context.Context) error)
	realCnFCtxFuncWithReturn := fun.Function.(func(context.Context) error)

	switch fun.Context {
	case internals.NoContext:
		if fun.Return == internals.NoReturn {
			realFunc()
		}

		if fun.Return == internals.ErrorReturn {
			err = realFuncWithReturn()
		}
	case internals.UseGoogleContext:
		err = execWithContext(func(ctx context.Context) error {
			if fun.Return == internals.NoReturn {
				realGCtxFunc(ctx)
			}

			if fun.Return == internals.ErrorReturn {
				err = realGCtxFuncWithReturn(ctx)
			}

			return nil
		}, 0)
	case internals.UseFauxContext:
		err = execWithContext(func(ctx context.Context) error {
			if fun.Return == internals.NoReturn {
				if realCnFCtxFunc != nil {
					realCnFCtxFunc(ctx)
				}
				if realFCtxFunc != nil {
					realFCtxFunc(ctx)
				}
			}
			if fun.Return == internals.ErrorReturn {
				if realFCtxFuncWithReturn != nil {
					err = realFCtxFuncWithReturn(ctx)
				}
				if realCnFCtxFuncWithReturn != nil {
					err = realCnFCtxFuncWithReturn(ctx)
				}
			}

			return nil
		}, 0)
	}

	if err != nil {
		tests.Failed("Function %q with alias %q failed StringOnlyFunction criterias: %+q", fun.Name, fun.NS, err)
		return
	}

	tests.Passed("Function %q with alias %q passes StringOnlyFunction criterias", fun.Name, fun.NS)
	return
}

// TestStringFunction validates the behaviour of a function that expects a string argument.
func TestStringFunction(fun internals.ShogunFunc) {
	var err error

	defer func() {
		if rec := recover(); rec != nil {
			switch drec := rec.(type) {
			case error:
				err = drec
			default:
				err = fmt.Errorf("Recover Error: %+q", rec)
			}
		}
	}()

	var incoming bytes.Buffer
	incoming.WriteString("Rock")

	realFunc := fun.Function.(func(string))
	realGCtxFunc := fun.Function.(func(context.Context, string))

	realFuncWithReturn := fun.Function.(func(string) error)
	realGCtxFuncWithReturn := fun.Function.(func(context.Context, string) error)

	switch fun.Context {
	case internals.NoContext:
		if fun.Return == internals.NoReturn {
			realFunc(incoming.String())
		}

		if fun.Return == internals.ErrorReturn {
			err = realFuncWithReturn(incoming.String())
		}
	case internals.UseGoogleContext:
		err = execWithContext(func(ctx context.Context) error {
			if fun.Return == internals.NoReturn {
				realGCtxFunc(ctx, incoming.String())
			}

			if fun.Return == internals.ErrorReturn {
				err = realGCtxFuncWithReturn(ctx, incoming.String())
			}

			return nil
		}, 0)
	}

	if err != nil {
		tests.Failed("Function %q with alias %q failed StringOnlyFunction criterias: %+q", fun.Name, fun.NS, err)
		return
	}

	tests.Passed("Function %q with alias %q passes StringOnlyFunction criterias", fun.Name, fun.NS)
	return
}

// TestStringWithWriterFunction validates the behaviour of a function that expects a string argument.
func TestStringWithWriterFunction(fun internals.ShogunFunc) {
	var err error

	defer func() {
		if rec := recover(); rec != nil {
			switch drec := rec.(type) {
			case error:
				err = drec
			default:
				err = fmt.Errorf("Recover Error: %+q", rec)
			}
		}
	}()

	var incoming, outgoing bytes.Buffer
	incoming.WriteString("Rock")

	realFunc := fun.Function.(func(string, io.WriteCloser))
	realGCtxFunc := fun.Function.(func(context.Context, string, io.WriteCloser))

	realFuncWithReturn := fun.Function.(func(string, io.WriteCloser) error)
	realGCtxFuncWithReturn := fun.Function.(func(context.Context, string, io.WriteCloser) error)

	switch fun.Context {
	case internals.NoContext:
		if fun.Return == internals.NoReturn {
			realFunc(incoming.String(), wopCloser{Writer: &outgoing})
		}

		if fun.Return == internals.ErrorReturn {
			err = realFuncWithReturn(incoming.String(), wopCloser{Writer: &outgoing})
		}
	case internals.UseGoogleContext:
		err = execWithContext(func(ctx context.Context) error {
			if fun.Return == internals.NoReturn {
				realGCtxFunc(ctx, incoming.String(), wopCloser{Writer: &outgoing})
			}

			if fun.Return == internals.ErrorReturn {
				err = realGCtxFuncWithReturn(ctx, incoming.String(), wopCloser{Writer: &outgoing})
			}

			return nil
		}, 0)
	}

	if err != nil {
		tests.Failed("Function %q with alias %q failed StringOnlyFunction criterias: %+q", fun.Name, fun.NS, err)
		return
	}

	if outgoing.Len() == 0 {
		tests.Failed("Function %q with alias %q should have responded with output", fun.Name, fun.NS)
	}

	tests.Passed("Function %q with alias %q passes StringOnlyFunction criterias", fun.Name, fun.NS)
	return
}

func execWithContext(fun interface{}, ctxTimeout time.Duration) error {
	switch dfunc := fun.(type) {
	case func(context.Context) error:
		var ctx context.Context
		var canceller func()

		if ctxTimeout == 0 {
			ctx = gctx.Background()
		} else {
			ctx, canceller = gctx.WithTimeout(gctx.Background(), ctxTimeout)
		}

		if canceller != nil {
			defer canceller()
		}

		return dfunc(ctx)
	}

	return errors.New("Unknown context type")
}

type wopCloser struct {
	io.Writer
}

// Close does nothing.
func (wopCloser) Close() error {
	return nil
}
