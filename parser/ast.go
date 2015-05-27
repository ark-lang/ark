package parser

import (
	"fmt"
	"container/List"

	"github.com/ark-lang/ark-go/util"
)

type Node interface {
	String() string
}

type Stat interface {
	Node
	statNode()
}

type Expr interface {
	Node
	exprNode()
}

type Decl interface {
	Node
	declNode()
}

type Variable struct {
	Type    Type
	Name    string
	Mutable bool
}

func (v *Variable) String() string {
	mut := ""
	if v.Mutable {
		mut = util.Green("[mutable] ")
	}
	return "(" + util.Blue("Variable") + ": " + mut + v.Name + ": " + util.Green(v.Type.GetTypeName()) + ")"
}

//
// Nodes
//

/**
 * Declarations
 */

// VariableDecl

type VariableDecl struct {
	Variable   *Variable
	Assignment Expr
}

func (v *VariableDecl) declNode() {}

func (v *VariableDecl) String() string {
	if v.Assignment == nil {
		return "(" + util.Blue("VariableDecl") + ": " + v.Variable.String() + ")"
	} else {
		return "(" + util.Blue("VariableDecl") + ": " + v.Variable.String() +
			" = " + v.Assignment.String() + ")"
	}
}

// StructDecl

type StructDecl struct {
	Name string
	Items list.List
	Packed bool
}

func (v *StructDecl) declNode() {}

func (v *StructDecl) String() string {
	var result = "((" + util.Blue("StructDecl") + ": " + v.Name + ")\n"
	for item := v.Items.Front(); item != nil; item = item.Next() {
		result += "\t" + item.Value.(*VariableDecl).String() + "\n"
	}
	result += ")"
	return result
}

// FunctionDecl

type FunctionDecl struct {
	Name string
	Parameters *List
	Type Type
	Mutable bool
}

func (v *FunctionDecl) declNode() {}

func (v *FunctionDecl) String() string {
	mut := ""
	if v.Mutable {
		mut = util.Green("[mutable] ")
	}
	return "(" + util.Blue("FunctionDecl") + ": " + mut + v.Name + " " + v.Parameters.String() + ")" + " (" + util.Blue("Return Type:") + " " + util.Green(v.Type.GetTypeName()) + ")"
}

/**
 * Expressions
 */

// RuneLiteral

type RuneLiteral struct {
	Value rune
}

func (v *RuneLiteral) exprNode() {}

func (v *RuneLiteral) String() string {
	return fmt.Sprintf("("+util.Blue("RuneLiteral")+": "+util.Yellow("%c")+")", v.Value)
}

// IntegerLiteral

type IntegerLiteral struct {
	Value uint64
}

func (v *IntegerLiteral) exprNode() {}

func (v *IntegerLiteral) String() string {
	return fmt.Sprintf("("+util.Blue("IntegerLiteral")+": "+util.Yellow("%d")+")", v.Value)
}

// FloatingLiteral

type FloatingLiteral struct {
	Value float64
}

func (v *FloatingLiteral) exprNode() {}

func (v *FloatingLiteral) String() string {
	return fmt.Sprintf("("+util.Blue("FloatingLiteral")+": "+util.Yellow("%f")+")", v.Value)
}

// StringLiteral

type StringLiteral struct {
	Value string
}

func (v *StringLiteral) exprNode() {}

func (v *StringLiteral) String() string {
	return "(" + util.Blue("StringLiteral") + ": " + util.Yellow(v.Value) + ")"
}

// BinaryExpr

type BinaryExpr struct {
	Lhand, Rhand Expr
	Op           BinOpType
}

func (v *BinaryExpr) exprNode() {}

func (v *BinaryExpr) String() string {
	return "(" + util.Blue("BinaryExpr") + ": " + v.Lhand.String() + " " +
		v.Op.String() + " " +
		v.Rhand.String() + ")"
}

// List

type List struct {
	Items list.List
}

func (v *List) listNode() {}

func (v *List) String() string {
	var result = "(" + util.Blue("List: ");
	for item := v.Items.Front(); item != nil; item = item.Next() {
		result += item.Value.(*VariableDecl).String()
	}
	result += ")"
	return result
}

// UnaryExpr

type UnaryExpr struct {
	Expr Expr
	Op   UnOpType
}

func (v *UnaryExpr) exprNode() {}

func (v *UnaryExpr) String() string {
	return "(" + util.Blue("UnaryExpr") + ": " +
		v.Op.String() + " " + v.Expr.String() + ")"
}
