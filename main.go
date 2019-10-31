package main

import (
	"fmt"

	"github.wdf.sap.corp/pylm/internal/pkg/chart"
	"go.starlark.net/starlark"
)

func main() {
	var repo = chart.LocalRepo{BaseDir: "example"}

	thread := &starlark.Thread{Name: "my thread"}
	c, err := chart.NewChart(thread, &repo, "cf")

	if err != nil {
		panic(err)
	}
	t, err := c.Template(&chart.Release{Name: "test", Namespace: "test"})
	if err != nil {
		panic(err)
	}
	fmt.Println(t)
}
