# Alloy Reference
This document is an informal specification for Alloy, a systems programming language.

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

## Program Structure
A program consists of multiple functions, one function by default called "main", is the entry point of execution. This is
the first function invoked at run time.

## Values and References
Alloy is pass by value, if you called a function that takes an array, if you were to simply pass the array,
it would make a copy of this array. If you want to pass a reference, one must explicitly pass a pointer to the array.

## Notation
The syntax is specified using Extended Backus-Naur Form (EBNF).

### Source Code Representation
The source code is in unicode text, encoded in utf-8. Source text is case-sensitive. Whitespace is blanks,
newlines, carriage returns, or tabs. Comments are denoted with `//` for single line, or `/* */` without nesting.
For simplicity, identifiers are treated as ASCII, however unicode may be supported in the future, but aren't a priority.

## Types

### Basic Types
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
	
Additionally, there are several platform-specific type aliases that are declared by Alloy: int, float, and double. The bit width of each of these types is natural for their respective types of the given platform. For example, an integer `int` is typically a i32 or a 32-bit architecture, and an i64 on a 64-bit architecture.

### Struct Types
Struct types are similar to C structs. Each member in a struct represents a variable within the data structure.

	struct Cat {
		string name;
		int age;
	}
	
Structure members can also be initialized on declaration, for instance:

	struct Cat {
		string name = "Jim";
		int age = 21;
	}
	
Is perfectly valid.
	
### Pointer Types
Pointers are similar to C, however pointer arithmetic is not permitted. They are also denoted with the caret symbol '^', instead of an asterisks '*'.

	int ^x;
	
### Blocks
There are two types of blocks, a multi-block, denoted with two curly braces `{}`. And a single-block, denoted with an arrow `->`. A multi-block contains multiple statements, and a single-block can only contain a single statement.

	{
		statement;
	}
	
	-> statement;
	
### Functions
Functions contain declarations and statements. They can be recursive. Functions can either return a value, or void. A function consists of a function prototype, which is then followed by a Block, which can either be a single-block or a multi-block.

	fn add(int a, int b): int {
		return a + b;
	}
	
	// simplified to
	fn add(int a, int b): int -> return a + b;
	
### Methods
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

	p.distance();

### Mutability
Alloy defines any variable or type instance as constant, unless otherwise specified with the `mut` keyword. This means that whenever you define/declare a variable or structure, it will be immutable. This means the following is **invalid**:

	int x = 5;
	x = 10; 			// error! 
	
This is because `x` by default is immutable, i.e it cannot be changed as it is constant. You can apply the `mut` keyword, to make the variable mutable.

	mut int x = 5;
	x = 10; 			// yay!
	
This applys to structures, structure members, parameters, variables, and function returns. For instance,

	struct Cat {
		int age = 10;
	}
	
	Cat cat;
	cat.age = 21;		// error!
	
However, the following is legal:

	struct Cat {
		mut int age = 10;
	}
	
	Cat cat;
	cat.age = 21;