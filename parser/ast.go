package parser

type Node interface {
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

//
// Nodes
//

/**
 * Declarations
 */

type VariableDecl struct {
	Variable   *Variable
	Assignment Expr
}

func (v *VariableDecl) declNode() {}

/**
 * Expressions
 */

type RuneLiteral struct {
	Value rune
}

func (v *RuneLiteral) exprNode() {}

type IntegerLiteral struct {
	Value uint64
}

func (v *IntegerLiteral) exprNode() {}

type FloatingLiteral struct {
	Value float64
}

func (v *FloatingLiteral) exprNode() {}

type StringLiteral struct {
	Value string
}

func (v *StringLiteral) exprNode() {}

