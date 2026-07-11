package sure_pkg_gen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yylego/sure"
)

// TestGeneratedFileNameTrimsExtensionExactly guards the output-file naming: the
// mode suffix must attach right before the trailing ".go", using TrimSuffix.
// A source file whose name contains ".go" in the middle (e.g. "x.gogen.go")
// would break the old strings.Replace(name, ".go", …) path — it hits the first
// ".go" and yields "x_must.gogen.go" instead of "x.gogen_must.go".
//
// TestGeneratedFileNameTrimsExtensionExactly 守护输出文件命名：模式后缀必须精确接在结尾的 ".go" 前
// （用 TrimSuffix）。文件名中间含 ".go"（如 "x.gogen.go"）时，旧的 strings.Replace(name, ".go", …)
// 会命中中间那处、生成 "x_must.gogen.go" 而非 "x.gogen_must.go"。
func TestGeneratedFileNameTrimsExtensionExactly(t *testing.T) {
	sourceRoot := t.TempDir()
	src := `package demo

func Do() error { return nil }
`
	require.NoError(t, os.WriteFile(filepath.Join(sourceRoot, "x.gogen.go"), []byte(src), 0o644))

	cfg := NewSurePackageConfig(sourceRoot, sure.MUST, "example.com/demo")
	GenerateSurePackageFiles(cfg)

	outDir := filepath.Join(sourceRoot, "demo_must")
	_, err := os.Stat(filepath.Join(outDir, "x.gogen_must.go"))
	require.NoError(t, err, "suffix must attach before the trailing .go")

	_, err = os.Stat(filepath.Join(outDir, "x_must.gogen.go"))
	require.True(t, os.IsNotExist(err), "must not replace a middle .go (the old strings.Replace bug)")
}
