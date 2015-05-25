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
	Type Type
	Name string
	Mutable bool
}


// Nodes

type VariableDecl struct {
	Variable *Variable
	Assignment Expr
}

func (v *VariableDecl) declNode() {}
