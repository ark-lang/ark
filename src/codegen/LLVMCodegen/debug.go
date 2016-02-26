package LLVMCodegen

import (
	"github.com/ark-lang/ark/src/ast"
	"github.com/ark-lang/go-llvm/llvm"
	"path/filepath"
)

type DebugCodegen struct {
	codegen *Codegen
	builder *llvm.DIBuilder
	unit    llvm.Metadata
	file    llvm.Metadata
}

func NewDebugCodegen(codegen *Codegen, mod llvm.Module) *DebugCodegen {
	res := &DebugCodegen{
		codegen: codegen,
		builder: llvm.NewDIBuilder(mod),
	}
	return res
}

func (v *DebugCodegen) finalize() {
	v.builder.Finalize()
}

func (v *DebugCodegen) createCompilationUnit(submod *ast.Submodule) {
	dir, file := filepath.Split(submod.File.Path)

	cu := llvm.DICompileUnit{
		Language:       llvm.DW_LANG_C,
		File:           file,
		Dir:            dir,
		Producer:       "Ark Compiler",
		Optimized:      v.codegen.OptLevel != 0,
		Flags:          "",
		RuntimeVersion: 0,
	}
	v.unit = v.builder.CreateCompileUnit(cu)
	v.file = v.builder.CreateFile(file, dir)

	v.builder.Finalize()
}

func (v *DebugCodegen) createFunction(decl *ast.FunctionDecl) llvm.Metadata {
	fun := decl.Function
	f := llvm.DIFunction{
		Name:        fun.Name,
		LinkageName: "", // TODO
		File:        v.file,
		Line:        decl.Pos().Line,
		Type:        v.createSubroutineType(fun.Type),
	}
	/*
	   Name         string
	   LinkageName  string
	   File         Metadata
	   Line         int
	   Type         Metadata
	   LocalToUnit  bool
	   IsDefinition bool
	   ScopeLine    int
	   Flags        int
	   Optimized    bool
	*/
	return v.builder.CreateFunction(v.unit, f)
}

func (v *DebugCodegen) createDebugType(typ ast.Type) llvm.Metadata {
	// TODO: llvmType := v.codegen.typeToLLVMType(typ, gcon)
	switch typ := typ.(type) {
	case ast.PrimitiveType:
		bt := llvm.DIBasicType{
			Name: typ.TypeName(),
			// TODO: SizeInBits: v.codegen.
		}
		return v.builder.CreateBasicType(bt)
	}

	// TODO:
	return llvm.Metadata{}
}

func (v *DebugCodegen) createSubroutineType(typ ast.FunctionType) llvm.Metadata {
	st := llvm.DISubroutineType{
		File: v.file,
	}
	return v.builder.CreateSubroutineType(st)
}
