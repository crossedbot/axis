package models

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInfoString(t *testing.T) {
	expected := "a:one,b:two,c:three"
	info := Info{
		"a": "one",
		"b": "two",
		"c": "three",
	}
	actual := info.String()
	require.Equal(t, expected, actual)
}

func TestInfoFromString(t *testing.T) {
	expected := Info{
		"a": "one",
		"b": "two",
		"c": "three",
	}
	actual := InfoFromString("a:one,b:two,c:three")
	require.Equal(t, expected, actual)
}
