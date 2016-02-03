PACKAGES = src/{codegen{/LLVMCodegen,},lexer,doc,parser,semantic,util,cmd/ark}

.PHONY: all fmt gen wc

all:
	@go install -gcflags '-N -l' github.com/ark-lang/ark/src/...

fmt:
	go fmt github.com/ark-lang/ark/...

gen:
	go generate ./...

wc:
	wc ./${PACKAGES}/*.go
