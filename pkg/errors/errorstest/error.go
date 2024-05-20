package errorstest

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"diskey/pkg/errors"
)

func NoError[T errors.Cause](t *testing.T, err errors.Error[T]) {
	t.Helper()

	assert.Truef(t, err.IsOk(), "expected no error, but got %s", err)
}

func ErrorIs[T errors.Cause](t *testing.T, err errors.Error[T], target T) {
	t.Helper()

	assert.Equalf(t, err.Cause(), target, "expected %T(%v), but got %s", err)
}
