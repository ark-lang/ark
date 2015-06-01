package parser

import (
	"fmt"

	"github.com/ark-lang/ark-go/util"
)

type Locatable interface {
	Pos() (line, char int)
	setPos(line, char int)
}

type Node interface {
	String() string
	NodeName() string
	analyze(*semanticAnalyzer)
	Locatable
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

type nodePos struct {
	lineNumber, charNumber int
}

func (v nodePos) Pos() (line, char int) {
	return v.lineNumber, v.charNumber
}

func (v *nodePos) setPos(line, char int) {
	v.lineNumber = line
	v.charNumber = char
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
	nodePos
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

func (v *Block) NodeName() string {
	return "block"
}

/**
 * Declarations
 */

// VariableDecl

type VariableDecl struct {
	nodePos
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

func (v *VariableDecl) NodeName() string {
	return "variable declaration"
}

// StructDecl

type StructDecl struct {
	nodePos
	Struct *StructType
}

func (v *StructDecl) declNode() {}

func (v *StructDecl) String() string {
	return "(" + util.Blue("StructDecl") + ": " + v.Struct.String() + ")"
}

func (v *StructDecl) NodeName() string {
	return "struct declaration"
}

// FunctionDecl

type FunctionDecl struct {
	nodePos
	Function *Function
}

func (v *FunctionDecl) declNode() {}

func (v *FunctionDecl) String() string {
	return "(" + util.Blue("FunctionDecl") + ": " + v.Function.String() + ")"
}

func (v *FunctionDecl) NodeName() string {
	return "function declaration"
}

/**
 * Statements
 */

// ReturnStat

type ReturnStat struct {
	nodePos
	Value Expr
}

func (v *ReturnStat) statNode() {}

func (v *ReturnStat) String() string {
	return "(" + util.Blue("ReturnStat") + ": " +
		v.Value.String() + ")"
}

func (v *ReturnStat) NodeName() string {
	return "return statement"
}

// CallStat

type CallStat struct {
	nodePos
	Call *CallExpr
}

func (v *CallStat) statNode() {}

func (v *CallStat) String() string {
	return "(" + util.Blue("CallStat") + ": " +
		v.Call.String() + ")"
}

func (v *CallStat) NodeName() string {
	return "call statement"
}

// AssignStat

type AssignStat struct {
	nodePos
	Deref      *DerefExpr // one of these should be nil, not neither or both
	Access     *AccessExpr
	Assignment Expr
}

func (v *AssignStat) statNode() {}

func (v *AssignStat) String() string {
	result := "(" + util.Blue("AssignStat") + ": " //)" // + v.Expr.String() + ")"
	if v.Deref != nil {
		result += v.Deref.String()

	} else if v.Access != nil {
		result += v.Access.String()
	}
	return result + " = " + v.Assignment.String() + ")"
}

func (v *AssignStat) NodeName() string {
	return "assignment statement"
}

/**
 * Expressions
 */

// RuneLiteral

type RuneLiteral struct {
	nodePos
	Value    rune
	typeHint Type
}

func (v *RuneLiteral) exprNode() {}

func (v *RuneLiteral) String() string {
	return fmt.Sprintf("(" + util.Blue("RuneLiteral") + ": " + colorizeEscapedString(escapeString(string(v.Value))) + " " + util.Green(v.GetType().TypeName()) + ")")
}

func (v *RuneLiteral) GetType() Type {
	return PRIMITIVE_rune
}

func (v *RuneLiteral) NodeName() string {
	return "rune literal"
}

// IntegerLiteral

type IntegerLiteral struct {
	nodePos
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

func (v *IntegerLiteral) NodeName() string {
	return "integer literal"
}

// FloatingLiteral

type FloatingLiteral struct {
	nodePos
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

func (v *FloatingLiteral) NodeName() string {
	return "floating-point literal"
}

// StringLiteral

type StringLiteral struct {
	nodePos
	Value string
}

func (v *StringLiteral) exprNode() {}

func (v *StringLiteral) String() string {
	return "(" + util.Blue("StringLiteral") + ": " + colorizeEscapedString((escapeString(v.Value))) + " " + util.Green(v.GetType().TypeName()) + ")"
}

func (v *StringLiteral) GetType() Type {
	return PRIMITIVE_str
}

func (v *StringLiteral) NodeName() string {
	return "string literal"
}

// BinaryExpr

type BinaryExpr struct {
	nodePos
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

func (v *BinaryExpr) NodeName() string {
	return "binary expression"
}

// UnaryExpr

type UnaryExpr struct {
	nodePos
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

func (v *UnaryExpr) NodeName() string {
	return "unary expression"
}

// CastExpr

type CastExpr struct {
	nodePos
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

func (v *CastExpr) NodeName() string {
	return "typecast expression"
}

// CallExpr

type CallExpr struct {
	nodePos
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

func (v *CallExpr) NodeName() string {
	return "call expression"
}

// AccessExpr

type AccessExpr struct {
	nodePos
	StructVariables []*Variable
	Variable        *Variable
}

func (v *AccessExpr) exprNode() {}

func (v *AccessExpr) String() string {
	result := "(" + util.Blue("AccessExpr") + ": "
	for _, struc := range v.StructVariables {
		result += struc.Name + "."
	}
	result += v.Variable.Name
	return result + ")"
}

func (v *AccessExpr) GetType() Type {
	return v.Variable.Type
}

func (v *AccessExpr) NodeName() string {
	return "access expression"
}

// DerefExpr

type DerefExpr struct {
	nodePos
	Expr Expr
	Type Type
}

func (v *DerefExpr) exprNode() {}

func (v *DerefExpr) String() string {
	return "(" + util.Blue("DerefExpr") + ": " + v.Expr.String() + ")"
}

func (v *DerefExpr) GetType() Type {
	return v.Type
}

func (v *DerefExpr) NodeName() string {
	return "dereference expression"
}
