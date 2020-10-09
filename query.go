package route

import (
	"fmt"
	"reflect"
	"strconv"
)

type getter interface {
	Get(string) string
}

type setter interface {
	Set(string, string)
}

func valForKind(val string, t reflect.Type) (v reflect.Value, err error) {

	// TODO(gdey): Add support for embedded structs. There is no reason not
	// to traverse down structs and fill them out as well.

	switch t.Kind() {

	case reflect.Ptr:
		elem, err := valForKind(val, t.Elem())
		if err != nil {
			return v, err
		}
		v = reflect.New(t.Elem())
		v.Elem().Set(elem)
		return v, nil

	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		return reflect.ValueOf(b), err

	case reflect.Int:
		i, err := strconv.ParseInt(val, 0, 0)
		return reflect.ValueOf(int(i)), err

	case reflect.Int8:
		i, err := strconv.ParseInt(val, 0, 8)
		return reflect.ValueOf(int8(i)), err

	case reflect.Int16:
		i, err := strconv.ParseInt(val, 0, 16)
		return reflect.ValueOf(int16(i)), err

	case reflect.Int32:
		i, err := strconv.ParseInt(val, 0, 32)
		return reflect.ValueOf(int32(i)), err

	case reflect.Int64:
		i, err := strconv.ParseInt(val, 0, 64)
		return reflect.ValueOf(int64(i)), err

	case reflect.Uint:
		i, err := strconv.ParseUint(val, 0, 0)
		return reflect.ValueOf(uint(i)), err

	case reflect.Uint8:
		i, err := strconv.ParseUint(val, 0, 8)
		return reflect.ValueOf(uint8(i)), err

	case reflect.Uint16:
		i, err := strconv.ParseUint(val, 0, 16)
		return reflect.ValueOf(uint16(i)), err

	case reflect.Uint32:
		i, err := strconv.ParseUint(val, 0, 32)
		return reflect.ValueOf(uint32(i)), err

	case reflect.Uint64:
		i, err := strconv.ParseUint(val, 0, 64)
		return reflect.ValueOf(uint64(i)), err

	case reflect.Float64:
		f, err := strconv.ParseFloat(val, 64)
		return reflect.ValueOf(float64(f)), err

	case reflect.Float32:
		f, err := strconv.ParseFloat(val, 32)
		return reflect.ValueOf(float32(f)), err

	case reflect.String:
		return reflect.ValueOf(val), nil

	default:
		return v, ErrQueryParseUnsupportedType{t}
	}
}

func setValue(val string, v reflect.Value) error {

	nv, err := valForKind(val, v.Type())
	if err != nil {
		return fmt.Errorf("failed to set value: %w", err)
	}
	v.Set(nv)
	return nil
}

// CreateQuery will iterate through the given struct or pointer to a struct, looking for
// fields tagged with `query`. For each tagged field that is not the zero value s.Set() is
// called with the string version of that field's value.
func CreateQuery(params interface{}, s setter) error {
	rv := reflect.ValueOf(params)

	if !rv.IsValid() || rv.IsZero() {
		return nil
	}
	// dereference the pointer if it's a pointer
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	t := rv.Type()
	if rv.Kind() != reflect.Struct {
		return ErrQueryParseUnsupportedType{Type: t}
	}
	for i := 0; i < t.NumField(); i++ {

		ftype := t.Field(i).Type
		fval := rv.Field(i)
	KIND_TEST:
		switch ftype.Kind() {
		case reflect.Complex128, reflect.Complex64: // may want to support this later
			fallthrough
		case reflect.Struct: // may want to support this later
			fallthrough
		case reflect.Slice, reflect.Chan, reflect.Func, reflect.Array, reflect.Interface, reflect.Invalid:
			continue // skip these types
		case reflect.Ptr:
			ftype = ftype.Elem()
			fval = fval.Elem()
			goto KIND_TEST
		}

		tag := t.Field(i).Tag.Get(tagName)
		if tag == "" || tag == "-" {
			continue
		}

		// don't set the value if it's not valid or the value
		// is the zero value.
		if !fval.IsValid() || fval.IsZero() {
			continue
		}

		s.Set(tag, fmt.Sprintf("%v", fval))
	}
	return nil
}

// ParseQuery will fill out the params based on the query tag.
func ParseQuery(r getter, params interface{}) error {

	rv := reflect.ValueOf(params)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return ErrNonPointer
	}

	// get the value pointed to by rv
	rvv := rv.Elem()
	if rvv.Kind() != reflect.Struct {
		return ErrQueryParseUnsupportedType{rvv.Type()}
	}

	// go through each of the fields and fill them in.
	t := rvv.Type()
	for i := 0; i < t.NumField(); i++ {

		tag := t.Field(i).Tag.Get(tagName)
		// if the type kind is a struct
		if t.Field(i).Type.Kind() == reflect.Struct {
			v := rvv.Field(i).Addr().Interface()
			if err := ParseQuery(r, v); err != nil {
				return err
			}
			continue
		}
		if tag == "" || tag == "-" {
			continue
		}

		fieldValue := rvv.Field(i)
		if !fieldValue.CanSet() {
			continue
		}

		val := r.Get(tag)
		if val == "" {
			continue
		}

		if err := setValue(val, fieldValue); err != nil {
			return err
		}
	}
	return nil
}
