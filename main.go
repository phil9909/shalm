package main

import (
	"fmt"

	"github.wdf.sap.corp/shalm/internal/pkg/chart"
	"go.starlark.net/starlark"
)

func main() {
	var repo = chart.LocalRepo{BaseDir: "example"}

	thread := &starlark.Thread{Name: "my thread"}
	c, err := chart.NewChart(thread, &repo, "cf")

	if err != nil {
		panic(err)
	}
	t, err := starlark.Call(thread, c.TemplateFunction(), starlark.Tuple{&chart.Release{Name: "cf", Namespace: "test", Service: "cf"}}, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(t.(starlark.String).GoString())
}
