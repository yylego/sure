COVERAGE_DIR ?= .coverage.out

# Regenerate every artifact. Each //go:generate directive above a gen_test.go states
# the tag that suits it, so a run gets through even when a broken artifact keeps its
# package from compiling. See sure.DirectiveSkipGenerate
# 重新生成全部产物。各 gen_test.go 顶上的 //go:generate 指令带着各自该用的标签，
# 于是产物坏掉、包编译不过时也照样跑得起来。参见 sure.DirectiveSkipGenerate
generate:
	@go generate ./...

test:
	@if [ -d $(COVERAGE_DIR) ]; then rm -r $(COVERAGE_DIR); fi
	@mkdir $(COVERAGE_DIR)
	make test-with-flags TEST_FLAGS='-v -race -covermode atomic -coverprofile $$(COVERAGE_DIR)/combined.txt -bench=. -benchmem -timeout 20m'

test-with-flags:
	@go test $(TEST_FLAGS) ./...
