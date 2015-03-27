# Alloy Reference
This document is an informal specification for Alloy, a systems programming language.
**IMPORTANT NOTICE: I've just re-written the compiler, so a lot of these are parsed, but the parser is still incomplete so I can't guarantee they work still!**

## Guiding Principles
Alloy is a systems programming language, intended as an alternative to C. It's main purpose is to modernize C, without
deviating from C's original goal of simplicity. 

The design is motivated by the following:

* Fast
* Multi-Paradigm
* Strongly Typed
* Concise, clean and semantic syntax
* **No** garbage collection
* Efficient

The language should be strong enough that it can be self-hosted.

## Language Definition
This is the definition for the Alloy programming language. This section is an attempt to accurately describe the language
in as much detail as possible, including grammars, syntax, etc.

### Program Structure
A program consists of multiple functions, one function by convention called "main", is the entry point of execution. This is
the first function invoked at run time, and the core of the program.

### Values and References
Alloy is pass by value, if you called a function that takes an array, if you were to simply pass the array,
it would make a copy of this array. If you want to pass a reference, one must explicitly pass a pointer to the array.

## Notation
Before reading the notation, it is recommended you have an understanding of Extended Backus-Naur Form, the code snippet that introduce certain
syntax will typically show the grammar for the syntax, and an example of the syntax in use.

## Characters and Letters

    digit = { "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9" } .
    letter = "A" | "a" | ... "Z" | "z" | "_" .

Letters and digits are ASCII for now, however we may allow for unicode later on.

## Operators
Below are the many operators availible. The sign operators are unary operators, which can prefix numeric literals. The `escaped_char` are
character escapes. The third section of characters are various logical, relational and arithmetic operations. The fourth set of characters
group specific operations into either the unary or binary categories.

	sign = "+" | "-" .
	
	escaped_char = "\" ( "a" | "b" | "f" | "n" | "r" | "t" | "v" | "\" | "'" | """ ) .
	
	logOp = "||" | "&&" .
	relOp = "==" | "!=" | "<" | "<=" | ">" | ">=" .
	addOp = "+" | "-" | "|" | "^" .
	mulOp = "*" | "/" | "%" | "<<" | ">>" | "&" .

	binaryOp = logOp | relOp | addOp | mulOp .
	unaryOp = "+" | "-" | "!" | "^" | "<" | ">" | "*" | "&" .

## Identifiers
An identifier is a name for an entity in the source code, for example a variable, a type, a function, etc. An identifier must **not** be
a reserved word.

	identifier = letter { letter | digit } .

	some_thing
	a
	_example
	Amazing_NumberTwo2

## Source Code Representation
The source code is in unicode text, encoded in utf-8. Source text is case-sensitive. Whitespace is blanks,
newlines, carriage returns, or tabs. Comments are denoted with `//` for single line, or `/* */` **without nesting**.
For simplicity, identifiers are treated as ASCII, however unicode may be supported in the future, but aren't a priority.

## Reserved Words
These words are reserved, i.e they cannot be used in identifiers.
     
	u64 u32 u16 u8 i64 i32 i16 i8 f64 f32 bool char int float struct
	enum fn void for loop while if else mut return continue break
	use do 

## Types

    Type = TypeName | ArrayType | PointerType .

## Basic Types
Alloy defines a number of basic types. 

	u64         unsigned 64-bit integer
	u32         unsigned 32-bit integer
	u16         unsigned 16-bit integer
	u8          unsigned 8-bit integer
	
	i64         signed 64-bit integer
	i32         signed 32-bit integer
	i16         signed 16-bit integer
	i8          signed 8-bit integer
	
	f64         IEEE-754 valid 64-bit floating point number
	f32         IEEE-754 valid 32-bit floating point number

Additionally, there are several platform-specific type aliases that are declared by Alloy: int, float, and double. The bit width of each of these types is natural for their respective types of the given
platform. For example, an integer `int` is typically an i32 or a 32-bit architecture, and an i64 on a 64-bit architecture.

Other basic types include:
	
	bool        alias of u8
	char        alias of i8

`true` and `false` are reserved words, which represent the corresponding boolean constant values.

## Struct Types
Struct types are similar to C structs. Each member in a struct represents a variable within the data structure.

	StructDecl = "struct" "{" [ FieldList ] "}" .
	FieldList = FieldDecl { ";" FieldDecl } .
	FieldDecl = [ "mut" ] Type IdentifierList .
	
	struct Cat {
		string name;
		int age;
	}

Structure members cannot be initialized, this is for semantics.
	
## Pointer Types
Pointers are similar to C, however pointer arithmetic is not permitted. They are also denoted with the caret symbol '^', instead of an asterisks '*'.

	int ^x;
	
## Blocks
There are two types of blocks, a multi-block, denoted with two curly braces `{}`. And a single-block, denoted with an arrow `->`. A multi-block contains multiple statements, and a single-block can only contain a single statement.

	Block = ( "{" [ StatementList ";" ] "}" | "->" Statement ) .
	

	{
		statement;
	}
	
	-> statement;
	
## Functions
Functions contain declarations and statements. They can be recursive. Functions can either return a value, or void. A function consists of a function prototype, which is then followed by a Block, which can either be a single-block or a multi-block.

	FunctionDecl = "fn" FunctionSignature ( ";" | Block ) .
	FunctionSignature = [ Receiver ] identifier Parameters ":" [ "mut" ] Type .
	Receiver = "(" [ "mut" ] Type identifier ")"
	Parameters = "(" [ parameterList ] ")" .
	ParameterList = ParameterSection { "," ParameterSection } .
	ParameterSection = [ "mut" ] Type IdentifierList .

	fn add(int a, int b): int {
		return a + b;
	}
	
	// simplified to
	fn add(int a, int b): int -> return a + b;
	
## Methods
A method is a function bound to a specific structure.

	struct Point {
		float x;
		float y;
	}
	
	fn (Point ^p) distance(): float {
		return p.x * p.x + p.y * p.y;
	}
	
Creates a method for the structure Point. It is worth noting that methods are not declared within their structure declaration. When the method is invoked, a method behaves like a function, in which the first argument is the receiver. However, in Alloy the receiver is bound to the method using the following notation:

	receiver.method();
	
For instance, given a Point, namely `p`, one may do the following:

	Point ^p = { 0.0, 0.0 };
	float dist = p.distance();

# TODO!
**We still have to write up the rest for the following!! However, I've given some basic guidelines for the syntax**

## Variables

	int x = 5;
	x = 10; // fails not mutable!
	
	mut int x = 5;
	x = 20; // yay!

## If Statements

	if x == 5 {
	
	}
	
## Else If

	if x == 5 {
	
	}
	else if x == 12 {
	
	}
	
## Else

	if x == 12 {
	
	}
	else {
	
	}
	
## For Loops

	for type index: (0, 10) {
	
	}
	
	int step = 2;
	for type _: (0, 100, step) {
	
	}
	
## While Loops

	while true {
		// do stuff
	}
	
	while x == 5 {
		// do stuff
	}
	
	do x == 5 {
		// do stuff at least once
	}
	
	loop {
		// infinite till break!
	}
	
## Match

	int y = 10;
	int some_value = 5;
	int x = some_value;
	match x {
		5 {
		
		},
		some_value -> y = 21,
		_ {
			// dont care
		}
	}
	
## Pointers!

	// short hand for malloc, and setting value
	int ^x = 10;
	
	// alloc it yourself	
	int ^y = alloc;
	
	// set the value
	^y = 10;
	
	// set x to the value of y
	int x = ^y;
	
## Memory Model!

	Just about how frees are inserted automatically 
	by the compiler!

## Enumeration

	// anonymous enumeration!
	enum _ {
		SWAG,
		ASDAS,
		SAINSBURYS
	}
	
	enum NotAnonymous {
		CAR,
		DOG,
		LION
	}
	
	int x = NotAnonymous::CAR;
	
	match x {
		NotAnonymous::CAR {
		
		},
		SWAG {
		
		}
	}

## Option Types

	fn divide(int a, int b): Option<int> {
		if b == 0 {
			return None;
		}
		else {
			return Some(a / b);
		}
	}

## Enumerated Structs
This is a work in progress feature, if you don't like it or have a better idea, go post an Issue!

	struct Cat {
		int x;
	}
	
	struct Dog {
		int x;
	}
	
	variant Something {
		Cat,
		Dog
	}	

	fn doStuff(Something s): void {
		s::Cat.x = 10;		// only do this if s is of type Cat
		s::Dog.x = 20;		// only do this if s is of type Dog
	}		
	









