package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAreEqual(t *testing.T) {
	i1 := 2
	i2 := 3
	require.False(t, AreEqual(i1, i2))
	i1 = i2
	require.True(t, AreEqual(i1, i2))

	f1 := 1.2
	f2 := 3.4
	require.False(t, AreEqual(f1, f2))
	f1 = f2
	require.True(t, AreEqual(f1, f2))

	c1 := 'a'
	c2 := 'b'
	require.False(t, AreEqual(c1, c2))
	c1 = c2
	require.True(t, AreEqual(c1, c2))

	s1 := "abc"
	s2 := "def"
	require.False(t, AreEqual(s1, s2))
	s1 = s2
	require.True(t, AreEqual(s1, s2))

	b1 := []byte("abc")
	b2 := []byte("def")
	require.False(t, AreEqual(b1, b2))
	b1 = b2
	require.True(t, AreEqual(b1, b2))
}

func TestContainsString(t *testing.T) {
	list := "hello world"
	found, ok := Contains(list, 123)
	require.False(t, ok)
	require.False(t, found)

	list = "hello world"
	elem := "today"
	found, ok = Contains(list, elem)
	require.True(t, ok)
	require.False(t, found)

	list = "hello world"
	elem = "world"
	found, ok = Contains(list, elem)
	require.True(t, ok)
	require.True(t, found)
}

func TestContainsSlice(t *testing.T) {}
