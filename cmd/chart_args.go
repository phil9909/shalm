package cmd

import (
	"strings"

	"github.com/spf13/pflag"
	"go.starlark.net/starlark"
)

type chartArgs struct {
	args  []string
	proxy bool
}

func (v *chartArgs) AddFlags(flagsSet *pflag.FlagSet) {
	flagsSet.StringArrayVar(&v.args, "set", nil, "Set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	flagsSet.BoolVar(&v.proxy, "proxy", false, "Install helm chart using a combination of CR and operator")
}

func (v *chartArgs) KwArgs() []starlark.Tuple {
	var result []starlark.Tuple
	for _, arg := range v.args {
		for _, a := range strings.Split(arg, ",") {
			val := strings.Split(a, "=")
			if len(val) == 2 {
				result = append(result, starlark.Tuple{starlark.String(val[0]), starlark.String(val[1])})
			}
		}
	}
	return result
}
