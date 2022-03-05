package models

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMatchString(t *testing.T) {
	expected := "iexact"
	actual := TextMatchIExact.String()
	require.Equal(t, expected, actual)
}

func TestToTextMatch(t *testing.T) {
	expected := TextMatchIExact
	actual, err := ToTextMatch("iexact")
	require.Nil(t, err)
	require.Equal(t, expected, actual)
}
