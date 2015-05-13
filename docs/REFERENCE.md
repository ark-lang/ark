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

The language should be strong enough that it can be self-hosted.

## Comments
Single-line comments are denoted with two forward slashes:

	// vident pls give me some bread

Multi-line comments are denoted with a forward slash followed by an asterisks. They must be closed
with an asterisks, followed by a forward slash. For instance:

	/*
		vident pls give me some bread
	 */

Documenting comments are written with a pound symbol '#', these should be used for
documenting functions, structures, enumerations, etc.

	# Does some stuff then it exits the program
	fn main(): int {

	}
	
Comments cannot be nested.

## Primitive Types
Alloy provides various primitive types:

	int			at least 16 bits in size
	float		IEEE 754 single-precision binary floating-point
	double		IEEE 754 double-precision binary floating-point
	bool		unsigned 8 bits
	char		signed 8 bits
	uint 		an unsigned integer at least 16 bits in size
	
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

## Literals
Integer literal:

	1234

Binary literal:

	0b10011010010

Hex literal:

	0x4d2

Octal literal:

	0o2322

Floating-point literal (Note that there is no `f` or `d` suffix. For a numerical literal to be floating point, just make sure it has a `.`):

	12.34

### Type Inference
The syntax for type inference is nearly identical to a typical variable declaration, all
you have to do is omit the type, for instance:

	my_type := 5;
	my_double := 5.3;

Type-inference is still an early implementation, decimal values will be stored as doubles
as opposed to floats. This is so you don't lose precision. Type-inference works with variables,
literals, and function calls. For instance:

	a := 5;
	b := 10;
	c := a + b * 2;

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

### Single-line Functions
If you have a function that consists of a single statement, it's suggested that you use the
`->` operator instead of an entire block. The `return` keyword is omitted.

	fn add(a: int, b: int): int -> a + b;

## Structures
A structure is a complex data type that defines a list of variables all grouped under one name
in memory:

	struct Cat {
	
	}
	
A structure contains variable definitions separated by commas. Trailing commas are allowed. For example, we can write
a structure to define a Cat's properties like so:

	struct cat {
		name: str,
		age: int,
		weight: float
	}
	
While structures are more complex types, they are defined just as any ordinary type such as an integer
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

### Default structure values
You can also have a structure that has default values if there are no
initial values given:

	struct Cat {
		name: str = "Terry",
		age: int = 0
	}

This means that when you make an instance of the structure, but do not
give a value for the name -- or any other member in the structure -- it
will fallback to the default value specified on the structures declaration.

	cat: ^Cat = {
		age: 12
	};

	printf("%s is %d years old\n", cat.name, cat.age);

Will give us:

	Terry is 12 years old

## Implementations & Methods
An `impl` is an implementation of the given `struct`, or structure. An implementation contains methods
that 'belong' to the structure, i.e you can call these methods through the given structure and the methods
can manipulate their owners data. First we must have a struct that will be the owner of these methods:

	struct Person {
		name: str,
		age: int,
		gender: Gender
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
		p: Person;
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

	!use "myfile"

We could write a really simple "math" library with some bindings to C's `<math.h>`:

	// math.aly
	fn acos(mut x: double): mut double;
	fn asin(mut x: double): mut double;
	fn atan(mut x: double): mut double;
	fn atan2(mut y: double, mut x: double): mut double;
	fn cos(mut x: double): mut double;
	fn cosh(mut x: double): mut double;
	fn sin(mut x: double): mut double;
	fn sinh(mut x: double): mut double;
	fn tanh(mut x: double): mut double;
	fn exp(mut x: double): mut double;
	fn log(mut x: double): mut double;
	fn log10(mut x: double): mut double;
	fn pow(mut x: double, mut y: double): mut double;
	fn sqrt(mut x: double): mut double;
	fn ceil(mut x: double): mut double;
	fn floor(mut x: double): mut double;

And we can use this library in a file:

	// main.aly
	!use "math" // note we don't use "math.aly", just "math"

	fn main(): int {
		mut x: double = pow(5, 2);
		printf("%f\n", x);
		return 0;
	}

Note that you need to compile any files that you use, so the code sample above
would be compiled as:

	alloyc math.aly main.aly

Note that the order matters too. We plan to fix this in the future.

### Pointers
The caret (`^`) is what we used to denote a pointer, i.e something that points to an
address in memory. Then there is the ampersand (`&`), which means **address of**. For instance:

	x: int = 5;
	y: ^int = &x;

In the above example, we create an integer `x` that stores the value `5` somewhere in memory. The variable `y` is
pointing to the address of `x`. So we have `y`, now if we try to print it out, you'll get an address. This is because
it just stores the address `x`. Now if we want to access the value at the address of x, we must dereference the pointer
that points to it. This is again done with the caret (`^`), for example:

	x: int = 5;		// 5
	y: ^int = &x;	// 0xDEADBEEF		(somewhat arbitrary address)
	z: int = ^y;	// 5				(get the value at our address 0xDEADBEEF)

We've introduced a new variable `z`, that stored the value at the address `y`.

### Managing Memory
Alloy is not a garbage collected language, therefore when you allocate memory, you must free it after you are
no longer using it.

## Flow Control
### If Statements
If statements are denoted with the `if` keyword followed by a condition. Parenthesis around an expression
are optional:

	if x == 1 {
		....
	}

### Match Statements
Match statements are very similar to switchs in C. Note that by default, a match clause will break instead
of continuing to other clauses. A match is denoted with the `match` keyword, followed by something to match
and then a block:

	match some_var {
		...
	}

Within the match statement, are match clauses. Match clauses consist of an expression, followed by a single
statement operator `->`, or a block if you want to do multiple statements:

	match x {
		0 -> ...;
		1 {

		};
	}

Each clause must end with a semi-colon, including blocks.

### For Loops
For loops are a little more different in Alloy. There are no while loops, or do while loops.

#### Infinite Loop
If you want to just loop till you break, you write a for loop with no condition, for instance:

	for {
		printf("loop....\n");
	}

#### "While" Loop
If you want to loop while a condition is true, you do the same for loop, but with a condition
after the `for` keyword:

	for x {
		printf("loop while x is true\n");
	}

#### "Traditional" For Loop
Finally, if you want to iterate from A to B or vice versa, you write a for loop with two conditions.
The first being the range, the second condition being the step. For instance:

	for x < 10, x++ {
		...
	}

Note that `x` is not defined in the for loop, but must be defined outside of the for loop. For instance:

	mut x: int = 0;
	for x < 10, x++ {
		...
	}

## Option Types
Option types represent an optional value, they can either be `Some` or `None`, i.e. they can either have
a value, or not have a value -- they are often paired with `match`.
An option type is denoted with an open angular bracket, the type that is optional, and a closing angle
bracket.

	fn example(a: ?int) {
		match a {
			Some -> printf("wow!\n");
			None -> printf("Damn\n");
		}
	}

Here's an example with an Option type as a function return type. These are especially useful for
cleanly checking for errors in your code. Note that the example below is semi-pseudo code, i.e.
the functions that it calls do not exist, since we haven't written any file IO libraries for Alloy
yet.

	fn read_file(name: str): ?str {
		read = non_existent_file_reading_function(str, ...);
		if read {
			return Some(read.contents);
		}
		return None;
	}

	fn main() {
		file_name: str = "vident_top_ten_favorite_bread_types.md";
		file_contents: str = read_file(file_name, ...);
		
		match file_contents {
			Some -> printf("file %s contains:\n %s", file_name, file_contents)
			None -> printf("failed to read file!");
		}

		return 0;
	}

## Enums
An enumeration is denoted with the `enum` keyword, followed by a name, and a block. The block contains
the enum items, which are identifiers (typically uppercase) with an optional default value. Enum items
must be terminated with a comma (excluding the final item in the enumerion). For example:

	enum DogBreed {
		POODLE,
		GRAYHOUND,
		SHIH_ZU
	}

To refer to the enum item, you need to specify the name of the enumeration, followed by two colons `::`,
and finally the enum item.

	fn main(): int {
		match x {
			DogBreed::POODLE -> ...;
		}
	}

## Arrays
An array is a collection of data that is the same type. They are defined as follows:

	mut name_of_array: [size]type;

Note that the size of the array is a must if you are not statically initializing the
array. For instance, an array of points would be defined as follows:

	mut points: [5]int; // can hold up to 5 points.

You can then set the values of the array by specifying the name of the array to modify,
an opening square bracket, the index of the array that you are changing, and a closing
bracket:

	// set the first value in the array to be 10.
	points[0] = 10;

To retrieve a value in an array, you do the same syntax, but you do not assign a value. For
instance, I could store the value in another variable like so:

	x: int = points[0]; // is now 10

### Statically Initializing an array
If you already know what data needs to be stored in the array, you can simply initialize
the array on its declaration:

	some_array: []int = [
		0, 1, 2, 3, 4
	];

Note that I did not specify a size in the block this time, and that there is also a semi-colon `;`
at the end of the initializing block, this is because it's still a statement.

## Generics
we're still thinking about this... want to suggest an idea/have your say? Post an issue, or comment
on an existing one (relevant to the topic) [here](https://github.com/felixangell/alloy/issues).

## Macro System
we're still thinking about this... want to suggest an idea/have your say? Post an issue, or comment
on an existing one (relevant to the topic) [here](https://github.com/felixangell/alloy/issues).
