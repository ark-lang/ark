package LLVMCodegen

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/ark-lang/ark/src/parser"
	"github.com/ark-lang/ark/src/util/log"

	"llvm.org/llvm/bindings/go/llvm"
)

type OutputType int

const (
	OUTPUT_ASSEMBLY OutputType = iota
	OUTPUT_OBJECT
	OUTPUT_LLVM_IR
	OUTPUT_LLVM_BC
	OUTPUT_EXECUTABLE
)

func (v *Codegen) createBitcode(file *WrappedModule) string {
	filename := v.OutputName + "-" + file.MangledName(parser.MANGLE_ARK_UNSTABLE) + ".bc"

	fileHandle, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		v.err("Couldn't create bitcode file "+filename+": `%s`", err.Error())
	}
	defer fileHandle.Close()

	if err := llvm.WriteBitcodeToFile(file.LlvmModule, fileHandle); err != nil {
		v.err("failed to write bitcode to file for "+file.Name.String()+": `%s`", err.Error())
	}

	return filename
}

func (v *Codegen) bitcodeToASM(filename string) string {
	asmName := filename

	if strings.HasSuffix(filename, ".bc") {
		asmName = asmName[:len(asmName)-3]
	}

	asmName += ".s"

	cmd := exec.Command("llc", filename, "-o", asmName)
	if out, err := cmd.CombinedOutput(); err != nil {
		v.err("Failed to convert bitcode to assembly: `%s`\n%s", err.Error(), string(out))
	}

	return asmName
}

func (v *Codegen) asmToObject(filename string) string {
	objName := filename

	if strings.HasSuffix(filename, ".s") {
		objName = objName[:len(objName)-2]
	}

	objName += ".o"

	if v.Compiler == "" {
		envcc := os.Getenv("CC")
		if envcc != "" {
			v.Compiler = envcc
		} else {
			v.Compiler = "cc"
		}
	}

	args := append(v.CompilerArgs, "-fno-PIE", "-c", filename, "-o", objName)

	cmd := exec.Command(v.Compiler, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		v.err("Failed to convert assembly to object file: `%s`\n%s", err.Error(), string(out))
	}

	return objName
}

func (v *Codegen) createIR(mod *WrappedModule) string {
	filename := v.OutputName + ".ll"

	err := ioutil.WriteFile(filename, []byte(mod.LlvmModule.String()), 0666)
	if err != nil {
		v.err("Couldn't write IR file "+filename+": `%s`", err.Error())
	}

	return filename
}

func (v *Codegen) createBinary() {
	// god this is a long and ugly function

	if v.OutputType == OUTPUT_LLVM_IR {
		for _, file := range v.input {
			log.Timed("creating ir", file.Name.String(), func() {
				v.createIR(file)
			})
		}
		return
	}

	linkArgs := append(v.LinkerArgs, "-fno-PIE", "-nodefaultlibs", "-lc", "-lm")
	libraries := [][]string{}

	bitcodeFiles := []string{}
	for _, file := range v.input {
		log.Timed("creating bitcode", file.Name.String(), func() {
			libraries = append(libraries, file.LinkedLibraries)
			bitcodeFiles = append(bitcodeFiles, v.createBitcode(file))
		})
	}

	if v.OutputType == OUTPUT_LLVM_BC {
		return
	}

	asmFiles := []string{}

	for _, name := range bitcodeFiles {
		log.Timed("creating asm", name, func() {
			asmName := v.bitcodeToASM(name)
			asmFiles = append(asmFiles, asmName)
		})
	}

	for _, bc := range bitcodeFiles {
		if err := os.Remove(bc); err != nil {
			v.err("Failed to remove "+bc+": `%s`", err.Error())
		}
	}

	if v.OutputType == OUTPUT_ASSEMBLY {
		return
	}

	objFiles := []string{}

	for idx, asmFile := range asmFiles {
		log.Timed("creating object", asmFile, func() {
			objName := v.asmToObject(asmFile)
			objFiles = append(objFiles, objName)
			linkArgs = append(linkArgs, objName)
			for _, lib := range libraries[idx] {
				linkArgs = append(linkArgs, fmt.Sprintf("-l%s", lib))
			}
		})
	}

	for _, asmFile := range asmFiles {
		os.Remove(asmFile)
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
