This document specifies the grammar for the Alloy programming language. It's still a work in progress.

	digit = { "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9" } .
	letter = "A" | "a" | ... "Z" | "z" | "_" .
	identifier = letter { letter | digit } .
	sign = "+" | "-" .
	raw_value = letter .
	escaped_char = "\" ( "a" | "b" | "f" | "n" | "r" | "t" | "v" | "\" | "'" | """ ) .

	NumberLiteral = [sign] digit [ "." { digit } ]	
	StringLiteral = """ { letter } """ . 
	CharacterLiteral = "'"  ( raw_value | escaped_char ) "'" .
	Literal = NumberLiteral | StringLiteral | CharacterLiteral .
	
	Type = TypeName | ArrayType | StructType | FunctionType | PointerType .
	
	TypeName = identifier .
	
	ArrayType = "[" [ Expression ] "]" Type .
	
	StructType = "struct" "{" [ FieldList ] "}" .
	FieldList = FieldDecl { ";" FieldDecl } .
	FieldDecl = IdentifierList Type .

	FunctionDecl = "fn" FunctionSignature ( ";" | Block ) .
	FunctionSignature = [ Receiver ] identifier Parameters ":" Type .
	Receiver = "(" Type identifier ")"
	Parameter = "(" [ parameterList ] ")" .
	ParameterList = ParameterSection { "," ParameterSection } .
	ParameterSection = Type IdentifierList .

	PointerType = "^" Type .
	
	Block = ( "{" [ StatementList ";" ] "}" | "->" Statement ) .
	IfStat = if Expression Block [ "else" Statement ] .
	StatementList = { Statement ";" } .

	// todo	
	MatchStat = "match" Expression "{" { MatchClause } "}" . 
	MatchClause = 
	
	// hmm
	ForStat = "for" Type identifier ":" "(" IdentList ")" Block .
	
	Declaration = VarDecl | FunctionDecl .
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
		| Expression "." identifier | Expression "." "(" Type ")" .
		
	Call = Expression "(" [ ExpressionList ] ")" .
	
	binaryOp = logOp | relOp | addOp | mulOp .
	logOp = "||" | "&&"
	relOp = "==" | "!=" | "<" | "<=" | ">" | ">=" .
	addOp = "+" | "-" | "|" | "^" .
	mulOp = "*" | "/" | "%" | "<<" | ">>" | "&" .
	unaryOp = "+" | "-" | "!" | "^" | "<" | ">" | "*" | "&" .
	
	Statement = ( StructuredStat | UnstructuredStat ) .
	StructuredStat = Block | IfStat | MatchStat | ForStat .
	UnstructuredStat = Declaration | ReturnStat | BreakStat | ContinueStat | IncDecStat | Assignment.
	Assignment = PrimaryExpr "=" Expression .
	ReturnStat = "return" [ Expression ] ";" .
	BreakStat = "break" ";"
	ContinueStat = "continue" ";"
	IncDecStat = Expression ( "++" | "--" ) .
	
	
	
	