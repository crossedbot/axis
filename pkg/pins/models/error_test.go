package models

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestError(t *testing.T) {
	reason := ErrorCodeStrings[ErrRequiredParamCode]
	details := "hello world"
	expected := fmt.Sprintf("%s: %s", reason, details)
	err := Error{reason, details}
	actual := err.Error()
	require.Equal(t, expected, actual)
}

func TestNewFailure(t *testing.T) {
	reasonCode := ErrRequiredParamCode
	reason := ErrorCodeStrings[reasonCode]
	details := "something about parameters"
	expected := Failure{
		Error: Error{
			Reason:  reason,
			Details: details,
		},
	}
	actual := NewFailure(reasonCode, details)
	require.Equal(t, expected, actual)
}
