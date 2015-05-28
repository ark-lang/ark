all:
	go install github.com/ark-lang/ark-go

fmt:
	go fmt github.com/ark-lang/ark-go/{codegen{/LLVMCodegen,},common,lexer,parser,util}

gen:
	go generate ./{codegen{/LLVMCodegen,},common,lexer,parser,util}
