package LLVMCodegen

import (
	"C"
	"unsafe"

	"github.com/ark-lang/ark-go/parser"
	//"llvm.org/llvm/bindings/go/llvm"
)

const intSize = int(unsafe.Sizeof(C.int(0)))

type Codegen struct {
}

func (v *Codegen) Generate(binaryName string, input []*parser.File) {

}
