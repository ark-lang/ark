PACKAGES = src/{codegen{/LLVMCodegen,},lexer,parser,util}

all:
	@go install github.com/ark-lang/ark

fmt:
	go fmt github.com/ark-lang/ark/${PACKAGES}

gen:
	go generate ./${PACKAGES}

wc:
	wc ./${PACKAGES}/*.go
