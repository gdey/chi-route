package route

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrNonPointer = errors.New("non-pointer passed")
)

// ErrQueryParseUnsupportedType is returned when a query tag is applied
// to an unsupported type
type ErrQueryParseUnsupportedType struct {
	Type reflect.Type
}

func (err ErrQueryParseUnsupportedType) Error() string {
	return fmt.Sprintf("unsupported type: %v (%v)", err.Type.Name(), err.Type.Kind())
}
