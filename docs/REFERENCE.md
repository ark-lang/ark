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
	
### Using Files
Currently, we're still figuring things out with c bindings and file inclusion, however you can call C functions like so:

	use "stdio.h";

	fn main(): int {
		int swag = 10;
		printf("this is my function to print out this variables value, which is: %d\n", swag);
		return 0;
	}

### Arrays [UNIMPLEMENTED]
Arrays are very similar to languages such as C, however we have opted for a cleaner approach for dynamic arrays.

#### Static Arrays
A static array is defined as follows:

	[data_type] [name] [[size as number]];
	
For example:

	int cats[256];
	my_struct structs[5];

Values can be defined to an index, for example:

	cats[12] = 21;

Values can also be declared in the arrays definition:

	int cats[4] = [0, 1, 2, 3];
	
#### Dynamic Arrays
Dynamic arrays are similar to a static array, however they are allocated on the heap, and are resized automatically. The dynamic
arrays are similar to static ararys, the difference is that when you exceed the initial size of the array, it will reallocate
enough memory to store more elements. Another neat feature with dynamic arrays in Alloy is that you can also define a step. This is
really useful, since you can define linear or exponential arrays, or your own pattern. 

For example, it is a good practice to exponentially increase the size of the array, let's say the arrays initial size is 16, and you
need some more space, instead of adding 1 to the size then having to reallocate after adding another item, you allocate enough space
for 16 more items, so the size is now 32. Once you reach that limit, you allocate 64 items, then 128, 256, and so on. This is good if
you know you are going to be adding a lot of items to an array, otherwise if you resized just enough space for one item and you are
adding loads of items to the array, it can be expensive since there are a lot of `realloc` calls!

Anyways, a dynamic array is defined as follows:

	[type] [name] <step>[initial_size];
	
An example would be:

	Token ^tokens <8>[8];
	
This would allocate enough memory for 8 tokens, and when you re-allocate it will allocate 8 more elements worth of memory. You can
also apply an expression, for example:

	Token ^tokens <tokens.size * 2>[8];
	
Which would allocate `size * 2` everytime you exceed the limit.

### Option Types [UNIMPLEMENTED]
todo ...

### Structures
todo...
















