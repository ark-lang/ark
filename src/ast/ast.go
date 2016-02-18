package ast

import (
	"bytes"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ark-lang/ark/src/lexer"
	"github.com/ark-lang/ark/src/parser"
	"github.com/ark-lang/ark/src/util"
)

type Locatable interface {
	Pos() lexer.Position
	SetPos(pos lexer.Position)
}

type Node interface {
	String() string
	NodeName() string
	Locatable
}

type Stat interface {
	Node
	statNode()
}

type Typed interface {
	GetType() *TypeReference
	SetType(*TypeReference)
}

type Expr interface {
	Node
	Typed
	exprNode()
}

type AccessExpr interface {
	Expr
	Mutable() bool
}

type Decl interface {
	Node
	declNode()
	IsPublic() bool
	SetPublic(bool)
}

// an implementation of Locatable that is used for Nodes
type nodePos struct {
	pos lexer.Position
}

func (v nodePos) Pos() lexer.Position {
	return v.pos
}

func (v *nodePos) SetPos(pos lexer.Position) {
	v.pos = pos
}

// TODO fix up all the incorrect &TypeReference{...}
type TypeReference struct {
	BaseType         Type
	GenericArguments []*TypeReference
}

func NewTypeReference(typ Type, gargs []*TypeReference) *TypeReference {
	return &TypeReference{
		BaseType:         typ,
		GenericArguments: gargs,
	}
}

func (v TypeReference) String() string {
	str := v.BaseType.TypeName()
	if len(v.GenericArguments) > 0 {
		str += "<"
		for i, arg := range v.GenericArguments {
			str += arg.String()
			if i < len(v.GenericArguments)-1 {
				str += ", "
			}
		}
		str += ">"
	}
	return str
}

func (v *TypeReference) Equals(t *TypeReference) bool {
	if !v.BaseType.Equals(t.BaseType) {
		return false
	}

	if len(v.GenericArguments) != len(t.GenericArguments) {
		return false
	}

	for i, arg := range v.GenericArguments {
		if !arg.Equals(t.GenericArguments[i]) {
			return false
		}
	}

	return true
}

func (v *TypeReference) ActualTypesEqual(t *TypeReference) bool {
	if !v.BaseType.ActualType().Equals(t.BaseType.ActualType()) {
		return false
	}

	if len(v.GenericArguments) != len(t.GenericArguments) {
		return false
	}

	for i, arg := range v.GenericArguments {
		if !arg.ActualTypesEqual(t.GenericArguments[i]) {
			return false
		}
	}

	return true
}

func (v *TypeReference) CanCastTo(t *TypeReference) bool {
	return v.BaseType.CanCastTo(t.BaseType) && len(v.GenericArguments) == 0 && len(t.GenericArguments) == 0
}

type Variable struct {
	Type         *TypeReference
	Name         string
	Mutable      bool
	Attrs        parser.AttrGroup
	FromStruct   bool
	ParentStruct StructType
	ParentModule *Module
	IsParameter  bool
	IsArgument   bool

	IsReceiver                bool             // TODO separate this out so it isn't as messy
	ReceiverGenericParameters []*TypeReference // only used if IsReceiver == true
}

func (v Variable) String() string {
	s := NewASTStringer("Variable")
	if v.Mutable {
		s.AddStringColored(util.TEXT_GREEN, " [mutable]")
	}
	s.AddAttrs(v.Attrs)
	s.AddString(v.Name)
	s.AddTypeReference(v.Type)
	return s.Finish()
}

func (v Variable) GetType() *TypeReference {
	return v.Type
}

// Note that for static methods, ``
type Function struct {
	Name       string
	Type       FunctionType
	Parameters []*VariableDecl
	Body       *Block
	Accesses   []*FunctionAccessExpr

	ParentModule *Module

	Receiver *VariableDecl // non-nil if non-static method

	StaticReceiverType Type // non-nil if static
}

func (v Function) String() string {
	s := NewASTStringer("Function")
	s.AddAttrs(v.Type.Attrs())
	s.AddString(v.Type.GenericParameters.String())
	s.AddString(v.Name)
	for _, par := range v.Parameters {
		s.Add(par)
	}
	if v.Type.Return != nil {
		s.AddString(":")
		s.AddTypeReference(v.Type.Return)
	}
	s.Add(v.Body)
	return s.Finish()
}

//
// Nodes
//

type Block struct {
	nodePos
	Nodes         []Node
	IsTerminating bool
	NonScoping    bool
}

func (v Block) String() string {
	s := NewASTStringer("Block")
	for _, n := range v.Nodes {
		s.AddString("\n\t")
		s.Add(n)
	}
	return s.Finish()
}

func (v *Block) appendNode(n Node) {
	v.Nodes = append(v.Nodes, n)
}

func (_ Block) NodeName() string {
	return "block"
}

// nil if no nodes
func (v Block) LastNode() Node {
	if len(v.Nodes) == 0 {
		return nil
	}
	return v.Nodes[len(v.Nodes)-1]
}

/**
 * Declarations
 */

type PublicHandler struct {
	public bool
}

func (v *PublicHandler) SetPublic(b bool) {
	v.public = b
}

func (v PublicHandler) IsPublic() bool {
	return v.public
}

// VariableDecl

type VariableDecl struct {
	nodePos
	PublicHandler
	Variable   *Variable
	Assignment Expr
	docs       []*parser.DocComment
}

func (_ VariableDecl) declNode() {}

func (v VariableDecl) String() string {
	s := NewASTStringer("VariableDecl")
	s.Add(v.Variable)
	if v.Assignment != nil {
		s.AddString(" =")
		s.Add(v.Assignment)
	}
	return s.Finish()
}

func (_ VariableDecl) NodeName() string {
	return "variable declaration"
}

func (v VariableDecl) DocComments() []*parser.DocComment {
	return v.docs
}

// DestructVarDecl
type DestructVarDecl struct {
	nodePos
	PublicHandler
	Variables     []*Variable
	ShouldDiscard []bool
	Assignment    Expr
	docs          []*parser.DocComment
}

func (_ DestructVarDecl) declNode() {}

func (v DestructVarDecl) String() string {
	s := NewASTStringer("DestructVarDecl")
	for _, vari := range v.Variables {
		s.Add(vari)
	}
	s.AddString(" =")
	s.Add(v.Assignment)
	return s.Finish()
}

func (_ DestructVarDecl) NodeName() string {
	return "destructuring variable declaration"
}

func (v DestructVarDecl) DocComments() []*parser.DocComment {
	return v.docs
}

// TypeDecl

type TypeDecl struct {
	nodePos
	PublicHandler
	NamedType *NamedType
}

func (_ TypeDecl) declNode() {}

func (v TypeDecl) String() string {
	return NewASTStringer("TypeDecl").Add(v.NamedType).Finish()
}

func (_ TypeDecl) NodeName() string {
	return "named type declaration"
}

func (v TypeDecl) DocComments() []*parser.DocComment {
	return nil // TODO
}

// FunctionDecl

type FunctionDecl struct {
	nodePos
	PublicHandler
	Function  *Function
	Prototype bool
	docs      []*parser.DocComment
}

func (_ FunctionDecl) declNode() {}

func (v FunctionDecl) String() string {
	return NewASTStringer("FunctionDecl").Add(v.Function).Finish()
}

func (_ FunctionDecl) NodeName() string {
	return "function declaration"
}

func (v FunctionDecl) DocComments() []*parser.DocComment {
	return v.docs
}

/**
 * Directives
 */

// UseDirective

type UseDirective struct {
	nodePos
	ModuleName UnresolvedName
}

func (_ UseDirective) declNode() {}

func (v UseDirective) String() string {
	return NewASTStringer("UseDirective").Add(v.ModuleName).Finish()
}

func (_ UseDirective) NodeName() string {
	return "use directive"
}

/**
 * Statements
 */

// BlockStat

type BlockStat struct {
	nodePos
	Block *Block
}

func (_ BlockStat) statNode() {}

func (v BlockStat) String() string {
	return NewASTStringer("BlockStat").Add(v.Block).Finish()
}

func (_ BlockStat) NodeName() string {
	return "block statement"
}

// ReturnStat

type ReturnStat struct {
	nodePos
	Value Expr
}

func (_ ReturnStat) statNode() {}

func (v ReturnStat) String() string {
	return NewASTStringer("ReturnStat").AddWithFallback("", v.Value, "void").Finish()
}

func (_ ReturnStat) NodeName() string {
	return "return statement"
}

// BreakStat

type BreakStat struct {
	nodePos
}

func (_ BreakStat) statNode() {}

func (v BreakStat) String() string {
	return NewASTStringer("BreakStat").Finish()
}

func (_ BreakStat) NodeName() string {
	return "break statement"
}

// NextStat

type NextStat struct {
	nodePos
}

func (_ NextStat) statNode() {}

func (v NextStat) String() string {
	return NewASTStringer("NextStat").Finish()
}

func (_ NextStat) NodeName() string {
	return "next statement"
}

// CallStat

type CallStat struct {
	nodePos
	Call *CallExpr
}

func (_ CallStat) statNode() {}

func (v CallStat) String() string {
	return NewASTStringer("CallStat").Add(v.Call).Finish()
}

func (_ CallStat) NodeName() string {
	return "call statement"
}

// DeferStat

type DeferStat struct {
	nodePos
	Call *CallExpr
}

func (_ DeferStat) statNode() {}

func (v DeferStat) String() string {
	return NewASTStringer("DeferStat").Add(v.Call).Finish()
}

func (_ DeferStat) NodeName() string {
	return "call statement"
}

// AssignStat

type AssignStat struct {
	nodePos
	Access     AccessExpr
	Assignment Expr
}

func (_ AssignStat) statNode() {}

func (v AssignStat) String() string {
	return NewASTStringer("AssignStat").Add(v.Access).AddString(" =").Add(v.Assignment).Finish()
}

func (_ AssignStat) NodeName() string {
	return "assignment statement"
}

// BinopAssignStat

type BinopAssignStat struct {
	nodePos
	Access     AccessExpr
	Operator   parser.BinOpType
	Assignment Expr
}

func (_ BinopAssignStat) statNode() {}

func (v BinopAssignStat) String() string {
	return NewASTStringer("BinopAssignStat").Add(v.Access).Add(v.Operator).AddString(" =").Add(v.Assignment).Finish()
}

func (_ BinopAssignStat) NodeName() string {
	return "binop assignment statement"
}

// DestructAssignStat

type DestructAssignStat struct {
	nodePos
	Accesses   []AccessExpr
	Assignment Expr
}

func (_ DestructAssignStat) statNode() {}

func (v DestructAssignStat) String() string {
	s := NewASTStringer("DestructAssignStat")
	for _, acc := range v.Accesses {
		s.Add(acc)
	}
	return s.AddString(" =").Add(v.Assignment).Finish()
}

func (_ DestructAssignStat) NodeName() string {
	return "destructuring assignment statement"
}

// DestructBinopAssignStat

type DestructBinopAssignStat struct {
	nodePos
	Accesses   []AccessExpr
	Operator   parser.BinOpType
	Assignment Expr
}

func (_ DestructBinopAssignStat) statNode() {}

func (v DestructBinopAssignStat) String() string {
	s := NewASTStringer("DestructBinopAssignStat")
	for _, acc := range v.Accesses {
		s.Add(acc)
	}
	return s.Add(v.Operator).AddString(" =").Add(v.Assignment).Finish()
}

func (_ DestructBinopAssignStat) NodeName() string {
	return "destructuring binop assignment statement"
}

// IfStat

type IfStat struct {
	nodePos
	Exprs  []Expr
	Bodies []*Block
	Else   *Block // can be nil
}

func (_ IfStat) statNode() {}

func (v IfStat) String() string {
	s := NewASTStringer("IfStat")
	for i, expr := range v.Exprs {
		s.Add(expr)
		s.Add(v.Bodies[i])
	}
	s.Add(v.Else)
	return s.Finish()
}

func (_ IfStat) NodeName() string {
	return "if statement"
}

// LoopStat

type LoopStatType int

const (
	LOOP_TYPE_UNSET LoopStatType = iota
	LOOP_TYPE_INFINITE
	LOOP_TYPE_CONDITIONAL
)

type LoopStat struct {
	nodePos
	LoopType LoopStatType

	Body *Block

	// LOOP_TYPE_CONDITIONAL
	Condition Expr
}

func (_ LoopStat) statNode() {}

func (v LoopStat) String() string {
	s := NewASTStringer("LoopStat")
	switch v.LoopType {
	case LOOP_TYPE_INFINITE:
		// noop
	case LOOP_TYPE_CONDITIONAL:
		s.Add(v.Condition)
	default:
		panic("invalid loop type")
	}
	s.Add(v.Body)
	return s.Finish()
}

func (_ LoopStat) NodeName() string {
	return "loop statement"
}

// MatchStat

type MatchStat struct {
	nodePos

	Target Expr

	Branches map[Expr]Node
}

func (_ MatchStat) statNode() {}

func (v MatchStat) String() string {
	s := NewASTStringer("MatchStat")
	s.Add(v.Target)
	for pattern, stmt := range v.Branches {
		s.AddString("\n\t")
		s.Add(pattern)
		s.AddString(" -> ")
		s.Add(stmt)
	}
	return s.Finish()
}

func (_ MatchStat) NodeName() string {
	return "match statement"
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

func (_ RuneLiteral) exprNode() {}

func (v RuneLiteral) String() string {
	return NewASTStringer("RuneLiteral").AddString(
		colorizeEscapedString(EscapeString(string(v.Value))),
	).AddTypeReference(v.GetType()).Finish()
}

func (v RuneLiteral) GetType() *TypeReference {
	return &TypeReference{BaseType: runeType}
}

func (_ RuneLiteral) NodeName() string {
	return "rune literal"
}

// NumericLiteral
type NumericLiteral struct {
	nodePos
	IntValue      *big.Int
	FloatValue    float64
	IsFloat       bool
	Type          *TypeReference
	floatSizeHint rune
}

func (_ NumericLiteral) exprNode() {}

func (v NumericLiteral) String() string {
	s := NewASTStringer("NumericLiteral")
	if v.IsFloat {
		s.AddStringColored(util.TEXT_YELLOW, fmt.Sprintf("%f", v.FloatValue))
	} else {
		s.AddStringColored(util.TEXT_YELLOW, fmt.Sprintf("%d", v.IntValue))
	}
	s.AddTypeReference(v.GetType())
	return s.Finish()
}

func (v NumericLiteral) GetType() *TypeReference {
	if v.Type != nil {
		return v.Type
	} else if v.IsFloat {
		typ := PRIMITIVE_f32
		switch v.floatSizeHint {
		case 'f':
			typ = PRIMITIVE_f32
		case 'd':
			typ = PRIMITIVE_f64
		case 'q':
			typ = PRIMITIVE_f128
		}
		return &TypeReference{BaseType: typ}
	} else {
		return &TypeReference{BaseType: PRIMITIVE_int}
	}
}

func (_ NumericLiteral) NodeName() string {
	return "numeric literal"
}

func (v NumericLiteral) AsFloat() float64 {
	if v.IsFloat {
		return v.FloatValue
	} else {
		return float64(v.IntValue.Int64())
	}
}

func (v NumericLiteral) AsInt() uint64 {
	if v.IsFloat {
		panic("downcasting floating point value to int")
	}

	return uint64(v.IntValue.Int64())
}

// StringLiteral

type StringLiteral struct {
	nodePos
	Value     string
	IsCString bool
	Type      *TypeReference
}

func (_ StringLiteral) exprNode() {}

func (v StringLiteral) String() string {
	return NewASTStringer("StringLiteral").AddString(colorizeEscapedString(EscapeString(v.Value))).AddTypeReference(v.GetType()).Finish()
}

func (v StringLiteral) GetType() *TypeReference {
	if v.Type != nil {
		return v.Type
	} else if v.IsCString {
		return &TypeReference{BaseType: PointerTo(&TypeReference{BaseType: PRIMITIVE_u8}, false)}
	} else {
		return &TypeReference{BaseType: stringType}
	}
}

func (_ StringLiteral) NodeName() string {
	return "string literal"
}

// BoolLiteral

type BoolLiteral struct {
	nodePos
	Value bool
}

func (_ BoolLiteral) exprNode() {}

func (v BoolLiteral) String() string {
	return NewASTStringer("BoolLiteral").AddStringColored(util.TEXT_YELLOW, strconv.FormatBool(v.Value)).Finish()
}

func (v BoolLiteral) GetType() *TypeReference {
	return &TypeReference{BaseType: PRIMITIVE_bool}
}

func (_ BoolLiteral) NodeName() string {
	return "boolean literal"
}

// TupleLiteral

type TupleLiteral struct {
	nodePos
	Members           []Expr
	Type              *TypeReference
	ParentEnumLiteral *EnumLiteral // only non-nil if this part of an enum literal
}

func (_ TupleLiteral) exprNode() {}

func (v TupleLiteral) String() string {
	s := NewASTStringer("TupleLiteral")
	for _, mem := range v.Members {
		s.Add(mem)
		s.AddString(",")
	}
	return s.Finish()
}

func (v TupleLiteral) GetType() *TypeReference {
	if v.Type != nil {
		return v.Type
	}

	tt := TupleType{Members: make([]*TypeReference, len(v.Members))}
	for idx, mem := range v.Members {
		if mem.GetType() == nil {
			return nil
		}
		tt.Members[idx] = mem.GetType()
	}
	return &TypeReference{BaseType: tt}
}

func (_ TupleLiteral) NodeName() string {
	return "tuple literal"
}

// CompositeLiteral

type CompositeLiteral struct {
	nodePos
	Type   *TypeReference
	Fields []string // len(Fields) == len(Values). empty fields represented as ""
	Values []Expr
	InEnum bool
}

func (_ CompositeLiteral) exprNode() {}

func (v CompositeLiteral) String() string {
	s := NewASTStringer("CompositeLiteral")
	for i, mem := range v.Values {
		s.AddString("\n\t")
		if field := v.Fields[i]; field != "" {
			s.AddString(field)
			s.AddString(":")
		}
		s.Add(mem)
		s.AddString(",")
	}
	s.AddTypeReference(v.Type)
	return s.Finish()
}

func (v CompositeLiteral) GetType() *TypeReference {
	return v.Type
}

func (_ CompositeLiteral) NodeName() string {
	return "composite literal"
}

// EnumLiteral

type EnumLiteral struct {
	nodePos
	Type   *TypeReference
	Member string

	TupleLiteral     *TupleLiteral
	CompositeLiteral *CompositeLiteral
}

func (_ EnumLiteral) exprNode() {}

func (v EnumLiteral) String() string {
	s := NewASTStringer("EnumLiteral")

	if v.TupleLiteral != nil {
		s.Add(v.TupleLiteral)
	} else if v.CompositeLiteral != nil {
		s.Add(v.CompositeLiteral)
	}
	return s.Finish()
}

func (v EnumLiteral) GetType() *TypeReference {
	return v.Type
}

func (_ EnumLiteral) NodeName() string {
	return "enum literal"
}

// BinaryExpr

type BinaryExpr struct {
	nodePos
	Lhand, Rhand Expr
	Op           parser.BinOpType
	Type         *TypeReference
}

func (_ BinaryExpr) exprNode() {}

func (v BinaryExpr) String() string {
	return NewASTStringer("BinaryExpr").Add(v.Op).Add(v.Lhand).Add(v.Rhand).Finish()
}

func (v BinaryExpr) GetType() *TypeReference {
	return v.Type
}

func (_ BinaryExpr) NodeName() string {
	return "binary expression"
}

// UnaryExpr

type UnaryExpr struct {
	nodePos
	Expr Expr
	Op   parser.UnOpType
	Type *TypeReference
}

func (_ UnaryExpr) exprNode() {}

func (v UnaryExpr) String() string {
	return NewASTStringer("UnaryExpr").Add(v.Op).Add(v.Expr).Finish()
}

func (v UnaryExpr) GetType() *TypeReference {
	return v.Type
}

func (_ UnaryExpr) NodeName() string {
	return "unary expression"
}

// CastExpr

type CastExpr struct {
	nodePos
	Expr Expr
	Type *TypeReference
}

func (_ CastExpr) exprNode() {}

func (v CastExpr) String() string {
	return NewASTStringer("CastExpr").Add(v.Expr).AddTypeReference(v.GetType()).Finish()
}

func (v CastExpr) GetType() *TypeReference {
	return v.Type
}

func (_ CastExpr) NodeName() string {
	return "typecast expression"
}

// CallExpr

type CallExpr struct {
	nodePos
	Function       Expr
	Arguments      []Expr
	ReceiverAccess Expr // nil if not method or if static
}

func (_ CallExpr) exprNode() {}

func (v CallExpr) String() string {
	s := NewASTStringer("CallExpr")
	s.Add(v.Function)
	for _, arg := range v.Arguments {
		s.Add(arg)
	}
	s.AddTypeReference(v.GetType())
	return s.Finish()
}

func (v CallExpr) GetType() *TypeReference {
	if v.Function != nil {
		fnType := v.Function.GetType()
		if fnType != nil {
			return fnType.BaseType.(FunctionType).Return
		}
	}
	return nil
}

func (_ CallExpr) NodeName() string {
	return "call expression"
}

// FunctionAccessExpr
type FunctionAccessExpr struct {
	nodePos

	Function       *Function
	ReceiverAccess Expr // should be same as on the callexpr

	GenericArguments []*TypeReference

	ParentFunction *Function // the function this access expression is located in
}

func (_ FunctionAccessExpr) exprNode() {}

func (v FunctionAccessExpr) String() string {
	return NewASTStringer("FunctionAccessExpr").AddGenericArguments(v.GenericArguments).AddString(v.Function.Name).Finish()
}

func (v FunctionAccessExpr) GetType() *TypeReference {
	ref := &TypeReference{
		BaseType:         v.Function.Type,
		GenericArguments: v.GenericArguments,
	}

	if len(v.GenericArguments) > 0 {
		return NewGenericContextFromTypeReference(ref).Replace(ref)
	}
	return ref
}

func (_ FunctionAccessExpr) NodeName() string {
	return "function access expression"
}

// VariableAccessExpr
type VariableAccessExpr struct {
	nodePos
	Name UnresolvedName

	Variable *Variable

	GenericArguments []*TypeReference // TODO check no gen args if variable, not function
}

func (_ VariableAccessExpr) exprNode() {}

func (v VariableAccessExpr) String() string {
	return NewASTStringer("VariableAccessExpr").Add(v.Name).AddGenericArguments(v.GenericArguments).Finish()
}

func (v VariableAccessExpr) GetType() *TypeReference {
	if v.Variable != nil && v.Variable.Type != nil {
		if len(v.Variable.Type.GenericArguments) > 0 {
			return NewGenericContextFromTypeReference(v.Variable.Type).Replace(v.Variable.Type)
		}
		return v.Variable.Type
	}
	return nil
}

func (_ VariableAccessExpr) NodeName() string {
	return "variable access expression"
}

func (v VariableAccessExpr) Mutable() bool {
	return v.Variable.Mutable
}

// StructAccessExpr
type StructAccessExpr struct {
	nodePos
	Struct AccessExpr
	Member string

	GenericArguments []*TypeReference // TODO check no gen args if variable, not function

	// Needed for when we convert an struct access to function access
	ParentFunction *Function
}

func (_ StructAccessExpr) exprNode() {}

func (v StructAccessExpr) String() string {
	s := NewASTStringer("StructAccessExpr")
	s.AddString("struct").Add(v.Struct)
	s.AddString("member").AddString(v.Member)
	return s.Finish()
}

func (v StructAccessExpr) GetType() *TypeReference {
	// TODO sort out type references and stuff
	stype := v.Struct.GetType()
	if stype == nil {
		return nil
	}

	if typ, ok := TypeWithoutPointers(stype.BaseType).(*NamedType); ok {
		fn := typ.GetMethod(v.Member)
		if fn != nil {
			return &TypeReference{BaseType: fn.Type, GenericArguments: v.GenericArguments}
		}
	}

	if stype == nil {
		return nil
	} else if pt, ok := stype.BaseType.(PointerType); ok {
		stype = pt.Addressee
	}

	if stype == nil {
		return nil
	} else if st, ok := stype.BaseType.ActualType().(StructType); ok {
		mem := st.GetMember(v.Member)
		if mem != nil {
			return NewGenericContextFromTypeReference(stype).Replace(mem.Type)
		}
	}

	return nil
}

func (_ StructAccessExpr) NodeName() string {
	return "struct access expression"
}

func (v StructAccessExpr) Mutable() bool {
	return v.Struct.Mutable()
}

// ArrayAccessExpr

type ArrayAccessExpr struct {
	nodePos
	Array     AccessExpr
	Subscript Expr
}

func (_ ArrayAccessExpr) exprNode() {}

func (v ArrayAccessExpr) String() string {
	s := NewASTStringer("ArrayAccessExpr")
	s.AddString("array").Add(v.Array)
	s.AddString("index").Add(v.Subscript)
	return s.Finish()
}

func (v ArrayAccessExpr) GetType() *TypeReference {
	if v.Array.GetType() != nil {
		return v.Array.GetType().BaseType.ActualType().(ArrayType).MemberType
	}
	return nil
}

func (_ ArrayAccessExpr) NodeName() string {
	return "array access expression"
}

func (v ArrayAccessExpr) Mutable() bool {
	return v.Array.Mutable()
}

// DerefAccessExpr

type DerefAccessExpr struct {
	nodePos
	Expr Expr
}

func (_ DerefAccessExpr) exprNode() {}

func (v DerefAccessExpr) String() string {
	return NewASTStringer("DerefAccessExpr").Add(v.Expr).Finish()
}

func (v DerefAccessExpr) GetType() *TypeReference {
	ret := getAdressee(v.Expr.GetType().BaseType)
	if ret == nil {
		return &TypeReference{BaseType: PRIMITIVE_void}
	}
	return ret
}

func (_ DerefAccessExpr) NodeName() string {
	return "dereference access expression"
}

func (v DerefAccessExpr) Mutable() bool {
	access, ok := v.Expr.(AccessExpr)
	if ok {
		if rt, ok := access.GetType().BaseType.(ReferenceType); ok {
			return rt.IsMutable
		}

		if pt, ok := access.GetType().BaseType.(PointerType); ok {
			return pt.IsMutable
		}

		return access.Mutable()
	} else {
		return true
	}
}

func getAdressee(t Type) *TypeReference {
	switch t := t.(type) {
	case PointerType:
		return t.Addressee
	case ReferenceType:
		return t.Referrer
	}
	return nil
}

// DiscardAccess

type DiscardAccessExpr struct {
	nodePos
}

func (_ DiscardAccessExpr) exprNode() {}

func (v DiscardAccessExpr) String() string {
	return NewASTStringer("DiscardAccessExpr").Finish()
}

func (v DiscardAccessExpr) GetType() *TypeReference {
	return nil
}

func (_ DiscardAccessExpr) NodeName() string {
	return "discard access expression"
}

func (v DiscardAccessExpr) Mutable() bool {
	return true
}

// EnumPatternExpr

type EnumPatternExpr struct {
	nodePos

	MemberName UnresolvedName
	Variables  []*Variable

	EnumType *TypeReference
}

func (_ EnumPatternExpr) exprNode() {}

func (v EnumPatternExpr) String() string {
	return NewASTStringer("EnumPatternExpr").Finish()
}

func (v EnumPatternExpr) GetType() *TypeReference {
	return nil
}

func (_ EnumPatternExpr) NodeName() string {
	return "enum match pattern"
}

// ReferenceToExpr

type ReferenceToExpr struct {
	nodePos
	IsMutable bool
	Access    Expr
}

func (_ ReferenceToExpr) exprNode() {}

func (v ReferenceToExpr) String() string {
	return NewASTStringer("ReferenceToExpr").Add(v.Access).AddTypeReference(v.GetType()).Finish()
}

func (v ReferenceToExpr) GetType() *TypeReference {
	if v.Access.GetType() != nil {
		return &TypeReference{BaseType: ReferenceTo(v.Access.GetType(), v.IsMutable)}
	}
	return nil
}

func (_ ReferenceToExpr) NodeName() string {
	return "reference-to expression"
}

// PointerToExpr

type PointerToExpr struct {
	nodePos
	IsMutable bool
	Access    Expr
}

func (_ PointerToExpr) exprNode() {}

func (v PointerToExpr) String() string {
	return NewASTStringer("PointerToExpr").Add(v.Access).AddTypeReference(v.GetType()).Finish()
}

func (v PointerToExpr) GetType() *TypeReference {
	if v.Access.GetType() != nil {
		return &TypeReference{BaseType: PointerTo(v.Access.GetType(), v.IsMutable)}
	}
	return nil
}

func (_ PointerToExpr) NodeName() string {
	return "pointer-to expression"
}

// LambdaExpr

type LambdaExpr struct {
	nodePos

	Function *Function
}

func (_ LambdaExpr) exprNode() {}

func (v LambdaExpr) String() string {
	return NewASTStringer("LambdaExpr").Add(v.Function).Finish()
}

func (v LambdaExpr) GetType() *TypeReference {
	return &TypeReference{BaseType: v.Function.Type}
}

func (v LambdaExpr) NodeName() string {
	return "lambda expr"
}

// ArrayLenExpr

type ArrayLenExpr struct {
	nodePos

	Expr Expr
	Type Type
}

func (_ ArrayLenExpr) exprNode() {}

func (v ArrayLenExpr) String() string {
	s := NewASTStringer("ArrayLenExpr")
	if v.Expr != nil {
		s.Add(v.Expr)
	} else {
		s.AddType(v.Type)
	}
	return s.Finish()
}

func (v ArrayLenExpr) GetType() *TypeReference {
	return &TypeReference{BaseType: PRIMITIVE_uint}
}

func (_ ArrayLenExpr) NodeName() string {
	return "array length expr"
}

// SizeofExpr

type SizeofExpr struct {
	nodePos
	// oneither Expr or Type is nil, not neither or both

	Expr Expr

	Type *TypeReference
}

func (_ SizeofExpr) exprNode() {}

func (v SizeofExpr) String() string {
	s := NewASTStringer("SizeofExpr")
	if v.Expr != nil {
		s.Add(v.Expr)
	} else {
		s.AddTypeReference(v.Type)
	}
	return s.Finish()
}

func (v SizeofExpr) GetType() *TypeReference {
	return &TypeReference{BaseType: PRIMITIVE_uint}
}

func (_ SizeofExpr) NodeName() string {
	return "sizeof expression"
}

// String representation util
type ASTStringer struct {
	buf   *bytes.Buffer
	first bool
}

func NewASTStringer(name string) *ASTStringer {
	res := &ASTStringer{
		buf:   new(bytes.Buffer),
		first: true,
	}
	res.buf.WriteRune('(')
	res.buf.WriteString(util.TEXT_BLUE)
	res.buf.WriteString(name)
	res.buf.WriteString(util.TEXT_RESET)
	return res
}

func (v *ASTStringer) AddAttrs(attrs parser.AttrGroup) *ASTStringer {
	for _, attr := range attrs {
		v.Add(attr)
	}
	return v
}

func (v *ASTStringer) Add(what fmt.Stringer) *ASTStringer {
	v.AddWithColor("", what)
	return v
}

func (v *ASTStringer) AddWithColor(color string, what fmt.Stringer) *ASTStringer {
	v.AddWithFallback(color, what, "nil")
	return v
}

func (v *ASTStringer) AddWithFallback(color string, what fmt.Stringer, fallback string) *ASTStringer {
	if !isNil(what) {
		v.AddStringColored(color, what.String())
	} else {
		v.AddStringColored(color, fallback)
	}
	return v
}

func (v *ASTStringer) AddString(what string) *ASTStringer {
	v.AddStringColored("", what)
	return v
}

func (v *ASTStringer) AddStringColored(color string, what string) *ASTStringer {
	if v.first {
		v.buf.WriteRune(':')
		v.first = false
	}
	v.buf.WriteRune(' ')
	v.buf.WriteString(color)
	v.buf.WriteString(what)
	v.buf.WriteString(util.TEXT_RESET)
	return v
}

func (v *ASTStringer) AddType(t Type) *ASTStringer {
	if t == nil {
		v.AddStringColored(util.TEXT_RED, "<type = nil>")
		return v
	}
	v.AddStringColored(util.TEXT_BLUE, t.TypeName())
	return v
}

func (v *ASTStringer) AddTypeReference(t *TypeReference) *ASTStringer {
	if t == nil || t.BaseType == nil {
		v.AddStringColored(util.TEXT_RED, "<type = nil>")
		return v
	}

	v.AddStringColored(util.TEXT_BLUE, t.BaseType.TypeName())
	if len(t.GenericArguments) > 0 {
		v.AddStringColored(util.TEXT_BLUE, "<")
		for _, a := range t.GenericArguments {
			v.AddTypeReference(a)
		}
		v.AddStringColored(util.TEXT_BLUE, ">")
	}
	return v
}

func (v *ASTStringer) AddGenericArguments(args []*TypeReference) *ASTStringer {
	if len(args) == 0 {
		return v
	}

	str := "<"
	for i, arg := range args {
		str += arg.String()

		if i < len(args)-1 {
			str += ", "
		}
	}
	str += ">"
	v.AddStringColored(util.TEXT_GREEN, str)
	return v
}

func (v *ASTStringer) Finish() string {
	v.buf.WriteRune(')')
	return v.buf.String()
}
