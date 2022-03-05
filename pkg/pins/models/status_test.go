package models

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStatusString(t *testing.T) {
	expected := "pinning"
	actual := StatusPinning.String()
	require.Equal(t, expected, actual)
}

func TestToStatus(t *testing.T) {
	expected := StatusPinning
	actual, err := ToStatus("pinning")
	require.Nil(t, err)
	require.Equal(t, expected, actual)
}
