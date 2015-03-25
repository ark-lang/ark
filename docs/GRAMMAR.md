This document specifies the grammar for the Alloy programming language. It's still a work in progress.

    digit = { "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9" } .
    letter = "A" | "a" | ... "Z" | "z" | "_" .
    identifier = letter { letter | digit } .
	
	Type = TypeName | ArrayType | StructType | FunctionType | PointerType .
	
	TypeName = identifier .
	
	ArrayType = "[" [ Expression ] "]" Type .
	
	StructType = "struct" "{" [ FieldList ] "}" .
	FieldList = FieldDecl { ";" FieldDecl } .
	FieldDecl = IdentifierList Type .

	FunctionType = "fn" FunctionSignature .
	FunctionSignature = [ Receiver ] identifier Parameters ":" Type .
	Receiver = "(" Type identifier ")"
	Parameter = "(" [ parameterList ] ")" .
	ParameterList = ParameterSection { "," ParameterSection } .
	ParameterSection = Type IdentifierList .

	PointerType = "^" Type .
	