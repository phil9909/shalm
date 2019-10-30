package main

import (
	"fmt"

	"github.wdf.sap.corp/pylm/internal/pkg/chart"
	"go.starlark.net/starlark"
)

func main() {
	thread := &starlark.Thread{Name: "my thread"}
	chart, err := chart.NewChart(thread, "cf")

	if err != nil {
		panic(err)
	}
	t, err := chart.Template()
	if err != nil {
		panic(err)
	}
	fmt.Println(t)
}
