package parser

import (
	"fmt"

	"github.com/ark-lang/ark-go/util"
)

type Node interface {
	String() string
	analyze(*semanticAnalyzer)
}

type Stat interface {
	Node
	statNode()
}

type Expr interface {
	Node
	exprNode()
	GetType() Type
	setTypeHint(Type) // the type of the parent node, nil if parent node's type is inferred
}

type Decl interface {
	Node
	declNode()
}

type Variable struct {
	Type    Type
	Name    string
	Mutable bool
	Attrs   []*Attr
}

func (v *Variable) String() string {
	result := "(" + util.Blue("Variable") + ": "
	if v.Mutable {
		result += util.Green("[mutable] ")
	}
	for _, attr := range v.Attrs {
		result += attr.String() + " "
	}
	return result + v.Name + " " + util.Green(v.Type.TypeName()) + ")"
}

type Function struct {
	Name       string
	Parameters []*VariableDecl
	ReturnType Type
	Mutable    bool
	Attrs      []*Attr
	Body       *Block
}

func (v *Function) String() string {
	result := "(" + util.Blue("Function") + ": "
	if v.Mutable {
		result += util.Green("[mutable] ")
	}
	for _, attr := range v.Attrs {
		result += attr.String() + " "
	}
	result += v.Name
	for _, par := range v.Parameters {
		result += " " + par.String()
	}
	result += ": " + util.Green(v.ReturnType.TypeName()) + " "
	if v.Body != nil {
		result += v.Body.String()
	}
	return result + ")"
}

//
// Nodes
//

type Block struct {
	Nodes []Node
}

func newBlock() *Block {
	return &Block{Nodes: make([]Node, 0)}
}

func (v *Block) String() string {
	if len(v.Nodes) == 0 {
		return "(" + util.Blue("Block") + ": )"
	}

	result := "(" + util.Blue("Block") + ":\n"
	for _, n := range v.Nodes {
		result += "\t" + n.String() + "\n"
	}
	return result + ")"
}

func (v *Block) appendNode(n Node) {
	v.Nodes = append(v.Nodes, n)
}

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
	Struct *StructType
}

func (v *StructDecl) declNode() {}

func (v *StructDecl) String() string {
	return "(" + util.Blue("StructDecl") + ": " + v.Struct.String() + ")"
}

// FunctionDecl

type FunctionDecl struct {
	Function *Function
}

func (v *FunctionDecl) declNode() {}

func (v *FunctionDecl) String() string {
	return "(" + util.Blue("FunctionDecl") + ": " + v.Function.String() + ")"
}

/**
 * Statements
 */

type ReturnStat struct {
	Value Expr
}

func (v *ReturnStat) statNode() {}

func (v *ReturnStat) String() string {
	return "(" + util.Blue("ReturnStat") + ": " +
		v.Value.String() + ")"
}

/**
 * Expressions
 */

// RuneLiteral

type RuneLiteral struct {
	Value    rune
	typeHint Type
}

func (v *RuneLiteral) exprNode() {}

func (v *RuneLiteral) String() string {
	return fmt.Sprintf("("+util.Blue("RuneLiteral")+": "+util.Yellow("%c")+" "+util.Green(v.GetType().TypeName())+")", v.Value)
}

func (v *RuneLiteral) GetType() Type {
	return PRIMITIVE_rune
}

// IntegerLiteral

type IntegerLiteral struct {
	Value    uint64
	Type     Type
	typeHint Type
}

func (v *IntegerLiteral) exprNode() {}

func (v *IntegerLiteral) String() string {
	return fmt.Sprintf("("+util.Blue("IntegerLiteral")+": "+util.Yellow("%d")+" "+util.Green(v.GetType().TypeName())+")", v.Value)
}

func (v *IntegerLiteral) GetType() Type {
	return v.Type
}

// FloatingLiteral

type FloatingLiteral struct {
	Value    float64
	Type     Type
	typeHint Type
}

func (v *FloatingLiteral) exprNode() {}

func (v *FloatingLiteral) String() string {
	return fmt.Sprintf("("+util.Blue("FloatingLiteral")+": "+util.Yellow("%f")+" "+util.Green(v.GetType().TypeName())+")", v.Value)
}

func (v *FloatingLiteral) GetType() Type {
	return v.Type
}

// StringLiteral

type StringLiteral struct {
	Value string
}

func (v *StringLiteral) exprNode() {}

func (v *StringLiteral) String() string {
	return "(" + util.Blue("StringLiteral") + ": " + util.Yellow(v.Value) + " " + util.Green(v.GetType().TypeName()) + ")"
}

func (v *StringLiteral) GetType() Type {
	return PRIMITIVE_str
}

// BinaryExpr

type BinaryExpr struct {
	Lhand, Rhand Expr
	Op           BinOpType
	Type         Type
	typeHint     Type
}

func (v *BinaryExpr) exprNode() {}

func (v *BinaryExpr) String() string {
	return "(" + util.Blue("BinaryExpr") + ": " + v.Lhand.String() + " " +
		v.Op.String() + " " +
		v.Rhand.String() + ")"
}

func (v *BinaryExpr) GetType() Type {
	return v.Type
}

// UnaryExpr

type UnaryExpr struct {
	Expr     Expr
	Op       UnOpType
	Type     Type
	typeHint Type
}

func (v *UnaryExpr) exprNode() {}

func (v *UnaryExpr) String() string {
	return "(" + util.Blue("UnaryExpr") + ": " +
		v.Op.String() + " " + v.Expr.String() + ")"
}

func (v *UnaryExpr) GetType() Type {
	return v.Type
}

// CastExpr

type CastExpr struct {
	Expr Expr
	Type Type
}

func (v *CastExpr) exprNode() {}

func (v *CastExpr) String() string {
	return "(" + util.Blue("CastExpr") + ": " + v.Expr.String() + " " + util.Green(v.GetType().TypeName()) + ")"
}

func (v *CastExpr) GetType() Type {
	return v.Type
}

// CallExpr

type CallExpr struct {
	Function  *Function
	Arguments []Expr
}

func (v *CallExpr) exprNode() {}

func (v *CallExpr) String() string {
	result := "(" + util.Blue("CallExpr") + ": " + v.Function.Name
	for _, arg := range v.Arguments {
		result += " " + arg.String()
	}
	return result + " " + util.Green(v.GetType().TypeName()) + ")"
}

func (v *CallExpr) GetType() Type {
	return v.Function.ReturnType
}
