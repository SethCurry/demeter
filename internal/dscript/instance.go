package dscript

import (
	"fmt"

	"github.com/glycerine/zygomys/v9/zygo"
)

func ExampleFunction(env *zygo.Zlisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	return zygo.SexpNull, fmt.Errorf("function not implemented: %s", name)
}

func NewInstance() *Instance {
	env := zygo.NewZlisp()
	env.AddGlobal("foo",
		&zygo.SexpStr{
		S: "bar"},
	)

	env.AddFunction("example", ExampleFunction)

	return &Instance{
		env: env,
	}
}

type Instance struct {
	env *zygo.Zlisp
}

type Error struct {
	StackTrace string
	err error
}

func (e *Error) Error() string {
	return e.err.Error()
}

func (i *Instance) Run() (string, error) {
	expr, err := i.env.Run()
	if err != nil {
		trace := i.env.GetStackTrace(err)

		i.env.Clear()

		return "", &Error{
			StackTrace: trace,
			err: err,
		}
	}

	ps := zygo.NewPrintState()

	asString := expr.SexpString(ps)

	return asString, nil
}

type Module struct {
	Name string
}
