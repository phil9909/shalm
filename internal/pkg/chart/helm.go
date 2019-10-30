package chart

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/blang/semver"
	"go.starlark.net/starlark"
	yaml "gopkg.in/yaml.v2"
)

// HelmChart -
type HelmChart struct {
	apiVersion  string         `json:"apiVersion,omitempty"`
	name        string         `json:"name,omitempty"`
	version     semver.Version `json:"version,omitempty"`
	description string         `json:"description,omitempty"`
	keywords    []string       `json:"keywords,omitempty"`
	home        string         `json:"home,omitempty"`
	sources     []string       `json:"sources,omitempty"`
	icon        string         `json:"icon,omitempty"`
}

// LoadHelmChart -
func LoadHelmChart(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return nil, fmt.Errorf("chart: expected paramater name")
	}

	return NewHelmChart(thread, args.Index(0).(starlark.String).GoString())
}

// NewHelmChart -
func NewHelmChart(thread *starlark.Thread, name string) (*Chart, error) {

	directory := repo.Directory(name)
	result := &Chart{directory: directory, name: name}
	var helmChart HelmChart
	err := loadYamlFile(filepath.Join(directory, "Chart.yaml"), &helmChart)
	if err != nil {
		return nil, err
	}
	result.version = helmChart.version
	var values map[string]interface{}
	err = loadYamlFile(filepath.Join(directory, "values.yaml"), &values)
	if err != nil {
		return nil, err
	}
	result.values = make(map[string]starlark.Value)
	for k, v := range values {
		result.values[k] = toStarlark(v)
	}
	return result, nil
}

func loadYamlFile(filename string, value interface{}) error {
	reader, err := os.Open(filename) // For read access.
	if err != nil {
		return fmt.Errorf("Unable to open file %s for parsing: %s", filename, err.Error())
	}
	defer reader.Close()
	decoder := yaml.NewDecoder(reader)
	err = decoder.Decode(value)
	if err != nil {
		return fmt.Errorf("Error during parsing file %s: %s", filename, err.Error())
	}
	return nil
}

func toStarlark(v interface{}) starlark.Value {
	switch v := reflect.ValueOf(v); v.Kind() {
	case reflect.String:
		return starlark.String(v.String())
	case reflect.Bool:
		return starlark.Bool(v.Bool())
	case reflect.Int:
		return starlark.MakeInt64(v.Int())
	case reflect.Float32:
		return starlark.Float(v.Float())
	case reflect.Float64:
		return starlark.Float(v.Float())
	case reflect.Slice:
		a := make([]starlark.Value, 0)
		for i := 0; i < v.Len(); i++ {
			a = append(a, toStarlark(v.Index(i)))
		}
		return starlark.NewList(a)
	case reflect.Map:
		d := starlark.NewDict(16)
		for _, key := range v.MapKeys() {
			strct := v.MapIndex(key)
			d.SetKey(toStarlark(key.Interface()), toStarlark(strct.Interface()))
		}
		return d

	default:
		panic(fmt.Errorf("cannot convert %v to starlark", v))
	}
}
