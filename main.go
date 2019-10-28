package main

import (
	"fmt"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func chart(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("struct: unexpected positional arguments")
	}
	return starlark.False, nil
}

func main() {
	thread := &starlark.Thread{Name: "my thread"}

	predeclared := starlark.StringDict{
		"struct": starlark.NewBuiltin("struct", starlarkstruct.Make),
		"chart":  starlark.NewBuiltin("chart", chart),
	}

	globals, err := starlark.ExecFile(thread, "example/package.star", nil, predeclared)
	if err != nil {
		panic(err)
	}

	init := globals["init"]

	v, err := starlark.Call(thread, init, starlark.Tuple{}, nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("init(10) = %v\n", v)
}
