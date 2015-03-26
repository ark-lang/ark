This document specifies the grammar for the Alloy programming language. It's still a work in progress,
some of the language may be missing, or some of the following may be incorrect/invalid or outdated.

	digit = { "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9" } .
	letter = "A" | "a" | ... "Z" | "z" | "_" .
	identifier = letter { letter | digit } .
	sign = "+" | "-" .
	raw_value = letter .
	escaped_char = "\" ( "a" | "b" | "f" | "n" | "r" | "t" | "v" | "\" | "'" | """ ) .
	binaryOp = logOp | relOp | addOp | mulOp .
	logOp = "||" | "&&"
	relOp = "==" | "!=" | "<" | "<=" | ">" | ">=" .
	addOp = "+" | "-" | "|" | "^" .
	mulOp = "*" | "/" | "%" | "<<" | ">>" | "&" .
	unaryOp = "+" | "-" | "!" | "^" | "<" | ">" | "*" | "&" .

	NumberLiteral = [sign] digit [ "." { digit } ]	
	StringLiteral = """ { letter } """ . 
	CharacterLiteral = "'"  ( raw_value | escaped_char ) "'" .
	Literal = NumberLiteral | StringLiteral | CharacterLiteral .
	
	Type = TypeName | ArrayType | PointerType .
	
	TypeName = identifier .
	
	ArrayType = Type "[" [ Expression ] "]" .
	
	StructDecl = "struct" "{" [ FieldList ] "}" .
	FieldList = FieldDecl { ";" FieldDecl } .
	FieldDecl = Type IdentifierList .

	FunctionDecl = "fn" FunctionSignature ( ";" | Block ) .
	FunctionSignature = [ Receiver ] identifier Parameters ":" Type .
	Receiver = "(" Type identifier ")"
	Parameter = "(" [ parameterList ] ")" .
	ParameterList = ParameterSection { "," ParameterSection } .
	ParameterSection = Type IdentifierList .

	PointerType = Type "^" .
	
	Block = ( "{" [ StatementList ";" ] "}" | "->" Statement ) .
	IfStat = if Expression Block [ "else" Statement ] .
	StatementList = { Statement ";" } .

	MatchStat = "match" Expression "{" { MatchClause "," } "}" . 
	MatchClause = Expression Block [ LeaveStat ] . 
	
	ForStat = "for" Type identifier ":" "(" IdentList ")" Block .
	
	Declaration = VarDecl | FunctionDecl | StructDecl.
	VarDecl = [ "mut" ] Type identifier [ "=" Expression ] ";" .
	
	Expression = BinaryExpr | UnaryExpr | PrimaryExpr .
	BinaryExpr = Expression binaryOp Expression .
	UnaryExpr = unaryOp Expression .

	PrimaryExpr =
		identifier 
		| Literal 
		| "(" Expression ")" 
		| Call 
		| Expression "[" Expression [ ":" Expression ] "]" 
		| Expression "." identifier.
		
	Call = Expression "(" [ ExpressionList ] ")" .
	
	Statement = ( StructuredStat | UnstructuredStat ) .
	StructuredStat = Block | IfStat | MatchStat | ForStat .
	UnstructuredStat = Declaration | LeaveStat | IncDecStat | Assignment.
	LeaveStat = ReturnStat | BreakStat | ContinueStat .
	Assignment = PrimaryExpr "=" Expression .
	ReturnStat = "return" [ Expression ] ";" .
	BreakStat = "break" ";"
	ContinueStat = "continue" ";"
	IncDecStat = Expression ( "++" | "--" ) .
	
	
	
	