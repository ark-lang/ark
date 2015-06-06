package doc

import (
	"fmt"
	"os"
	"time"

	"github.com/ark-lang/ark/parser"
	"github.com/ark-lang/ark/util"
)

type Docgen struct {
	Input []*parser.File
	Dir   string

	output    []*File
	curOutput *File
}

func (v *Docgen) Generate(verbose bool) {
	if verbose {
		fmt.Println(util.TEXT_BOLD + util.TEXT_GREEN + "Started docgenning" + util.TEXT_RESET)
	}
	t := time.Now()

	v.output = make([]*File, 0)

	v.traverse(verbose)

	v.generate()

	dur := time.Since(t)
	if verbose {
		fmt.Printf(util.TEXT_BOLD+util.TEXT_GREEN+"Finished docgenning"+util.TEXT_RESET+" (%.2fms)\n",
			float32(dur.Nanoseconds())/1000000)
	}
}

func (v *Docgen) traverse(verbose bool) {
	for _, file := range v.Input {
		v.curOutput = &File{Name: file.Name}

		for _, n := range file.Nodes {
			switch n.(type) {
			case parser.Decl:
				decl := &Decl{
					AstDecl: n.(parser.Decl),
				}

				for _, comm := range decl.AstDecl.DocComments() {
					decl.Docs += comm.Contents + "\n"
				}

				v.curOutput.Decls = append(v.curOutput.Decls, decl)
			}
		}

		v.output = append(v.output, v.curOutput)
		v.curOutput = nil
	}
}

func (v *Docgen) generate() {
	outputDir := "./" + v.Dir + "/"

	err := os.MkdirAll(outputDir, os.ModeDir|0777)
	if err != nil {
		panic(err)
	}

	indexFile, err := os.OpenFile(outputDir+"index.html", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
	defer indexFile.Close()
	v.generateIndex(indexFile)

	/*for _, outputFile := range v.output {
		for _, decl := range outputFile.Decls {

		}
	}*/
}
