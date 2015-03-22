# Alloy Reference
This is a reference for the Alloy programming language. Please note that the language is in constant development,
therefore this reference can become outdated at any time, and it can be days before it is updated to the current stage
of the compiler.

## Introduction
The aim of this compiler is to reinvent C. We love C, it's simple, fast, expressive, and cross-platform. However, it
has it's flaws:

* buffer overflows
* no string type
* no booleans without stdbool
* 30 years old
* inconsistencies, especially in error handling

Our goal with Alloy is to fix these errors, yet maintaining a cleaner, simpler syntax.
A lot of the syntax for Alloy is inspired by existing languages, such as Rust, Go, and Java. The language itself is
also heavily inspired by the simplicity of C.

Things you wont see in Alloy:

* Garbage Collection
* Preprocessor
* Null Pointers

## Memory Model
todo

## Data Types
Alloy has no type inference, data types must be explicitlly defined. Here's a list of data types availible:

	type		equivalent
	string		char*
	
	u64			unsigned long long
	u32			unsigned int
	u16 		unsigned short
	u8			unsigned char
	
	s64			long long
	s32			int
	s16			short
	s8			char
	
	f64			float
	f32			double

## Syntax

### Semi-colons
Semi-colons are enforced in the Alloy programming language.

### Variables
#### Variable Definitions
Variable definitions are as follows:

	[type] [name];
	
For example:

	int i;
	double d;
	float f;
	bool b;
	structure_name s;

#### Variable Declarations
Variable declarations are as follows:

	[type] [name] = [expression];
	
For example:

	int x = 5;
	double d = 10.0;
	float f = 3.21;
	bool b = false;

### Functions
Function are defined as follows:

	// multiple statements
	fn [function_name]([type] [name], ...):[return_type] {
		[statement];
		[statement];
		[statement];
		[statement];
		[statement];
	}
	
	// or
	
	// single statement afterwards
	fn [function_name]([type] [name], ...):[return_type] -> [statement];

For example:

	fn do_stuff(int a): void {
		a = 5;
		global_variable = a + 2;
	}
	
	fn add(int a, int b): int -> return a + b;
	
#### Function Redirect
A function redirect is where you direct the return value from the function into a variable, for example:

	fn add(int a, int b): int -> return a + b;
	
	fn main(): void {
		int x;
		add(5, 5) -> x;
	}

The syntax is as follows:

	[function_call] -> [variable];
	
The variable must be defined.

### Using Files
Currently, we're still figuring things out with c bindings and file inclusion, however you can call C functions like so:

	use "stdio.h";

	fn main(): int {
		int swag = 10;
		printf("this is my function to print out this variables value, which is: %d\n", swag);
		return 0;
	}

Note the first line `use "stdio.h"`. This will include the standard input/output library from C. However, if you plan to use
function redirects, they won't with C function calls, for example:

	use "stdio.h"
	use "stdlib.h"
	
	fn main(): int {
		int ^swag;
		malloc(sizeof(5)) -> swag;
		^swag = 10;
		printf("swag is %d\n", ^swag);
	}

It looks like it should work, however function redirects do not work yet with C bindings. We're still working on the inclusion
of alloy files.

### Structures
todo...
















