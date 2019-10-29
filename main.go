package main

import (
	"fmt"

	"go.starlark.net/starlark"
)

func main() {
	thread := &starlark.Thread{Name: "my thread"}
	chart, err := NewChart(thread, "cf")

	if err != nil {
		panic(err)
	}
	t, err := chart.Template()
	if err != nil {
		panic(err)
	}
	fmt.Println(t)
}
