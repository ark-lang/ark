package doc

import (
	"github.com/ark-lang/ark/src/util/log"
	"os"
	"time"

	"github.com/ark-lang/ark/src/parser"
	"github.com/ark-lang/ark/src/util"
)

type Docgen struct {
	Input []*parser.Module
	Dir   string

	output    []*File
	curOutput *File
}

func (v *Docgen) Generate() {
	log.Verboseln("docgen", util.TEXT_BOLD+util.TEXT_GREEN+"Started docgenning"+util.TEXT_RESET)
	t := time.Now()

	v.output = make([]*File, 0)

	v.traverse()

	v.generate()

	dur := time.Since(t)
	log.Verbose("docgen", util.TEXT_BOLD+util.TEXT_GREEN+"Finished docgenning"+util.TEXT_RESET+" (%.2fms)\n",
		float32(dur.Nanoseconds())/1000000)
}

func (v *Docgen) traverse() {
	for _, file := range v.Input {
		v.curOutput = &File{
			Name: file.Name,
		}

		for _, n := range file.Nodes {
			switch n.(type) {
			case parser.Decl:
				decl := &Decl{
					Node: n.(parser.Documentable),
				}

				for _, comm := range decl.Node.DocComments() {
					decl.Docs += comm.Contents + "\n"
				}

				decl.process()

				switch n.(type) {
				case *parser.FunctionDecl:
					v.curOutput.FunctionDecls = append(v.curOutput.FunctionDecls, decl)
				case *parser.StructDecl:
					v.curOutput.StructDecls = append(v.curOutput.StructDecls, decl)
				case *parser.TraitDecl:
					v.curOutput.TraitDecls = append(v.curOutput.TraitDecls, decl)
				case *parser.ImplDecl:
					v.curOutput.ImplDecls = append(v.curOutput.ImplDecls, decl)
				case *parser.VariableDecl:
					v.curOutput.VariableDecls = append(v.curOutput.VariableDecls, decl)
				default:
					panic("dammit")
				}
			}
		}

		v.output = append(v.output, v.curOutput)
		v.curOutput = nil
	}
}

func (v *Docgen) generate() {
	if v.Dir[len(v.Dir)-1] != '/' {
		v.Dir += "/"
	}

	err := os.MkdirAll(v.Dir+"files", os.ModeDir|0777)
	if err != nil {
		panic(err)
	}

	v.generateStyle()
	v.generateIndex()

	for _, outputFile := range v.output {
		v.generateFile(outputFile)
	}
}
