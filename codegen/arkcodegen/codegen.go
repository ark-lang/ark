package arkcodegen

import (
	"fmt"
	"os"
	"time"

	"github.com/ark-lang/ark-go/parser"
	"github.com/ark-lang/ark-go/util"
)

func (v *Codegen) err(err string, stuff ...interface{}) {
	fmt.Printf(util.TEXT_RED+util.TEXT_BOLD+"Ark codegen error:"+util.TEXT_RESET+" %s\n",
		fmt.Sprintf(err, stuff...))
	os.Exit(2)
}

func (v *Codegen) write(format string, stuff ...interface{}) {
	fmt.Fprintf(v.curOutput, format, stuff...)
}

// outputs a newline if not in minified mode
func (v *Codegen) nl() {
	if !v.Minify {
		v.write("\n")
		for i := 0; i < v.indent; i++ {
			v.write("\t")
		}
	}
}

type Codegen struct {
	Minify bool

	input     []*parser.File
	curOutput *os.File

	indent int // number of tabs indented
}

func (v *Codegen) Generate(input []*parser.File, verbose bool) {
	v.input = input

	for _, infile := range input {
		if verbose {
			fmt.Println(util.TEXT_BOLD+util.TEXT_GREEN+"Started codegenning"+util.TEXT_RESET, infile.Name)
		}
		t := time.Now()

		outname := infile.Name + ".out"

		var err error
		v.curOutput, err = os.OpenFile(outname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			v.err("Couldn't create output file `%s`"+err.Error(), outname)
		}
		defer v.curOutput.Close()

		for _, node := range infile.Nodes {
			v.genNode(node)
		}

		dur := time.Since(t)
		if verbose {
			fmt.Printf(util.TEXT_BOLD+util.TEXT_GREEN+"Finished codegenning"+util.TEXT_RESET+" %s (%.2fms)\n",
				infile.Name, float32(dur.Nanoseconds())/1000000)
		}
	}
}

func (v *Codegen) genNode(n parser.Node) {
	switch n.(type) {
	case parser.Decl:
		v.genDecl(n.(parser.Decl))
	case parser.Expr:
		v.genExpr(n.(parser.Expr))
	case parser.Stat:
		// TODO
	}
}

func (v *Codegen) genDecl(n parser.Decl) {
	switch n.(type) {
	case *parser.FunctionDecl:
		// TODO
	case *parser.StructDecl:
		v.genStructDecl(n.(*parser.StructDecl))
	case *parser.VariableDecl:
		v.genVariableDecl(n.(*parser.VariableDecl), true)
	default:
		panic("")
	}
}

func (v *Codegen) writeAttrs(attrs []*parser.Attr) {

}

func (v *Codegen) genStructDecl(n *parser.StructDecl) {
	v.write("struct ")
	v.writeAttrs(n.Struct.Attrs())
	v.write("%s {", n.Struct.Name)
	v.indent++
	v.nl()

	for i, member := range n.Struct.Variables {
		v.genVariableDecl(member, false)
		if i == len(n.Struct.Variables)-1 {
			v.indent--
		} else {
			v.write(",")
		}
		v.nl()
	}

	v.write("}")
	v.nl()
}

func (v *Codegen) genVariableDecl(n *parser.VariableDecl, semicolon bool) {
	/*for _, attr := range n.Variable.Attrs {
		// TODO
	}*/
	v.writeAttrs(n.Variable.Attrs)
	if n.Variable.Mutable {
		v.write("%s ", parser.KEYWORD_MUT)
	}
	v.write("%s: %s", n.Variable.Name, n.Variable.Type.TypeName())
	if n.Assignment != nil {
		v.write(" = ")
		v.genExpr(n.Assignment)
	}
	if semicolon {
		v.write(";")
		v.nl()
	}
}

func (v *Codegen) genExpr(n parser.Expr) {
	switch n.(type) {
	case *parser.RuneLiteral:
		v.genRuneLiteral(n.(*parser.RuneLiteral))
	case *parser.IntegerLiteral:
		v.genIntegerLiteral(n.(*parser.IntegerLiteral))
	case *parser.FloatingLiteral:
		v.genFloatingLiteral(n.(*parser.FloatingLiteral))
	case *parser.StringLiteral:
		v.genStringLiteral(n.(*parser.StringLiteral))
	case *parser.BinaryExpr:
		v.genBinaryExpr(n.(*parser.BinaryExpr))
	case *parser.UnaryExpr:
		v.genUnaryExpr(n.(*parser.UnaryExpr))
	case *parser.CastExpr:
		v.genCastExpr(n.(*parser.CastExpr))
	case *parser.CallExpr: // TODO
	case *parser.AccessExpr: // TODO
	case *parser.DerefExpr:
		v.genDerefExpr(n.(*parser.DerefExpr))
	default:
		panic("unimplemented expr")
	}
}

func (v *Codegen) genRuneLiteral(n *parser.RuneLiteral) {
	v.write("'%s'", parser.EscapeString(string(n.Value)))
}

func (v *Codegen) genIntegerLiteral(n *parser.IntegerLiteral) {
	v.write("%d", n.Value)
}

func (v *Codegen) genFloatingLiteral(n *parser.FloatingLiteral) {
	v.write("%f", n.Value)
}

func (v *Codegen) genStringLiteral(n *parser.StringLiteral) {
	v.write("\"%s\"", parser.EscapeString(n.Value))
}

func (v *Codegen) genBinaryExpr(n *parser.BinaryExpr) {
	v.genExpr(n.Lhand)
	v.write(" %s ", n.Op.OpString())
	v.genExpr(n.Rhand)
}

func (v *Codegen) genUnaryExpr(n *parser.UnaryExpr) {
	v.write(" %s ", n.Op.OpString())
	v.genExpr(n.Expr)
}

func (v *Codegen) genCastExpr(n *parser.CastExpr) {
	v.write("%s(", n.Type.TypeName())
	v.genExpr(n.Expr)
	v.write(")")
}

func (v *Codegen) genDerefExpr(n *parser.DerefExpr) {
	v.write("^")
	v.genExpr(n.Expr)
}
