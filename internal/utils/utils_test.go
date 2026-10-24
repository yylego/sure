package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNeatBytes(t *testing.T) {
	data, err := NeatBytes(map[string]int{"num": 1})
	require.NoError(t, err)
	require.Equal(t, "{\n\t\"num\": 1\n}", string(data))
}

func TestIsFileExistAndRootExist(t *testing.T) {
	base := t.TempDir()
	onePath := filepath.Join(base, "one.txt")
	require.NoError(t, os.WriteFile(onePath, []byte("x"), 0o644))

	hasFile, err := IsFileExist(onePath)
	require.NoError(t, err)
	require.True(t, hasFile)

	fileOnBase, err := IsFileExist(base) // a root is not a file
	require.NoError(t, err)
	require.False(t, fileOnBase)

	missFile, err := IsFileExist(filepath.Join(base, "none.txt"))
	require.NoError(t, err)
	require.False(t, missFile)

	hasBase, err := IsRootExist(base)
	require.NoError(t, err)
	require.True(t, hasBase)

	baseOnFile, err := IsRootExist(onePath) // a file is not a root
	require.NoError(t, err)
	require.False(t, baseOnFile)
}
