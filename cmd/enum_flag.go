package cmd

import (
	"fmt"
	"github.com/spf13/pflag"
	"strings"
)

type enum[T ~string] struct {
	Value   *T
	Allowed []T

	stringOptions []string
}

func enumVarP[T ~string](set *pflag.FlagSet, options []T, p *T, value T, name, short, usage string) {
	flag := newEnum(options, p, value)
	typeOptionString := fmt.Sprintf("[options: %s]", strings.Join(flag.stringOptions, ", "))

	set.VarP(
		flag,
		name,
		short,
		usage+" "+typeOptionString,
	)
}

// newEnum give a list of allowed flag parameters, where the second argument is the default
func newEnum[T ~string](allowed []T, d *T, val T) *enum[T] {
	stringOptions := make([]string, len(allowed))
	for i, opt := range allowed {
		stringOptions[i] = string(opt)
	}

	enumFlag := &enum[T]{
		Allowed:       allowed,
		Value:         d,
		stringOptions: stringOptions,
	}
	*enumFlag.Value = val
	return enumFlag
}

func (a *enum[T]) String() string {
	return string(*a.Value)
}

func (a *enum[T]) Set(p string) error {
	isIncluded := func(opts []T, val string) bool {
		for _, opt := range opts {
			if val == string(opt) {
				return true
			}
		}
		return false
	}
	if !isIncluded(a.Allowed, p) {
		return fmt.Errorf("%s is not included in %s", p, strings.Join(a.stringOptions, ","))
	}
	*a.Value = T(p)
	return nil
}

func (a *enum[T]) Type() string {
	return "string"
}
