package renderer

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/k14s/ytt/pkg/template"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/k14s/ytt/pkg/yamltemplate"
	"go.starlark.net/starlark"
)

// YttFileRenderer -
func YttFileRenderer(value starlark.Value) func(filename string, writer io.Writer) error {
	return func(filename string, writer io.Writer) error {
		return yttRenderFile(value, filename, writer)
	}
}

func yttRenderFile(value starlark.Value, filename string, writer io.Writer) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return yttRender(value, f, filename, writer)
}

func yttRender(value starlark.Value, reader io.Reader, associatedName string, writer io.Writer) error {
	prefix := bytes.NewBuffer([]byte("#@ load(\"self\", \"self\")\n"))
	content, err := ioutil.ReadAll(io.MultiReader(prefix, reader))
	if err != nil {
		return err
	}
	docSet, err := yamlmeta.NewDocumentSetFromBytes(content, yamlmeta.DocSetOpts{AssociatedName: associatedName})
	if err != nil {
		return err
	}
	compiledTemplate, err := yamltemplate.NewTemplate(associatedName, yamltemplate.TemplateOpts{}).Compile(docSet)
	if err != nil {
		return err
	}

	thread := &starlark.Thread{Name: "test", Load: func(thread *starlark.Thread, module string) (starlark.StringDict, error) {
		if module == "self" {
			return starlark.StringDict{
				"self": value,
			}, nil
		}
		return nil, fmt.Errorf("Unknown module '%s'", module)
	}}

	_, newVal, err := compiledTemplate.Eval(thread, template.NoopCompiledTemplateLoader{})
	if err != nil {
		return err
	}

	typedNewVal, ok := newVal.(interface{ AsBytes() ([]byte, error) })
	if !ok {
		return fmt.Errorf("Invalid return type of CompiledTemplate.Eval")
	}

	body, err := typedNewVal.AsBytes()
	if err != nil {
		return err
	}
	writer.Write(body)
	return nil
}
