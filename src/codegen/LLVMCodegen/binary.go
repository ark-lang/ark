package LLVMCodegen

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"llvm.org/llvm/bindings/go/llvm"

	"github.com/ark-lang/ark/src/parser"
	"github.com/ark-lang/ark/src/util/log"
)

type OutputType int

const (
	OUTPUT_ASSEMBLY OutputType = iota
	OUTPUT_OBJECT
	OUTPUT_LLVM_IR
	OUTPUT_EXECUTABLE
)

func (v *Codegen) createIR(mod *WrappedModule) string {
	filename := v.OutputName + ".ll"

	err := ioutil.WriteFile(filename, []byte(mod.LlvmModule.String()), 0666)
	if err != nil {
		v.err("Couldn't write IR file "+filename+": `%s`", err.Error())
	}

	return filename
}

func (v *Codegen) createObjectOrAssembly(mod *WrappedModule, typ llvm.CodeGenFileType) string {
	filename := v.OutputName + "-" + mod.MangledName(parser.MANGLE_ARK_UNSTABLE)
	if typ == llvm.AssemblyFile {
		filename += ".s"
	} else {
		filename += ".o"
	}

	membuf, err := v.targetMachine.EmitToMemoryBuffer(mod.LlvmModule, typ)
	if err != nil {
		v.err("Couldn't generate file "+filename+": `%s`", err.Error())
	}

	err = ioutil.WriteFile(filename, membuf.Bytes(), 0666)
	if err != nil {
		v.err("Couldn't create file "+filename+": `%s`", err.Error())
	}

	return filename
}

func (v *Codegen) createBinary() {
	if v.OutputType == OUTPUT_LLVM_IR {
		for _, mod := range v.input {
			log.Timed("creating ir", mod.Name.String(), func() {
				v.createIR(mod)
			})
		}
		return
	} else if v.OutputType == OUTPUT_ASSEMBLY {
		for _, mod := range v.input {
			log.Timed("creating assembly", mod.Name.String(), func() {
				v.createObjectOrAssembly(mod, llvm.AssemblyFile)
			})
		}
		return
	}

	linkArgs := append(v.LinkerArgs, "-fno-PIE", "-nodefaultlibs", "-lc", "-lm")

	objFiles := []string{}

	for _, mod := range v.input {
		log.Timed("creating object", mod.Name.String(), func() {
			objName := v.createObjectOrAssembly(mod, llvm.ObjectFile)
			objFiles = append(objFiles, objName)
			linkArgs = append(linkArgs, objName)
			for _, lib := range mod.LinkedLibraries {
				linkArgs = append(linkArgs, fmt.Sprintf("-l%s", lib))
			}
		})
	}

	if v.OutputType == OUTPUT_OBJECT {
		return
	}

	if v.OutputName == "" {
		panic("OutputName is empty")
	}

	linkArgs = append(linkArgs, "-o", v.OutputName)

	if v.Linker == "" {
		v.Linker = "cc"
	}

	log.Timed("linking", "", func() {
		cmd := exec.Command(v.Linker, linkArgs...)
		if out, err := cmd.CombinedOutput(); err != nil {
			v.err("failed to link object files: `%s`\n%s", err.Error(), string(out))
		}
	})

	for _, objFile := range objFiles {
		os.Remove(objFile)
	}
}
