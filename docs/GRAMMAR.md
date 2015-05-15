This document specifies the grammar for the Alloy programming language. It's still a work in progress,
some of the language may be missing, or some of the following may be incorrect/invalid or out-dated.
The grammar pretty much maps straight onto the parsers source.

	digit = { "0" ... "9" }
	letter = "a" ... "Z" | "_";
	hex_digit = "0" ... "9" | "A" ... "F" | "a" ... "f"
	
	identifier = letter { letter | digit };
	
	sign = "+" | "-";
	escaped_char = "\" ( "a" | "b" | "f" | "n" | "r" | "t" | "v" | "\" | "'" | """ );
	binaryOp = logOp | relOp | addOp | mulOp;
	logOp = "||" | "&&"
	relOp = "==" | "!=" | "<" | "<=" | ">" | ">=";
	addOp = "+" | "-" | "|" | "^";
	mulOp = "*" | "/" | "%" | "<<" | ">>" | "&";
	unaryOp = "+" | "-" | "!" | "^" | "<" | ">" | "*" | "&";
	hex_literal = "0" ( "x" | "X" ) hex_digit { hex_digit };
	
	Literal = NumberLiteral | StringLiteral | CharacterLiteral;
	NumberLiteral = [sign] digit [ "." { digit } ]	
	StringLiteral = """ { letter } """; 
	CharacterLiteral = "'"  ( letter | escaped_char ) "'";
	
	IdentifierList = identifier [ { "," identifier } ]
	ExpressionList = Expression { "," Expression }
	
	Type = TypeName | TypeLit.
	TypeLit = ArrayType | PointerType;
	TypeName = identifier;
	PointerType = "^" BaseType;
	ArrayType = "[" [ Expression ] "]" Type.
	
	FunctionSignature = identifier Parameters ":" [ "mut" ] Type;
	Parameters = "(" [ parameterList ] ")";
	ParameterList = ParameterSection { "," ParameterSection };
	ParameterSection = [ "mut" ] Type identifier;

	StructDecl = "struct" identifier "{" [ FieldDeclList ] "}";
	FieldDeclList = FieldDecl { FieldDecl };
	FieldDecl = [ "mut" ] Type IdentifierList ";";
	FunctionDecl = FunctionSignature ( ";" | Block );

	Impl = "impl" identifier [ "as" identifier ] "{" { FunctionDecl } "}"
	Block = ( "{" [ StatementList ] "}" | "->" Statement );
	IfStat = if Expression Block [ "else" Statement ];
	StatementList = { Statement ";" };
	MatchStat = "match" Expression "{" { MatchClause "," } "}"; 
	MatchClause = Expression Block ";"; 

	ForInfiniteLoop = "for" Block;
	ForWhileLoop = "for" PrimaryExpr Block;
	ForStepLoop = "for" PrimaryExpr "," PrimaryExpr Block;
	ForStat = ForInfiniteLoop | ForWhileLoop | ForStepLoop;

	Declaration = VarDecl | FunctionDecl | StructDecl.
	VarDecl = [ "mut" ] identifier : [ Type ] [ "=" Expression ] ";";

	UnaryExpr = unaryOp PrimaryExpr;
	MemberAccess = [ { PrimaryExpr "." } ] identifier

	PrimaryExpr =
		Type
		| "(" PrimaryExpr ")"
		| UnaryExpr
		| "^" MemberAccess
		| Literal
		| Call;
		
	Call = PrimaryExpr "(" [ ExpressionList ] ")";

	Statement = ( StructuredStat | UnstructuredStat );
	StructuredStat = Block | IfStat | MatchStat | ForStat;
	UnstructuredStat = Declaration | Impl | LeaveStat | IncDecStat | Assignment.
	LeaveStat = ReturnStat | BreakStat | ContinueStat;
	Assignment = PrimaryExpr "=" Expression;
	ReturnStat = "return" [ Expression ] ";";
	BreakStat = "break" ";"
	ContinueStat = "continue" ";"
	IncDecStat = Expression ( "++" | "--" );
	
	
	
	