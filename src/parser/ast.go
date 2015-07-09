package parser

import (
	"fmt"
	"strconv"

	"github.com/ark-lang/ark/src/util"
)

//
type Locatable interface {
	Pos() (filename string, line, char int)
	setPos(filename string, line, char int)
}

type Node interface {
	String() string
	NodeName() string
	resolve(*Resolver, *Scope)
	analyze(*SemanticAnalyzer)
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

type AccessExpr interface {
	Expr
	Mutable() bool
}

type Decl interface {
	Node
	declNode()
}

type Documentable interface {
	DocComments() []*DocComment
}

// an implementation of Locatable that is used for Nodes
type nodePos struct {
	filename               string
	lineNumber, charNumber int
}

func (v nodePos) Pos() (filename string, line, char int) {
	return v.filename, v.lineNumber, v.charNumber
}

func (v *nodePos) setPos(filename string, line, char int) {
	v.filename = filename
	v.lineNumber = line
	v.charNumber = char
}

type DocComment struct {
	Contents           string
	StartLine, EndLine int
}

type Variable struct {
	Type         Type
	Name         string
	Mutable      bool
	Attrs        AttrGroup
	scope        *Scope
	Uses         int
	ParentStruct *StructType
}

func (v *Variable) String() string {
	result := "(" + util.Blue("Variable") + ": "
	if v.Mutable {
		result += util.Green("[mutable] ")
	}
	for _, attr := range v.Attrs {
		result += attr.String() + " "
	}
	return result + v.Name + util.Magenta(" <"+v.MangledName(MANGLE_ARK_UNSTABLE)+"> ") + util.Green(v.Type.TypeName()) + ")"
}

func (v *Variable) Scope() *Scope {
	return v.scope
}

type Function struct {
	Name       string
	Parameters []*VariableDecl
	ReturnType Type
	Mutable    bool
	IsVariadic bool
	Attrs      AttrGroup
	Body       *Block
	Uses       int
	scope      *Scope
}

func (v *Function) Scope() *Scope {
	return v.scope
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
	if v.ReturnType != nil {
		result += ": " + util.Green(v.ReturnType.TypeName()) + " "
	}

	if v.Body != nil {
		result += v.Body.String()
	}
	return result + util.Magenta(" <"+v.MangledName(MANGLE_ARK_UNSTABLE)+">") + ")"
}

//
// Nodes
//

type Block struct {
	nodePos
	Nodes         []Node
	scope         *Scope
	IsTerminating bool
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
	docs       []*DocComment
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

func (v *VariableDecl) DocComments() []*DocComment {
	return v.docs
}

// ModuleDecl

type ModuleDecl struct {
	nodePos
	Module *Module
	docs   []*DocComment
}

func (v *ModuleDecl) declNode() {}

func (v *ModuleDecl) String() string {
	result := "(" + util.Blue(v.Module.Name) + " (in " + util.Green(v.Module.Path) + "): \n"
	// todo interfaces for this shit
	for _, function := range v.Module.Functions {
		result += "\t" + function.String() + "\n"
	}
	for _, variable := range v.Module.Variables {
		result += "\t" + variable.String() + "\n"
	}
	result += ")"
	return result
}

func (v *ModuleDecl) NodeName() string {
	return "mod decl"
}

func (v *ModuleDecl) DocComments() []*DocComment {
	return nil // TODO
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

func (v *StructDecl) DocComments() []*DocComment {
	return nil // TODO
}

// TraitDecl

type TraitDecl struct {
	nodePos
	Trait *TraitType
}

func (v *TraitDecl) declNode() {}

func (v *TraitDecl) String() string {
	return "(" + util.Blue("TraitDecl") + ": " + v.Trait.String() + ")"
}

func (v *TraitDecl) NodeName() string {
	return "trait declaration"
}

func (v *TraitDecl) DocComments() []*DocComment {
	return nil // TODO
}

// EnumDecl
type EnumVal struct {
	Name  string
	Value Expr
}

func (v *EnumVal) String() string {
	if v.Value != nil {
		return "(" + util.Blue("EnumVal") + ": " + v.Name + " = " + v.Value.String() + ")"
	}
	return "(" + util.Blue("EnumVal") + ": " + v.Name + ")"
}

type EnumDecl struct {
	nodePos
	Name        string
	Body        []*EnumVal
	IsAnonymous bool
}

func (v *EnumDecl) declNode() {}

func (v *EnumDecl) String() string {
	result := "\n"
	for _, val := range v.Body {
		result += "\t" + val.String() + "\n"
	}
	return "(" + util.Blue("EnumDecl") + ": " + result + ")"
}

func (v *EnumDecl) NodeName() string {
	return "enum declaration"
}

func (v *EnumDecl) DocComments() []*DocComment {
	return nil // TODO
}

// ImplDecl

type ImplDecl struct {
	nodePos
	StructName string
	TraitName  string
	Functions  []*FunctionDecl
}

func (v *ImplDecl) declNode() {}

func (v *ImplDecl) String() string {
	var result string
	for _, decl := range v.Functions {
		result += "\t" + decl.String() + "\n"
	}
	return "(" + util.Blue("ImplDecl") + ": " + result + ")"
}

func (v *ImplDecl) NodeName() string {
	return "impl declaration"
}

func (v *ImplDecl) DocComments() []*DocComment {
	return nil // TODO
}

// UseDecl

type UseDecl struct {
	nodePos
	ModuleName string
	Scope      *Scope
}

func (v *UseDecl) declNode() {}

func (v *UseDecl) String() string {
	return "(" + util.Blue("UseDecl") + ": " + v.ModuleName + ")"
}

func (v *UseDecl) NodeName() string {
	return "use declaration"
}

// FunctionDecl

type FunctionDecl struct {
	nodePos
	Function  *Function
	Prototype bool
	docs      []*DocComment
}

func (v *FunctionDecl) declNode() {}

func (v *FunctionDecl) String() string {
	return "(" + util.Blue("FunctionDecl") + ": " + v.Function.String() + ")"
}

func (v *FunctionDecl) NodeName() string {
	return "function declaration"
}

func (v *FunctionDecl) DocComments() []*DocComment {
	return v.docs
}

// DirectiveDecl

type DirectiveDecl struct {
	nodePos
	Name     string
	Argument *StringLiteral
}

func (v *DirectiveDecl) declNode() {}

func (v *DirectiveDecl) String() string {
	return "(" + util.Blue("DirectiveDecl") + ": #" + v.Name + " " + v.Argument.String() + ")"
}

func (v *DirectiveDecl) NodeName() string {
	return "function declaration"
}

/**
 * Statements
 */

// BlockStat

type BlockStat struct {
	nodePos
	Block *Block
}

func (v *BlockStat) statNode() {}

func (v *BlockStat) String() string {
	return "(" + util.Blue("BlockStat") + ": " +
		v.Block.String() + ")"
}

func (v *BlockStat) NodeName() string {
	return "block statement"
}

// ReturnStat

type ReturnStat struct {
	nodePos
	Value Expr
}

func (v *ReturnStat) statNode() {}

func (v *ReturnStat) String() string {
	ret := "(" + util.Blue("ReturnStat") + ": "
	if v.Value == nil {
		ret += "void"
	} else {
		ret += v.Value.String()
	}
	return ret + ")"
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

// DeferStat

type DeferStat struct {
	nodePos
	Call *CallExpr
}

func (v *DeferStat) statNode() {}

func (v *DeferStat) String() string {
	return "(" + util.Blue("DeferStat") + ": " +
		v.Call.String() + ")"
}

func (v *DeferStat) NodeName() string {
	return "call statement"
}

// AssignStat

type AssignStat struct {
	nodePos
	Deref      *DerefExpr // one of these should be nil, not neither or both. felix: what even for x = 5?
	Access     AccessExpr
	Assignment Expr
}

func (v *AssignStat) statNode() {}

func (v *AssignStat) String() string {
	result := "(" + util.Blue("AssignStat") + ": "
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

// IfStat

type IfStat struct {
	nodePos
	Exprs  []Expr
	Bodies []*Block
	Else   *Block // can be nil
}

func (v *IfStat) statNode() {}

func (v *IfStat) String() string {
	result := "(" + util.Blue("IfStat") + ": "
	for i, expr := range v.Exprs {
		result += expr.String() + " "
		result += v.Bodies[i].String()
	}
	if v.Else != nil {
		result += v.Else.String()
	}
	return result + ")"
}

func (v *IfStat) NodeName() string {
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

func (v *LoopStat) statNode() {}

func (v *LoopStat) String() string {
	result := "(" + util.Blue("LoopStat") + ": "

	switch v.LoopType {
	case LOOP_TYPE_INFINITE:
	case LOOP_TYPE_CONDITIONAL:
		result += v.Condition.String() + " "
	default:
		panic("invalid loop type")
	}

	result += v.Body.String()

	return result + ")"
}

func (v *LoopStat) NodeName() string {
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

func (v *MatchStat) statNode() {}

func (v *MatchStat) String() string {
	result := "(" + util.Blue("MatchStat") + ": " + v.Target.String() + ":\n"

	for pattern, stmt := range v.Branches {
		result += "\t" + pattern.String() + " -> " + stmt.String() + "\n"
	}

	return result + ")"
}

func (v *MatchStat) NodeName() string {
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

func (v *RuneLiteral) exprNode() {}

func (v *RuneLiteral) String() string {
	return fmt.Sprintf("(" + util.Blue("RuneLiteral") + ": " + colorizeEscapedString(EscapeString(string(v.Value))) + " " + util.Green(v.GetType().TypeName()) + ")")
}

func (v *RuneLiteral) GetType() Type {
	return PRIMITIVE_rune
}

func (v *RuneLiteral) NodeName() string {
	return "rune literal"
}

// NumericLiteral
type NumericLiteral struct {
	nodePos
	IntValue   uint64
	FloatValue float64
	IsFloat    bool
	Type       Type
	typeHint   Type
}

func (v *NumericLiteral) exprNode() {}

func (v *NumericLiteral) String() string {
	if v.IsFloat {
		return fmt.Sprintf("("+util.Blue("NumericLiteral")+": "+util.Yellow("%d")+" "+util.Green(v.GetType().TypeName())+")", v.FloatValue)
	} else {
		return fmt.Sprintf("("+util.Blue("NumericLiteral")+": "+util.Yellow("%d")+" "+util.Green(v.GetType().TypeName())+")", v.IntValue)
	}
}

func (v *NumericLiteral) GetType() Type {
	return v.Type
}

func (v *NumericLiteral) NodeName() string {
	return "numeric literal"
}

func (v *NumericLiteral) AsFloat() float64 {
	if v.IsFloat {
		return v.FloatValue
	} else {
		return float64(v.IntValue)
	}
}

func (v *NumericLiteral) AsInt() uint64 {
	if v.IsFloat {
		panic("downcasting floating point value to int")
	}

	return v.IntValue
}

// StringLiteral

type StringLiteral struct {
	nodePos
	Value  string
	StrLen int
}

func (v *StringLiteral) exprNode() {}

func (v *StringLiteral) String() string {
	return "(" + util.Blue("StringLiteral") + ": " + colorizeEscapedString((EscapeString(v.Value))) + " " + util.Green(v.GetType().TypeName()) + ")"
}

func (v *StringLiteral) GetType() Type {
	return PRIMITIVE_str
}

func (v *StringLiteral) NodeName() string {
	return "string literal"
}

// BoolLiteral

type BoolLiteral struct {
	nodePos
	Value bool
}

func (v *BoolLiteral) exprNode() {}

func (v *BoolLiteral) String() string {
	res := "(" + util.Blue("BoolLiteral") + ": "
	if v.Value {
		res += util.Yellow("true")
	} else {
		res += util.Yellow("false")
	}
	return res + ")"
}

func (v *BoolLiteral) GetType() Type {
	return PRIMITIVE_bool
}

func (v *BoolLiteral) NodeName() string {
	return "boolean literal"
}

// TupleLiteral

type TupleLiteral struct {
	nodePos
	Members []Expr
	Type    Type
}

func (v *TupleLiteral) exprNode() {}

func (v *TupleLiteral) String() string {
	return "tuple"
}

func (v *TupleLiteral) GetType() Type {
	return v.Type
}

func (v *TupleLiteral) NodeName() string {
	return "tuple literal"
}

// ArrayLiteral

type ArrayLiteral struct {
	nodePos
	Members []Expr
	Type    Type
}

func (v *ArrayLiteral) exprNode() {}

func (v *ArrayLiteral) String() string {
	res := "(" + util.Blue("ArrayLiteral") + ":"
	for _, mem := range v.Members {
		res += " " + mem.String()
	}
	return res + ")"
}

func (v *ArrayLiteral) GetType() Type {
	return v.Type
}

func (v *ArrayLiteral) NodeName() string {
	return "array literal"
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
	Function       *Function
	functionSource Expr
	Arguments      []Expr
}

func (v *CallExpr) exprNode() {}

func (v *CallExpr) String() string {
	result := "(" + util.Blue("CallExpr") + ": " + v.Function.Name
	for _, arg := range v.Arguments {
		result += " " + arg.String()
	}
	if v.GetType() != nil {
		result += " " + util.Green(v.GetType().TypeName())
	}
	return result + ")"
}

func (v *CallExpr) GetType() Type {
	return v.Function.ReturnType
}

func (v *CallExpr) NodeName() string {
	return "call expression"
}

// VariableAccessExpr
type VariableAccessExpr struct {
	nodePos
	Name unresolvedName

	Variable *Variable
}

func (v *VariableAccessExpr) exprNode() {}

func (v *VariableAccessExpr) String() string {
	result := "(" + util.Blue("VariableAccessExpr") + ": name"
	result += v.Name.String()
	return result + ")"
}

func (v *VariableAccessExpr) GetType() Type {
	return v.Variable.Type
}

func (v *VariableAccessExpr) NodeName() string {
	return "variable access expression"
}

func (v *VariableAccessExpr) Mutable() bool {
	return v.Variable.Mutable
}

// StructAccessExpr
type StructAccessExpr struct {
	nodePos
	Struct AccessExpr
	Member string

	Variable *Variable
}

func (v *StructAccessExpr) exprNode() {}

func (v *StructAccessExpr) String() string {
	result := "(" + util.Blue("StructAccessExpr") + ": struct"
	result += v.Struct.String() + ", member "
	result += v.Member
	return result + ")"
}

func (v *StructAccessExpr) GetType() Type {
	return v.Variable.Type
}

func (v *StructAccessExpr) NodeName() string {
	return "struct access expression"
}

func (v *StructAccessExpr) Mutable() bool {
	return true
}

// ArrayAccessExpr

type ArrayAccessExpr struct {
	nodePos
	Array     AccessExpr
	Subscript Expr
}

func (v *ArrayAccessExpr) exprNode() {}

func (v *ArrayAccessExpr) String() string {
	result := "(" + util.Blue("ArrayAccessExpr") + ": array"
	result += v.Array.String() + ", index "
	result += v.Subscript.String()
	return result + ")"
}

func (v *ArrayAccessExpr) GetType() Type {
	return v.Array.GetType().(ArrayType).MemberType
}

func (v *ArrayAccessExpr) NodeName() string {
	return "array access expression"
}

func (v *ArrayAccessExpr) Mutable() bool {
	return v.Array.Mutable()
}

// TupleAccessExpr

type TupleAccessExpr struct {
	nodePos
	Tuple AccessExpr
	Index uint64
}

func (v *TupleAccessExpr) exprNode() {}

func (v *TupleAccessExpr) String() string {
	result := "(" + util.Blue("TupleAccessExpr") + ": tuple"
	result += v.Tuple.String() + ", index "
	result += strconv.FormatUint(v.Index, 10)
	return result + ")"
}

func (v *TupleAccessExpr) GetType() Type {
	return v.Tuple.GetType().(*TupleType).Members[v.Index]
}

func (v *TupleAccessExpr) NodeName() string {
	return "tuple access expression"
}

func (v *TupleAccessExpr) Mutable() bool {
	return v.Tuple.Mutable()
}

// AddressOfExpr

type AddressOfExpr struct {
	nodePos
	Access Expr
}

func (v *AddressOfExpr) exprNode() {}

func (v *AddressOfExpr) String() string {
	return "(" + util.Blue("AddressOfExpr") + ": " + v.Access.String() + " " + util.Green(v.GetType().TypeName()) + ")"
}

func (v *AddressOfExpr) GetType() Type {
	return pointerTo(v.Access.GetType())
}

func (v *AddressOfExpr) NodeName() string {
	return "address-of expression"
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

// SizeofExpr

type SizeofExpr struct {
	nodePos
	// oneither Expr or Type is nil, not neither or both

	Expr Expr

	Type Type
}

func (v *SizeofExpr) exprNode() {}

func (v *SizeofExpr) String() string {
	ret := "(" + util.Blue("SizeofExpr") + ": "
	if v.Expr != nil {
		ret += v.Expr.String()
	} else {
		ret += v.Type.TypeName()
	}
	return ret + ")"
}

func (v *SizeofExpr) GetType() Type {
	return PRIMITIVE_uint
}

func (v *SizeofExpr) NodeName() string {
	return "sizeof expression"
}

// DefaultMatchBranch

type DefaultMatchBranch struct {
	nodePos
}

func (v *DefaultMatchBranch) exprNode() {}

func (v *DefaultMatchBranch) String() string {
	return "(" + util.Blue("DefaultMatchBranch") + ")"
}

func (v *DefaultMatchBranch) GetType() Type {
	return nil
}

func (v *DefaultMatchBranch) NodeName() string {
	return "default match branch"
}
