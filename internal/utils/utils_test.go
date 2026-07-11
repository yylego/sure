package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestC0IsUppercase guards the exported-name check that codegen relies on to
// decide which functions/types to wrap. A wrong verdict here corrupts output.
func TestC0IsUppercase(t *testing.T) {
	require.True(t, C0IsUppercase("Apple"))
	require.True(t, C0IsUppercase("Z"))
	require.False(t, C0IsUppercase("apple"))
	require.False(t, C0IsUppercase(""))
	require.False(t, C0IsUppercase("123"))
	require.False(t, C0IsUppercase("_priv"))
}

func TestSetDoubleQuotes(t *testing.T) {
	require.Equal(t, `"abc"`, SetDoubleQuotes("abc"))
	require.Equal(t, `""`, SetDoubleQuotes(""))
}

func TestIn(t *testing.T) {
	require.True(t, In(2, []int{1, 2, 3}))
	require.False(t, In(9, []int{1, 2, 3}))
	require.True(t, In("b", []string{"a", "b"}))
	require.False(t, In("x", []string{}))
}

func TestNeat2json(t *testing.T) {
	require.Contains(t, Neat2json(map[string]int{"num": 1}), "num")
}

func TestIsFileExistAndRootExist(t *testing.T) {
	base := t.TempDir()
	onePath := filepath.Join(base, "one.txt")
	require.NoError(t, os.WriteFile(onePath, []byte("x"), 0o644))

	hasFile, err := IsFileExist(onePath)
	require.NoError(t, err)
	require.True(t, hasFile)

	fileOnBase, err := IsFileExist(base) // a directory is not a file
	require.NoError(t, err)
	require.False(t, fileOnBase)

	missFile, err := IsFileExist(filepath.Join(base, "none.txt"))
	require.NoError(t, err)
	require.False(t, missFile)

	hasBase, err := IsRootExist(base)
	require.NoError(t, err)
	require.True(t, hasBase)

	baseOnFile, err := IsRootExist(onePath) // a file is not a directory
	require.NoError(t, err)
	require.False(t, baseOnFile)
}

func TestMustLs(t *testing.T) {
	base := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(base, "a.go"), []byte("x"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(base, "b.go"), []byte("x"), 0o644))
	require.ElementsMatch(t, []string{"a.go", "b.go"}, MustLs(base))
}
