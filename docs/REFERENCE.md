# Alloy Reference
This document is an informal specification for Alloy, a systems programming language. 

## Guiding Principles
Alloy is a systems programming language, intended as an alternative to C. It's main purpose is to modernize C, without deviating from C's original goal of simplicity. Alloy is written in C, the frontend and backend is all hand-written, i.e no parser or lexer libraries, and no LLVM, etc.

The design is motivated by the following:

* Fast
* Multi-Paradigm
* Strongly Typed
* Concise, clean and semantic syntax
* **No** garbage collection
* Efficient

The language should be strong enough that it can be self-hosted.

## Comments
Single-line comments are denoted with two forward slashes:

	// vedant is a scrotum

Multi-line comments are denoted with a forward slash followed by an asterisks. They must be closed
with an asterisks, followed by a forward slash. For instance:

	/*
		Vedant is a scrotum	
	 */
	
Comments cannot be nested.

## Primitive Types
Alloy provides various primitive types:

	int			at least 16 bits in size
	float		IEEE 754 single-precision binary floating-point
	double		IEEE 754 double-precision binary floating-point
	bool		unsigned 8 bits
	char		signed 8 bits

## Precision Types
The precision types are there for when you want a data type to be a specific
size. If you're writing portable code, we suggest that you use the primitive
types, however the precision types are available for when you need them.
	
	f32			IEEE 754 single-precision binary floating-point
	f64			IEEE 754 double-precision binary floating-point
	
	i8			signed 8 bits
	i16			signed 16 bits
	i32			signed 32 bits
	i64			signed 64 bits
	
	u8			unsigned 8 bits
	u16			unsigned 16 bits
	u32			unsigned 32 bits
	u64			unsigned 64 bits

## Other
Other types that don't quite fit a category.

### `usize`
The `usize` or unsigned size, it can not represent any negative values. It is used
when you are counting something that cannot be negative, typically it is used for
memory.

	usize		unsigned at least 16 bits
	
## Variables
Unlike languages like C or C++, variables are immutable unless otherwise specified with
the `mut` keyword. A variable can be defined like so:

	name: type;
	
And declared as follows:

	name: type = some_val;

### Type Inference
The syntax for type inference is nearly identical to a typical variable declaration, all
you have to do is omit the type, for instance:

	my_type := 5;
	a_float := 5.3f;
	a_double := 5.2d;

Note how for floats and doubles, you have to be somewhat explicit, and suffix the value
with an `f` for a float, or a `d` for double. If the value is a decimal number and is
not suffixed with an `f` or a `d`, the value will be assumed to be a double. This is
because it is guaranteed that no floating-point precision errors will be caused by
the type inference.

## Tuples
A tuple is define similarly to a variable, however you specify the types in parenthesis. For
instance:

	mut my_tuple: (int, int);

To initialize the tuple with values, we use a similar notation to the tuples signature, and we
wrap the values in parenthesis. The order should be corresponding to the types in the signature,
for instance:

	// this is valid
	mut my_type: (int, double) = (10, 3.4);
	
	// the following errors, note the order of the types/values
	mut another_type: (int, double) = (3.4, 10);

### Mutability
When a variable is mutable, it can be mutated or changed. When a variable is immutable,
it cannot be changed. By default, alloy assumes that a variable you define is immutable.
You can however, specify the `mut` keyword, this will indicate to the compiler that you intend
to modify the variable later on in the code. For instance:

	x: int = 5;
	x = 10;			// ERROR: x is constant!
	
We can also error it like so:

	x: int;			// ERROR: no value assigned for constant!
	
Why? Because variables are treated as constants unless otherwise specified, therefore they must 
have a value assigned on definition.

## Functions
A function defines a sequence of statements and an optional return value, along with a name,
and a set of parameters. Functions are declared with the keyword `fn`. Followed by a name to
identify the function, and then a list of parameters. Finally, an optional colon `:` followed by
a return type, e.g. a struct, data type or `void`. Note that if you do not specify a colon and
a return type, the function is assumed to be void by default.

An example of a function:

	mut int result;

	// if the function returns void,
	// the return type is optional, thus
	// fn add(x: int, y: int) { }
	// is perfectly valid.
	fn add(x: int, y: int): void {
		result = x + y;
	}

### Function Return Types
We can simplify this using a return type of `int`, for instance:

	fn add(x: int, y: int): mut int {
		return x + y;
	}

## Structures
A structure is a complex structure that define a list of variables all grouped under one name
in memory:

	struct Cat {
	
	}
	
A structure contains variable definitions terminated by a semi-colon. For example, we can write
a structure to define a Cat's properties like so:

	struct cat {
		name: str;
		age: int;
		weight: float;
	}
	
Structures are more complex types, therefore they are defined like an orindary type such as an integer
or float:

	mut terry: cat;
	
Note how the structure declared is mutable, this is because we aren't declaring any of the fields in the
structure. We can define the contents of the structure like so:

	terry: cat = {
		name: "Terry",
		age: 2,
		weight: 3.12
	};
	
The struct initializer is a statement, therefore it must be terminated with a semi-colon. Note that
the values in the struct initializer do not have to be in order, but we suggest you do to keep things
consistent.

## Implementations & Methods
An `impl` is an implementation of the given `struct`, or structure. An implementation contains methods
that 'belong' to the structure, i.e you can call these methods through the given structure and the methods
can manipulate their owners data. First we must have a struct that will be the owner of these methods:

	struct Person {
		name: str;
		age: int;
		gender: Gender;
	}

We can then define various methods for this structure. To do so, they must be wrapped in an `impl`. The `impl`
or implementation, must declare what it is implementing. In this case `Person`.

	impl Person {
		...
	}

The functions for `Person` are declared inside of this `impl`:

	impl Person {
		fn say() {
			println("Hi my name is %s and I'm %d years of age.", self.name, self.age);
		}
	}

To access the structure that the we're implementing, you use the `self` keyword. To call the `say`
function we defined, you need to have an instance of the structure. Functions can then be accessed
via the dot operator:

	fn main(): int {
		Person p;
		p.say();
	}

## Function Prototypes
A function prototype is similar to the syntax for a function declaration, however instead of using
curly braces to start a new block, you end the statement with a semi-colon. For example, a function prototype
for a function `add` that takes two parameters (both integers), and returns an integer would be as follows:

	fn add(a: int, b: int): int;

### Calling C Functions
You can use the function prototypes showcased above to call c functions. Say we wanted to use the `printf`
function in `stdio`, we create a prototype for it. Note that the printf is a variadic function, i.e. it can
take an unspecified amount of arguments. This is denoted with an ellipses in C, and in Alloy, an underscore.
Note that this is mostly for backwards compatibility with C code, and we don't suggest you use it in your code
generally. Once you've created the prototype, it is called like any other function.

Here's an example of printf in Alloy:

	// main.aly
	fn printf(format: str, _): int;

	// usage
	fn main(): int {
		printf("this is a test\n");
		return 0;
	}

## File Inclusion
File inclusion is very simple in Alloy. One of the problems with C is the tedious header files. To include
a file, you must use the `use` macro<sup>disclaimer: not an actual macro yet, but it still works</sup>, which is
the `use` keyword followed by the filename to include (minus the `aly` extension) in quotes. For example:

	use "myfile"

We could write a really simple "math" library with some bindings to C's `<math.h>`:

	// math.aly
	fn acos(x: double): double;
	fn asin(x: double): double;
	fn atan(x: double): double;
	fn atan2(y: double, x: double): double;
	fn cos(x: double): double;
	fn cosh(x: double): double;
	fn sin(x: double): double;
	fn sinh(x: double): double;

	// just for the printf example, it's somewhat irrelevant
	fn printf(format: str, _): int;

And we can use this library in a file:

	// main.aly
	use "math" // note we don't use "math.aly", just "math"

	fn main(): int {
		x: double = cos(3.141);
		printf("%d\n", x);
		return 0;
	}

