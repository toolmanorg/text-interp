package interp

import (
	"encoding"
	"fmt"
)

var debugf = func(msg string, args ...interface{}) (int, error) { return 0, nil }

// Resolver is used by Interpolator to lookup values for interpolated variables.
type Resolver interface {
	// Resolve returns the Value associated with the given variable. If no Value
	// is found, Resolve may return nil and an error. Any error returned here
	// will ultimately be returned by Replace.
	Resolve(variable string) (value Value, err error)
}

// An Interpolator is used to interpolate variable expressions embedded within
// a body of text. By default, variable expression are of the form `${varname}`
// and use a backslash as an escape character. An alternate expression format
// may also be specified by calling `NewWithFormat`.
//
// For both Interpolator constructors, the caller must provide a Resolver for
// resolving variable names to their associated values.
type Interpolator struct {
	vfmt    *VarFormat
	resolvr Resolver
}

// New returns a new Interpolator for the given Resolver using the default
// variable expression format.
func New(g Resolver) *Interpolator {
	return &Interpolator{stdFormat, g}
}

// NewWithFormat returns a new Interpolator for the given Resolver and
// specified VarFormat.
func NewWithFormat(g Resolver, b *VarFormat) *Interpolator {
	return &Interpolator{b, g}
}

// Interpolate repeatedly searchs the given string for the inner-most variable
// expression replacing each one with its associated value acquired from the
// Interpolator's Resolver. If the Resolver returns an error, or if its
// returned Value cannot be stringified, an empty string and error are
// returned.
//
// Interpolate continues this process until it no longer finds a variable
// expression in the provided string, or an error is encountered.
func (i *Interpolator) Interpolate(s string) (string, error) {
	out := s

	for {
		rs := i.vfmt.replString(out)

		var t *token
		if t = rs.next(); t == nil {
			return out, nil
		}

		v, err := i.resolvr.Resolve(t.name)
		if err != nil {
			return "", err
		}

		ns, err := valueString(v)
		if err != nil {
			return "", err
		}

		out = t.replace(ns)
	}
}

// Value is an enpty interface type returned by a Resolvers' Resolve method.
// Its underlying value should be something that is ultimately stringable with
// the following guidelines:
//
//   * string values are used directly
//   * String() is called for fmt.Stringer values
//   * If Value is an encoding.TextMarshaler, its MarshalText()
//     method is called and the results are converted to a string
//   * Otherwise the value is passed through fmt.Sprintf("%v")
//
type Value interface{}

func valueString(v Value) (string, error) {
	switch tv := v.(type) {
	case string:
		return tv, nil

	case fmt.Stringer:
		return tv.String(), nil

	case encoding.TextMarshaler:
		b, err := tv.MarshalText()
		if err != nil {
			return "", err
		}
		return string(b), nil

	default:
		return fmt.Sprintf("%v", v), nil
	}
}
