PACKAGES = src/{codegen{/LLVMCodegen,},lexer,parser{/checks,},util,cmd/ark}

all:
	@go install github.com/ark-lang/ark/src/...

fmt:
	go fmt github.com/ark-lang/ark/...

gen:
	go generate ./...

wc:
	wc ./${PACKAGES}/*.go
