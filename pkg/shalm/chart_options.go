package shalm

import (
	"strings"

	"github.com/spf13/pflag"
	"go.starlark.net/starlark"
)

// AddFlags -
func (v *ChartOptions) AddFlags(flagsSet *pflag.FlagSet) {
	flagsSet.StringArrayVar(&v.cmdArgs, "set", nil, "Set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	flagsSet.BoolVarP(&v.proxy, "proxy", "p", false, "Install helm chart using a combination of CR and operator")
	flagsSet.StringVarP(&v.namespace, "namespace", "n", "default", "Namespace for installation")
	flagsSet.StringVarP(&v.suffix, "suffix", "s", "", "Suffix which is used to build the chart name")
}

// Options -
func (v *ChartOptions) Options() ChartOption {
	if len(v.kwargs) == 0 {
		v.kwargs = v.kwArgs()
	}
	return func(o *ChartOptions) {
		*o = *v
	}
}

func (v *ChartOptions) kwArgs() []starlark.Tuple {
	var result []starlark.Tuple
	for _, arg := range v.cmdArgs {
		for _, a := range strings.Split(arg, ",") {
			val := strings.Split(a, "=")
			if len(val) == 2 {
				result = append(result, starlark.Tuple{starlark.String(val[0]), starlark.String(val[1])})
			}
		}
	}
	return result
}

func chartOptions(opts []ChartOption) *ChartOptions {
	co := ChartOptions{}
	for _, option := range opts {
		option(&co)
	}
	if co.namespace == "" {
		co.namespace = "default"
	}
	return &co
}
