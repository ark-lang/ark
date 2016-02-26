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

type DocComment struct {
	Contents string
	Where    lexer.Span
}

type Documentable interface {
	DocComments() []*DocComment
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

func (v LocatedString) IsEmpty() bool {
	return v.Value == ""
}

// main tree
type ParseTree struct {
	baseNode
	Source *lexer.Sourcefile
	Nodes  []ParseNode
	//Name   string
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

// directives
type LinkDirectiveNode struct {
	baseNode
	Library LocatedString
}

type UseDirectiveNode struct {
	baseNode
	Module *NameNode
}

// types
type ReferenceTypeNode struct {
	baseNode
	TargetType *TypeReferenceNode
	Mutable    bool
}

type PointerTypeNode struct {
	baseNode
	TargetType *TypeReferenceNode
	Mutable    bool
}

type TupleTypeNode struct {
	baseNode
	MemberTypes []*TypeReferenceNode
}

type FunctionTypeNode struct {
	baseNode
	ParameterTypes []*TypeReferenceNode
	ReturnType     *TypeReferenceNode
	IsVariadic     bool
}

type ArrayTypeNode struct {
	baseNode
	MemberType    *TypeReferenceNode
	IsFixedLength bool
	Length        int
}

type NamedTypeNode struct {
	baseNode
	Name *NameNode
}

type TypeReferenceNode struct {
	baseNode
	Type             ParseNode
	GenericArguments []*TypeReferenceNode
}

// decls

type DeclNode interface {
	ParseNode
	IsPublic() bool // only used for top-level nodes
	SetPublic(bool)
}

type baseDecl struct {
	baseNode
	public bool
}

func (v *baseDecl) SetPublic(p bool) {
	v.public = p
}

func (v baseDecl) IsPublic() bool {
	return v.public
}

type InterfaceTypeNode struct {
	baseNode
	Functions    []*FunctionHeaderNode
	GenericSigil *GenericSigilNode
}

type StructTypeNode struct {
	baseNode
	Members      []*StructMemberNode
	GenericSigil *GenericSigilNode
}

type StructMemberNode struct {
	baseNode
	Public bool
	Name   LocatedString
	Type   *TypeReferenceNode
}

type FunctionHeaderNode struct {
	baseNode
	Anonymous    bool
	Name         LocatedString
	GenericSigil *GenericSigilNode
	Arguments    []*VarDeclNode
	ReturnType   *TypeReferenceNode
	Variadic     bool

	StaticReceiverType *NamedTypeNode // use this if static
	Receiver           *VarDeclNode   // use this if not static. this would be so much nicer with tagged unions...
}

type FunctionNode struct {
	baseNode
	Header *FunctionHeaderNode
	Body   *BlockNode
	Stat   ParseNode
	Expr   ParseNode
}

type FunctionDeclNode struct {
	baseDecl
	Function *FunctionNode
}

type LambdaExprNode struct {
	baseNode
	Function *FunctionNode
}

type EnumTypeNode struct {
	baseNode
	Members      []*EnumEntryNode
	GenericSigil *GenericSigilNode
}

type EnumEntryNode struct {
	baseNode
	Name       LocatedString
	Value      *NumberLitNode
	TupleBody  *TupleTypeNode
	StructBody *StructTypeNode
}

type VarDeclNode struct {
	baseDecl
	Name    LocatedString
	Type    *TypeReferenceNode
	Value   ParseNode
	Mutable LocatedString

	IsReceiver           bool
	ReceiverGenericSigil *GenericSigilNode
}

type DestructVarDeclNode struct {
	baseDecl
	Names   []LocatedString
	Mutable []bool
	Value   ParseNode
}

type TypeDeclNode struct {
	baseDecl
	Name         LocatedString
	GenericSigil *GenericSigilNode
	Type         ParseNode
}

type GenericSigilNode struct {
	baseNode
	GenericParameters []*TypeParameterNode
}

type TypeParameterNode struct {
	baseNode
	Name        LocatedString
	Constraints []ParseNode // should be all interface types
}

// statements
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

type BreakStatNode struct {
	baseNode
}

type NextStatNode struct {
	baseNode
}

// expressions
type BinaryExprNode struct {
	baseNode
	Lhand    ParseNode
	Rhand    ParseNode
	Operator BinOpType
}

type ArrayLenExprNode struct {
	baseNode
	ArrayExpr ParseNode
}

type SizeofExprNode struct {
	baseNode
	Value ParseNode
	Type  *TypeReferenceNode
}

type AddrofExprNode struct {
	baseNode
	Value       ParseNode
	Mutable     bool
	IsReference bool
}

type CastExprNode struct {
	baseNode
	Type  *TypeReferenceNode
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

type GenericNameNode struct { // TODO what is this
	baseNode
	Name              *NameNode
	GenericParameters []*TypeReferenceNode
}

// access expressions
type VariableAccessNode struct {
	baseNode
	Name              *NameNode
	GenericParameters []*TypeReferenceNode // TODO rename to GArguments
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

type DiscardAccessNode struct {
	baseNode
}

type EnumPatternNode struct {
	baseNode
	MemberName *NameNode
	Names      []LocatedString
}

// literals

type TupleLiteralNode struct {
	baseNode
	Values []ParseNode
}

type CompositeLiteralNode struct {
	baseNode
	Type   *TypeReferenceNode
	Fields []LocatedString // has same length as Values. missing fields have zero value.
	Values []ParseNode
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
	Value     string
	IsCString bool
}

type RuneLitNode struct {
	baseNode
	Value rune
}
