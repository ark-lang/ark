package parser

import (
	"bytes"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ark-lang/ark/src/lexer"
	"github.com/ark-lang/ark/src/util"
)

type Locatable interface {
	Pos() lexer.Position
	setPos(pos lexer.Position)
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
	GetType() Type
	SetType(Type)
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

type Documentable interface {
	DocComments() []*DocComment
}

// an implementation of Locatable that is used for Nodes
type nodePos struct {
	pos lexer.Position
}

func (v nodePos) Pos() lexer.Position {
	return v.pos
}

func (v *nodePos) setPos(pos lexer.Position) {
	v.pos = pos
}

type DocComment struct {
	Contents string
	Where    lexer.Span
}

type Variable struct {
	Type         Type
	Name         string
	Mutable      bool
	Attrs        AttrGroup
	FromStruct   bool
	ParentStruct StructType
	ParentModule *Module
	IsParameter  bool
	IsArgument   bool
}

func (v Variable) String() string {
	s := NewASTStringer("Variable")
	if v.Mutable {
		s.AddStringColored(util.TEXT_GREEN, " [mutable]")
	}
	s.AddAttrs(v.Attrs)
	s.AddString(v.Name)
	s.AddMangled(v)
	s.AddType(v.Type)
	return s.Finish()
}

func (v Variable) GetType() Type {
	return v.Type
}

// Note that for static methods, ``
type Function struct {
	Name       string
	Type       FunctionType
	Parameters []*VariableDecl
	Body       *Block

	ParentModule *Module

	Receiver *VariableDecl // non-nil if non-static method

	StaticReceiverType Type // non-nil if static
}

func (v Function) String() string {
	s := NewASTStringer("Function")
	s.AddAttrs(v.Type.Attrs())
	s.AddString(v.Name)
	for _, par := range v.Parameters {
		s.Add(par)
	}
	if v.Type.Return != nil {
		s.AddString(":")
		s.AddType(v.Type.Return)
	}
	s.Add(v.Body)
	s.AddMangled(v)
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
	docs       []*DocComment
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

func (v VariableDecl) DocComments() []*DocComment {
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

func (v TypeDecl) DocComments() []*DocComment {
	return nil // TODO
}

// FunctionDecl

type FunctionDecl struct {
	nodePos
	PublicHandler
	Function  *Function
	Prototype bool
	docs      []*DocComment
}

func (_ FunctionDecl) declNode() {}

func (v FunctionDecl) String() string {
	return NewASTStringer("FunctionDecl").Add(v.Function).Finish()
}

func (_ FunctionDecl) NodeName() string {
	return "function declaration"
}

func (v FunctionDecl) DocComments() []*DocComment {
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
	Operator   BinOpType
	Assignment Expr
}

func (_ BinopAssignStat) statNode() {}

func (v BinopAssignStat) String() string {
	return NewASTStringer("BinopAssignStat").Add(v.Access).Add(v.Operator).AddString(" =").Add(v.Assignment).Finish()
}

func (_ BinopAssignStat) NodeName() string {
	return "binop assignment statement"
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

func newMatch() *MatchStat {
	return &MatchStat{Branches: make(map[Expr]Node)}
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
	).AddType(v.GetType()).Finish()
}

func (v RuneLiteral) GetType() Type {
	return PRIMITIVE_rune
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
	Type          Type
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
	s.AddType(v.GetType())
	return s.Finish()
}

func (v NumericLiteral) GetType() Type {
	return v.Type
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
	Type      Type
}

func (_ StringLiteral) exprNode() {}

func (v StringLiteral) String() string {
	return NewASTStringer("StringLiteral").AddString(colorizeEscapedString(EscapeString(v.Value))).AddType(v.GetType()).Finish()
}

func (v StringLiteral) GetType() Type {
	return v.Type
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

func (v BoolLiteral) GetType() Type {
	return PRIMITIVE_bool
}

func (_ BoolLiteral) NodeName() string {
	return "boolean literal"
}

// TupleLiteral

type TupleLiteral struct {
	nodePos
	Members []Expr
	Type    Type
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

func (v TupleLiteral) GetType() Type {
	return v.Type
}

func (_ TupleLiteral) NodeName() string {
	return "tuple literal"
}

// CompositeLiteral

type CompositeLiteral struct {
	nodePos
	Type   Type
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
	s.AddType(v.Type)
	return s.Finish()
}

func (v CompositeLiteral) GetType() Type {
	return v.Type
}

func (_ CompositeLiteral) NodeName() string {
	return "composite literal"
}

// EnumLiteral

type EnumLiteral struct {
	nodePos
	Type   Type
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

func (v EnumLiteral) GetType() Type {
	return v.Type
}

func (_ EnumLiteral) NodeName() string {
	return "enum literal"
}

// BinaryExpr

type BinaryExpr struct {
	nodePos
	Lhand, Rhand Expr
	Op           BinOpType
	Type         Type
}

func (_ BinaryExpr) exprNode() {}

func (v BinaryExpr) String() string {
	return NewASTStringer("BinaryExpr").Add(v.Op).Add(v.Lhand).Add(v.Rhand).Finish()
}

func (v BinaryExpr) GetType() Type {
	return v.Type
}

func (_ BinaryExpr) NodeName() string {
	return "binary expression"
}

// UnaryExpr

type UnaryExpr struct {
	nodePos
	Expr Expr
	Op   UnOpType
	Type Type
}

func (_ UnaryExpr) exprNode() {}

func (v UnaryExpr) String() string {
	return NewASTStringer("UnaryExpr").Add(v.Op).Add(v.Expr).Finish()
}

func (v UnaryExpr) GetType() Type {
	return v.Type
}

func (_ UnaryExpr) NodeName() string {
	return "unary expression"
}

// CastExpr

type CastExpr struct {
	nodePos
	Expr Expr
	Type Type
}

func (_ CastExpr) exprNode() {}

func (v CastExpr) String() string {
	return NewASTStringer("CastExpr").Add(v.Expr).AddType(v.GetType()).Finish()
}

func (v CastExpr) GetType() Type {
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

	GenericParameters []Type
}

func (_ CallExpr) exprNode() {}

func (v CallExpr) String() string {
	s := NewASTStringer("CallExpr")
	s.Add(v.Function)
	for _, arg := range v.Arguments {
		s.Add(arg)
	}
	s.AddType(v.GetType())
	return s.Finish()
}

func (v CallExpr) GetType() Type {
	if v.Function != nil {
		fnType := v.Function.GetType()
		if fnType != nil {
			return fnType.(FunctionType).Return
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

	Function *Function

	GenericParameters []Type
}

func (_ FunctionAccessExpr) exprNode() {}

func (v FunctionAccessExpr) String() string {
	return NewASTStringer("FunctionAccessExpr").AddString(v.Function.Name).Finish()
}

func (v FunctionAccessExpr) GetType() Type {
	if v.Function != nil {
		return v.Function.Type
	}
	return nil
}

func (_ FunctionAccessExpr) NodeName() string {
	return "function access expression"
}

// VariableAccessExpr
type VariableAccessExpr struct {
	nodePos
	Name UnresolvedName

	Variable *Variable

	GenericParameters []Type
}

func (_ VariableAccessExpr) exprNode() {}

func (v VariableAccessExpr) String() string {
	return NewASTStringer("VariableAccessExpr").Add(v.Name).Finish()
}

func (v VariableAccessExpr) GetType() Type {
	if v.Variable != nil {
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
}

func (_ StructAccessExpr) exprNode() {}

func (v StructAccessExpr) String() string {
	s := NewASTStringer("StructAccessExpr")
	s.AddString("struct").Add(v.Struct)
	s.AddString("member").AddString(v.Member)
	return s.Finish()
}

func (v StructAccessExpr) GetType() Type {
	stype := v.Struct.GetType()

	if typ, ok := TypeWithoutPointers(stype).(*NamedType); ok {
		fn := typ.GetMethod(v.Member)
		if fn != nil {
			return fn.Type
		}
	}

	if stype == nil {
		return nil
	} else if pt, ok := stype.(PointerType); ok {
		stype = pt.Addressee
	}

	if stype == nil {
		return nil
	} else if st, ok := stype.ActualType().(StructType); ok {
		mem := st.GetMember(v.Member)
		if mem != nil {
			return mem.Type
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

func (v ArrayAccessExpr) GetType() Type {
	if v.Array.GetType() != nil {
		return v.Array.GetType().ActualType().(ArrayType).MemberType
	}
	return nil
}

func (_ ArrayAccessExpr) NodeName() string {
	return "array access expression"
}

func (v ArrayAccessExpr) Mutable() bool {
	return v.Array.Mutable()
}

// TupleAccessExpr

type TupleAccessExpr struct {
	nodePos
	Tuple AccessExpr
	Index uint64
}

func (_ TupleAccessExpr) exprNode() {}

func (v TupleAccessExpr) String() string {
	s := NewASTStringer("TupleAccessExpr")
	s.AddString("tuple").Add(v.Tuple)
	s.AddString("index").AddString(strconv.FormatUint(v.Index, 10))
	return s.Finish()
}

func (v TupleAccessExpr) GetType() Type {
	if v.Tuple.GetType() != nil {
		return v.Tuple.GetType().ActualType().(TupleType).Members[v.Index]
	}
	return nil
}

func (_ TupleAccessExpr) NodeName() string {
	return "tuple access expression"
}

func (v TupleAccessExpr) Mutable() bool {
	return v.Tuple.Mutable()
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

func (v DerefAccessExpr) GetType() Type {
	return getAdressee(v.Expr.GetType())
}

func (_ DerefAccessExpr) NodeName() string {
	return "dereference access expression"
}

func (v DerefAccessExpr) Mutable() bool {
	access, ok := v.Expr.(AccessExpr)
	if ok {
		return access.Mutable()
	} else {
		return true
	}
}

func getAdressee(t Type) Type {
	switch t := t.(type) {
	case PointerType:
		return t.Addressee
	case MutableReferenceType:
		return t.Referrer
	case ConstantReferenceType:
		return t.Referrer
	}
	return nil
}

// AddressOfExpr

type AddressOfExpr struct {
	nodePos
	Mutable  bool
	Access   Expr
	TypeHint Type
}

func (_ AddressOfExpr) exprNode() {}

func (v AddressOfExpr) String() string {
	return NewASTStringer("AddressOfExpr").Add(v.Access).AddType(v.GetType()).Finish()
}

func (v AddressOfExpr) GetType() Type {
	if v.Access.GetType() != nil {
		if v.Mutable {
			return mutableReferenceTo(v.Access.GetType())
		} else {
			return constantReferenceTo(v.Access.GetType())
		}
	}
	return nil
}

func (_ AddressOfExpr) NodeName() string {
	return "address-of expression"
}

type LambdaExpr struct {
	nodePos

	Function *Function
}

func (_ LambdaExpr) exprNode() {}

func (v LambdaExpr) String() string {
	return NewASTStringer("LambdaExpr").Add(v.Function).Finish()
}

func (v LambdaExpr) GetType() Type {
	return v.Function.Type
}

func (v LambdaExpr) NodeName() string {
	return "lambda expr"
}

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

func (v ArrayLenExpr) GetType() Type {
	return PRIMITIVE_uint
}

func (_ ArrayLenExpr) NodeName() string {
	return "array length expr"
}

// SizeofExpr

type SizeofExpr struct {
	nodePos
	// oneither Expr or Type is nil, not neither or both

	Expr Expr

	Type Type
}

func (_ SizeofExpr) exprNode() {}

func (v SizeofExpr) String() string {
	s := NewASTStringer("SizeofExpr")
	if v.Expr != nil {
		s.Add(v.Expr)
	} else {
		s.AddType(v.Type)
	}
	return s.Finish()
}

func (v SizeofExpr) GetType() Type {
	return PRIMITIVE_uint
}

func (_ SizeofExpr) NodeName() string {
	return "sizeof expression"
}

// DefaultMatchBranch

type DefaultMatchBranch struct {
	nodePos
}

func (_ DefaultMatchBranch) exprNode() {}

func (v DefaultMatchBranch) String() string {
	return NewASTStringer("DefaultMatchBranch").Finish()
}

func (v DefaultMatchBranch) GetType() Type {
	return nil
}

func (_ DefaultMatchBranch) NodeName() string {
	return "default match branch"
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

func (v *ASTStringer) AddAttrs(attrs AttrGroup) *ASTStringer {
	for _, attr := range attrs {
		v.Add(attr)
	}
	return v
}

func (v *ASTStringer) AddMangled(mangled Mangled) *ASTStringer {
	v.AddStringColored(util.TEXT_MAGENTA, fmt.Sprintf(" <%s>", mangled.MangledName(MANGLE_ARK_UNSTABLE)))
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

func (v *ASTStringer) Finish() string {
	v.buf.WriteRune(')')
	return v.buf.String()
}
