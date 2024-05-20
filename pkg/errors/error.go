package errors

import (
	"errors" //nolint:depguard // reason: Imported to shadow this package.
	"fmt"
)

type Standard uint

const (
	StandardCauseInternalFailure = iota + 1
)

type Cause interface {
	~uint
	fmt.Stringer
}

type Error[T Cause] struct {
	cause     T
	rootCause fmt.Stringer
	message   string
}

func Ok[T Cause]() Error[T] {
	err := Error[T]{}
	err.rootCause = err.cause
	return err
}

func New[T Cause](cause T, format string, formatArgs ...any) Error[T] {
	// Do not invoke fmt.Sprintf() if not necessary to avoid performance impact.
	if len(formatArgs) != 0 {
		format = fmt.Sprintf(format, formatArgs...)
	}

	newErr := Error[T]{
		cause:     cause,
		rootCause: cause,
		message:   format,
	}

	return newErr
}

func NewWithErr[T Cause](cause T, err error) Error[T] {
	newErr := Error[T]{
		cause:     cause,
		rootCause: cause,
		message:   err.Error(),
	}

	return newErr
}

func FromError[From Cause, To Cause](cause To, err Error[From]) Error[To] {
	newErr := Error[To]{
		cause:     cause,
		rootCause: err.cause,
		message:   err.message,
	}

	return newErr
}

func (self Error[T]) Cause() T { //nolint:ireturn // reason: The return value is not an interface.
	return self.cause
}

func (self Error[T]) IsOk() bool {
	return self.cause == 0
}

func (self Error[T]) IsErr() bool {
	return self.cause != 0
}

func (self Error[T]) Error() string {
	return self.String()
}

func (self Error[T]) String() string {
	var zero T
	if self.rootCause == zero {
		return self.rootCause.String() + "(Ok)"
	}
	return self.rootCause.String() + "(" + self.message + ")"
}

//nolint:gochecknoglobals // reason: This shadows golang "errors" for feature parity.
var (
	As = errors.As
	Is = errors.Is
)
