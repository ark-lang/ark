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
	value rune
}

func (v *RuneLiteral) exprNode() {}

type IntegerLiteral struct {
	value uint64
}

func (v *IntegerLiteral) exprNode() {}

type FloatingLiteral struct {
	value float64
}

func (v *FloatingLiteral) exprNode() {}

type StringLiteral struct {
	value string
}

func (v *StringLiteral) exprNode() {}

