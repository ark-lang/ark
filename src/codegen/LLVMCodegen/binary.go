package LLVMCodegen

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/ark-lang/ark/src/parser"
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

func (v *Codegen) createBitcode(file *parser.Module) string {
	filename := v.OutputName + "-" + file.Name + ".bc"

	fileHandle, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		v.err("Couldn't create bitcode file "+filename+": `%s`", err.Error())
	}
	defer fileHandle.Close()

	if err := llvm.WriteBitcodeToFile(file.Module, fileHandle); err != nil {
		v.err("failed to write bitcode to file for "+file.Name+": `%s`", err.Error())
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

func (v *Codegen) createIR(mod *parser.Module) string {
	filename := v.OutputName + ".ll"

	err := ioutil.WriteFile(filename, []byte(mod.Module.String()), 0666)
	if err != nil {
		v.err("Couldn't write IR file "+filename+": `%s`", err.Error())
	}

	return filename
}

func (v *Codegen) createBinary() {
	// god this is a long and ugly function

	if v.OutputType == OUTPUT_LLVM_IR {
		for _, file := range v.input {
			v.createIR(file)
		}
		return
	}

	linkArgs := append(v.LinkerArgs, "-fno-PIE", "-nodefaultlibs", "-lc", "-lm")

	bitcodeFiles := []string{}

	for _, file := range v.input {
		bitcodeFiles = append(bitcodeFiles, v.createBitcode(file))
	}

	if v.OutputType == OUTPUT_LLVM_BC {
		return
	}

	asmFiles := []string{}

	for _, name := range bitcodeFiles {
		asmName := v.bitcodeToASM(name)
		asmFiles = append(asmFiles, asmName)
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

	for _, asmFile := range asmFiles {
		objName := v.asmToObject(asmFile)

		objFiles = append(objFiles, objName)
		linkArgs = append(linkArgs, objName)
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

	cmd := exec.Command(v.Linker, linkArgs...)
	if out, err := cmd.CombinedOutput(); err != nil {
		v.err("failed to link object files: `%s`\n%s", err.Error(), string(out))
	}

	for _, objFile := range objFiles {
		os.Remove(objFile)
	}
}
