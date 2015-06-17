package arkcodegen

import (
	"fmt"
	"os"
	"time"

	"github.com/ark-lang/ark/parser"
	"github.com/ark-lang/ark/util"
)

func (v *Codegen) err(err string, stuff ...interface{}) {
	fmt.Printf(util.TEXT_RED+util.TEXT_BOLD+"Ark codegen error:"+util.TEXT_RESET+" %s\n",
		fmt.Sprintf(err, stuff...))
	os.Exit(util.EXIT_FAILURE_CODEGEN)
}

func (v *Codegen) write(format string, stuff ...interface{}) {
	if v.lastWasNl {
		for i := 0; i < v.indent; i++ {
			v.curOutput.Write([]byte("\t"))
		}
		v.lastWasNl = false
	}
	fmt.Fprintf(v.curOutput, format, stuff...)
}

// outputs a newline if not in minified mode
func (v *Codegen) nl() {
	if !v.Minify {
		v.curOutput.Write([]byte("\n"))
		v.lastWasNl = true
	}
}

type Codegen struct {
	Minify bool

	input     []*parser.Module
	curOutput *os.File

	indent    int // number of tabs indented
	lastWasNl bool
}

func (v *Codegen) Generate(input []*parser.Module, verbose bool) {
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
		v.genStat(n.(parser.Stat))
	}
}

func (v *Codegen) genStat(n parser.Stat) {
	switch n.(type) {
	case *parser.ReturnStat:
		v.genReturnStat(n.(*parser.ReturnStat))
	case *parser.CallStat:
		v.genCallStat(n.(*parser.CallStat))
	case *parser.AssignStat:
		v.genAssignStat(n.(*parser.AssignStat))
	case *parser.IfStat:
		v.genIfStat(n.(*parser.IfStat))
	case *parser.LoopStat:
		v.genLoopStat(n.(*parser.LoopStat))
	default:
		panic("unimplimented stat")
	}
}

func (v *Codegen) genReturnStat(n *parser.ReturnStat) {
	v.write("%s ", parser.KEYWORD_RETURN)
	v.genExpr(n.Value)
	v.write(";")
	v.nl()
}

func (v *Codegen) genCallStat(n *parser.CallStat) {
	v.genExpr(n.Call)
	v.write(";")
	v.nl()
}

func (v *Codegen) genAssignStat(n *parser.AssignStat) {
	if n.Deref != nil {
		v.genExpr(n.Deref)
	} else {
		v.genExpr(n.Access)
	}
	v.write(" = ")
	v.genExpr(n.Assignment)
	v.write(";")
	v.nl()
}

func (v *Codegen) genIfStat(n *parser.IfStat) {
	v.nl()
	for i, expr := range n.Exprs {
		v.write("%s ", parser.KEYWORD_IF)
		v.genExpr(expr)
		v.write(" ")
		v.genBlock(n.Bodies[i])

		if i == len(n.Exprs)-1 { // last stat
			break
		}

		v.write(" else ")
	}

	if n.Else != nil {
		v.write(" else ")
		v.genBlock(n.Else)
	}

	v.nl()
	v.nl()
}

func (v *Codegen) genLoopStat(n *parser.LoopStat) {
	v.nl()
	v.write("%s ", parser.KEYWORD_FOR)
	switch n.LoopType {
	case parser.LOOP_TYPE_INFINITE:
	case parser.LOOP_TYPE_CONDITIONAL:
		v.genExpr(n.Condition)
		v.write(" ")
	default:
		panic("invalid loop type")
	}

	v.genBlock(n.Body)
	v.nl()
	v.nl()
}

func (v *Codegen) genDecl(n parser.Decl) {
	switch n.(type) {
	case *parser.FunctionDecl:
		v.nl()
		v.genFunctionDecl(n.(*parser.FunctionDecl))
	case *parser.StructDecl:
		v.nl()
		v.genStructDecl(n.(*parser.StructDecl))
	case *parser.TraitDecl:
		// nothing to gen
	case *parser.ImplDecl:
		// TODO
	case *parser.VariableDecl:
		v.genVariableDecl(n.(*parser.VariableDecl), true)
	default:
		panic("unimplimented decl")
	}
}

func (v *Codegen) writeAttrs(attrs []*parser.Attr) {
	for _, attr := range attrs {
		v.write("[%s", attr.Key)
		if attr.Value != "" {
			v.write("=\"%s\"", attr.Value)
		}
		v.write("] ")
	}
}

func (v *Codegen) genBlock(n *parser.Block) {
	v.write("{")
	v.indent++
	v.nl()
	for _, stat := range n.Nodes {
		v.genNode(stat)
	}
	v.indent--
	v.write("}")
}

func (v *Codegen) genFunctionDecl(n *parser.FunctionDecl) {
	v.writeAttrs(n.Function.Attrs)
	v.write("%s ", parser.KEYWORD_FUNC)
	v.write("%s(", n.Function.Name)
	for i, par := range n.Function.Parameters {
		v.genVariableDecl(par, false)
		if i < len(n.Function.Parameters)-1 {
			v.write(", ")
		}
	}
	v.write(")")

	if n.Function.ReturnType != nil {
		v.write(": %s", n.Function.ReturnType.TypeName())
	}
	v.write(" ")
	v.genBlock(n.Function.Body)
	v.nl()
}

func (v *Codegen) genStructDecl(n *parser.StructDecl) {
	v.writeAttrs(n.Struct.Attrs())
	v.write("%s ", parser.KEYWORD_STRUCT)
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
	case *parser.CallExpr:
		v.genCallExpr(n.(*parser.CallExpr))
	case *parser.AccessExpr:
		v.genAccessExpr(n.(*parser.AccessExpr))
	case *parser.DerefExpr:
		v.genDerefExpr(n.(*parser.DerefExpr))
	case *parser.BracketExpr:
		v.genBracketExpr(n.(*parser.BracketExpr))
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

func (v *Codegen) genCallExpr(n *parser.CallExpr) {
	v.write("%s(", n.Function.Name)
	for i, arg := range n.Arguments {
		v.genExpr(arg)
		if i < len(n.Arguments)-1 {
			v.write(", ")
		}
	}
	v.write(")")
}

func (v *Codegen) genAccessExpr(n *parser.AccessExpr) {
	for _, struc := range n.StructVariables {
		v.write("%s.", struc.Name)
	}
	v.write(n.Variable.Name)
}

func (v *Codegen) genDerefExpr(n *parser.DerefExpr) {
	v.write("^")
	v.genExpr(n.Expr)
}

func (v *Codegen) genBracketExpr(n *parser.BracketExpr) {
	v.write("(")
	v.genExpr(n.Expr)
	v.write(")")
}
