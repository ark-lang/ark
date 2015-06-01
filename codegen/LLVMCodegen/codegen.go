package LLVMCodegen

import (
	"C"
	"unsafe"

	//"llvm.org/llvm/bindings/go/llvm"
)

const intSize = int(unsafe.Sizeof(C.int(0)))

type Codegen struct {
}

func (v *Codegen) Generate() {

}
