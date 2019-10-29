package main

import (
	"fmt"

	"go.starlark.net/starlark"
)

func main() {
	thread := &starlark.Thread{Name: "my thread"}
	chart, err := LoadChart(thread, nil, starlark.Tuple{starlark.String("chart")}, nil)

	if err != nil {
		panic(err)
	}
	fmt.Println(chart.String())
}
