package parser

import (
	"math/big"

	"github.com/ark-lang/ark/src/lexer"
)

type ParseNode interface {
	Where() lexer.Span
	SetWhere(lexer.Span)

	Attrs() AttrGroup
	SetAttrs(AttrGroup)

	Documentable
	SetDocComments([]*DocComment)
}

// utility
type baseNode struct {
	where lexer.Span
	attrs AttrGroup
	dcs   []*DocComment
}

func (v *baseNode) Where() lexer.Span                { return v.where }
func (v *baseNode) SetWhere(where lexer.Span)        { v.where = where }
func (v *baseNode) Attrs() AttrGroup                 { return v.attrs }
func (v *baseNode) SetAttrs(attrs AttrGroup)         { v.attrs = attrs }
func (v *baseNode) DocComments() []*DocComment       { return v.dcs }
func (v *baseNode) SetDocComments(dcs []*DocComment) { v.dcs = dcs }

type LocatedString struct {
	Where lexer.Span
	Value string
}

func NewLocatedString(token *lexer.Token) LocatedString {
	return LocatedString{Where: token.Where, Value: token.Contents}
}

// main tree
type ParseTree struct {
	baseNode
	Source *lexer.Sourcefile
	Nodes  []ParseNode
}

func (v *ParseTree) AddNode(node ParseNode) {
	v.Nodes = append(v.Nodes, node)
}

// for handling modules
type NameNode struct {
	baseNode
	Modules []LocatedString
	Name    LocatedString
}

// types
type ReferenceTypeNode struct {
	baseNode
	TargetType ParseNode
}

type PointerTypeNode struct {
	baseNode
	TargetType ParseNode
}

type TupleTypeNode struct {
	baseNode
	MemberTypes []ParseNode
}

type ArrayTypeNode struct {
	baseNode
	MemberType ParseNode
	Length     int
}

type TypeReferenceNode struct {
	baseNode
	Reference      *NameNode
	TypeParameters []ParseNode
}

// decls

type StructTypeNode struct {
	baseNode
	Members []*VarDeclNode
}

type UseDeclNode struct {
	baseNode
	Module *NameNode
}

type FunctionHeaderNode struct {
	baseNode
	Name         LocatedString
	GenericSigil *GenericSigilNode
	Arguments    []*VarDeclNode
	ReturnType   ParseNode
	Variadic     bool

	IsMethod           bool
	IsStatic           bool
	StaticReceiverType *TypeReferenceNode // use this if static
	Receiver           *VarDeclNode       // use this if not static. this would be so much nicer with tagged unions...
}

type FunctionDeclNode struct {
	baseNode
	Header *FunctionHeaderNode
	Body   *BlockNode
	Stat   ParseNode
	Expr   ParseNode
}

type EnumTypeNode struct {
	baseNode
	Members []*EnumEntryNode
}

type EnumEntryNode struct {
	baseNode
	Name       LocatedString
	Value      *NumberLitNode
	TupleBody  *TupleTypeNode
	StructBody *StructTypeNode
}

type VarDeclNode struct {
	baseNode
	Name    LocatedString
	Type    ParseNode
	Value   ParseNode
	Mutable LocatedString
}

type TypeDeclNode struct {
	baseNode
	Name         LocatedString
	GenericSigil *GenericSigilNode
	Type         ParseNode
}

type GenericSigilNode struct {
	baseNode
	Parameters []*TypeParameterNode
}

type TypeParameterNode struct {
	baseNode
	Name         LocatedString
	Restrictions []*NameNode
}

// statements

type DefaultStatNode struct {
	baseNode
	Target ParseNode
}

type DeferStatNode struct {
	baseNode
	Call *CallExprNode
}

type IfStatNode struct {
	baseNode
	Parts    []*ConditionBodyNode
	ElseBody *BlockNode
}

type ConditionBodyNode struct {
	baseNode
	Condition ParseNode
	Body      *BlockNode
}

type MatchStatNode struct {
	baseNode
	Value ParseNode
	Cases []*MatchCaseNode
}

type MatchCaseNode struct {
	baseNode
	Pattern ParseNode
	Body    ParseNode
}

type DefaultPatternNode struct {
	baseNode
}

type LoopStatNode struct {
	baseNode
	Condition ParseNode
	Body      *BlockNode
}

type ReturnStatNode struct {
	baseNode
	Value ParseNode
}

type BlockStatNode struct {
	baseNode
	Body *BlockNode
}

type BlockNode struct {
	baseNode
	NonScoping bool
	Nodes      []ParseNode
}

type CallStatNode struct {
	baseNode
	Call *CallExprNode
}

type AssignStatNode struct {
	baseNode
	Target ParseNode
	Value  ParseNode
}

type BinopAssignStatNode struct {
	baseNode
	Target   ParseNode
	Operator BinOpType
	Value    ParseNode
}

// expressions
type BinaryExprNode struct {
	baseNode
	Lhand    ParseNode
	Rhand    ParseNode
	Operator BinOpType
}

type SizeofExprNode struct {
	baseNode
	Value ParseNode
	Type  ParseNode
}

type DefaultExprNode struct {
	baseNode
	Target ParseNode
}

type AddrofExprNode struct {
	baseNode
	Value ParseNode
}

type CastExprNode struct {
	baseNode
	Type  ParseNode
	Value ParseNode
}

type UnaryExprNode struct {
	baseNode
	Value    ParseNode
	Operator UnOpType
}

type CallExprNode struct {
	baseNode
	Function  ParseNode
	Arguments []ParseNode
}

type GenericNameNode struct {
	baseNode
	Name       *NameNode
	Parameters []ParseNode
}

// access expressions
type VariableAccessNode struct {
	baseNode
	Name *NameNode
}

type StructAccessNode struct {
	baseNode
	Struct ParseNode
	Member LocatedString
}

type ArrayAccessNode struct {
	baseNode
	Array ParseNode
	Index ParseNode
}

type TupleAccessNode struct {
	baseNode
	Tuple ParseNode
	Index int
}

// literals
type ArrayLiteralNode struct {
	baseNode
	Values []ParseNode
}

type TupleLiteralNode struct {
	baseNode
	Values []ParseNode
}

type StructLiteralNode struct {
	baseNode
	Name    *NameNode
	Members []LocatedString
	Values  []ParseNode
}

type BoolLitNode struct {
	baseNode
	Value bool
}

type NumberLitNode struct {
	baseNode
	IsFloat    bool
	IntValue   *big.Int
	FloatValue float64
	FloatSize  rune
}

type StringLitNode struct {
	baseNode
	Value string
}

type RuneLitNode struct {
	baseNode
	Value rune
}
