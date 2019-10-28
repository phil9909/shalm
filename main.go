package main

import (
	"fmt"

	"go.starlark.net/starlark"
)

func main() {
	thread := &starlark.Thread{Name: "my thread"}

	predeclared := starlark.StringDict{
		"chart": starlark.NewBuiltin("chart", LoadChart),
	}

	globals, err := starlark.ExecFile(thread, "example/chart.star", nil, predeclared)
	if err != nil {
		panic(err)
	}

	init := globals["init"]

	v, err := starlark.Call(thread, init, starlark.Tuple{&Chart{}}, nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("init() = %v\n", v)
}
