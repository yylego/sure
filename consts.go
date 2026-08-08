package sure

// Build constraint that keeps a broken artifact from locking its own package
//
// A generated wrapping class lands in the package that holds the wrapped type, and
// the test that regenerates it lives there too. That shape can lock: once the
// generated file stops compiling, its package stops compiling, and the test inside
// cannot run to replace it, so the sole escape is hand-editing a file headed
// DO NOT EDIT.
//
// A build constraint breaks the lock. Generated files state
//
//	//go:build !sure_generate
//
// so a standard build compiles them, while a run with -tags=sure_generate leaves
// them out. The codegen then faces a package that compiles regardless of what the
// preceding run produced, and overwrites the artifact.
//
// Put the matching directive above the test that regenerates, keeping the command
// beside the code it produces:
//
//	//go:generate go test -run ^TestGen$ -tags=sure_generate
//
// Note 1: a tagged test run compiles the whole test package too. When a test file
// in that package uses the generated API, state the same //go:build !sure_generate
// line on that test file too, since those symbols vanish with the tag and a test
// build that breaks takes the codegen down with it.
//
// Note 2: the tag suits targeted regeneration, not whole-module builds. When a
// package holds nothing besides generated files, the tag leaves it without a single
// Go file, and `go build -tags=sure_generate ./...` reports that build constraints
// exclude each Go file there. A cls_stub_gen target package often has that shape, so
// expect it; `go test -tags=sure_generate` aimed at one package keeps on going
// regardless, which is what regeneration needs. Hence sure_pkg_gen, which emits
// standalone packages that cannot lock, leaves the constraint blank as its default,
// and offers WithGoBuildDirective where an output does land somewhere that locks.
//
// 让破损产物无法卡住自身所在包的构建约束
//
// 生成的包装类会落在被包装类型所在的那个包里，重新生成它的测试也住在那儿。
// 这个格局会卡住：生成文件一旦编译不过，它所在的包就编译不过，
// 包里那个测试也就跑不起来、没法把它换掉，
// 唯一的出路是手工去编辑一个写着"请勿编辑"的文件。
//
// 构建约束能解开这个锁。生成文件写上
//
//	//go:build !sure_generate
//
// 于是普通编译会编译它，而带 -tags=sure_generate 运行时会把它排除在外。
// 这样生成器面对的包，无论上一轮生成出什么都照样能编译，然后把产物覆盖掉。
//
// 把对应的指令写在负责重新生成的测试上方，让命令和它产出的代码待在一起：
//
//	//go:generate go test -run ^TestGen$ -tags=sure_generate
//
// 注意一：带标签跑测试时整个测试包仍然要编译。若同包里有测试文件用到了生成出来的 API，
// 也要给那些测试文件写上同一行 //go:build !sure_generate——
// 带上标签后那些符号就不存在了，测试编译不过，会把生成器一起拖垮。
//
// 注意二：这个标签适合定点重新生成，不适合全模块构建。若某个包里除生成文件之外别无他物，
// 带上标签后那个包就一个 Go 文件都不剩，
// `go build -tags=sure_generate ./...` 会在那里报"构建约束排除了该处的每个 Go 文件"。
// cls_stub_gen 的目标包常常就是这个形状，得有心理准备；
// 而按包跑 `go test -tags=sure_generate` 照样跑得通，重新生成要的正是这个。
// 因此 sure_pkg_gen（它生成的是不会卡住的独立新包）默认留空不写这道约束，
// 若产物确实落在会卡住的地方，可用 WithGoBuildDirective 主动写上。
const (
	// BuildTagGenerate is the tag to set when regenerating
	// BuildTagGenerate 是重新生成时要带上的标签
	BuildTagGenerate = "sure_generate"

	// DirectiveSkipGenerate is the go:build directive placed in generated files
	// DirectiveSkipGenerate 是写进生成文件的 go:build 指令
	DirectiveSkipGenerate = "!" + BuildTagGenerate
)
