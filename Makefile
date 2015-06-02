all:
	go install github.com/ark-lang/ark

fmt:
	go fmt github.com/ark-lang/ark/{codegen{/LLVMCodegen,},common,lexer,parser,util}

gen:
	go generate ./{codegen{/LLVMCodegen,},common,lexer,parser,util}
