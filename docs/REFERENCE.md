# Alloy Reference
This document is an informal specification for Alloy, a systems programming language. 

# Table of Contents

- [Alloy Reference](#alloy-reference)
- [Table of Contents](#table-of-contents)
  - [Guiding Principles](#guiding-principles)
  - [Comments](#comments)
  - [Primitive Types](#primitive-types)
  - [Precision Types](#precision-types)
    - [`usize`](#usize)
  - [Variables](#variables)
  - [Type Inference](#type-inference)
  - [Type Casting](#type-casting)
  - [Literals](#literals)
    - [Numeric Literals](#numeric-literals)
    - [Text Literals](#text-literals)
    - [Mutability](#mutability)
  - [Tuples](#tuples)
  - [Functions](#functions)
    - [Function Return Types](#function-return-types)
    - [Single-line Functions](#single-line-functions)
  - [Structures](#structures)
    - [Default structure values](#default-structure-values)
  - [Implementations & Methods](#implementations-&-methods)
  - [Function Prototypes](#function-prototypes)
    - [Calling C Functions](#calling-c-functions)
  - [File Inclusion](#file-inclusion)
  - [Pointers](#pointers)
  - [Managing Memory](#managing-memory)
  - [Flow Control](#flow-control)
    - [If Statements](#if-statements)
    - [Match Statements](#match-statements)
    - [For Loops](#for-loops)
      - [Infinite Loop](#infinite-loop)
      - ["While" Loop](#while-loop)
      - [Indexed For Loop](#indexed-for-loop)
  - [Option Types](#option-types)
  - [Enums](#enums)
  - [Arrays](#arrays)
    - [Statically Initializing an array](#statically-initializing-an-array)
  - [Generics](#generics)
  - [Macro System](#macro-system)

## Guiding Principles
Alloy is a systems programming language, intended as an alternative to C. It's main purpose is to modernize C, without deviating from C's original goal of simplicity. Alloy's frontend is written in C, with LLVM being used for the backend. For users who have followed the project for a while now, it may seem very abrupt, the fact that we've decided to use LLVM as the backend. After rigorous discussions, we were of the idea that generating C had some serious limitations as far as feature implementations were concerned, ergo we decided to use LLVM. Apart from giving us ground to expand Alloy's features, it also gave us comparable performance, both of which we aimed for initially.

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
	func main(): int {

	}
	
Comments cannot be nested.

## Primitive Types
Alloy provides various primitive types:

|Type|Description|
|----|-----------|
|int|at least 16 bits in size|
|float|IEEE 754 single-precision binary floating-point|
|double|IEEE 754 double-precision binary floating-point|
|bool|unsigned 8 bits|
|uint|an unsigned integer at least 16 bits in size|
|rune|signed 32 bits, used for holding a UTF-8 character (TODO)|
	
## Precision Types
The precision types are there for when you want a data type to be a specific
size. If you're writing portable code, we suggest that you use the primitive
types, however the precision types are available for when you need them.
Note: the C `char` type corresponds to the `i8` type.

|Type|Description|
|----|-----------|
|f32|IEEE 754 single-precision binary floating-point|
|f64|IEEE 754 double-precision binary floating-point|
|f128|IEEE 754 quadruple-precision binary floating-point|
|i8|signed 8 bits|
|i16|signed 16 bits|
|i32|signed 32 bits|
|i64|signed 64 bits|
|i128|signed 128 bits|
|u8|unsigned 8 bits|
|u16|unsigned 16 bits|
|u32|unsigned 32 bits|
|u64|unsigned 64 bits|
|u128|unsigned 128 bits|

Warning: the `i128`, `u128` and `f128` types are only supported on the LLVM backend. On the C backend, they are simple aliases for their 64-bit equivelants.

### `usize`
The `usize` or unsigned size, it can not represent any negative values. It is used
when you are counting something that cannot be negative, typically it is used for
memory.

	usize		//unsigned at least 16 bits
	
## Variables
Unlike languages like C or C++, variables are immutable unless otherwise preceeded by
the `mut` keyword. A variable can be defined like so:

	name: type;
	
And declared as follows:

	name: type = some_val;

Variables whose type is to be inferred can be declared as follows:

    name := some_val;
More about type inference is discussed below.

## Type Inference
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

We can also infer types in the following manners:

    fn add(a: int, b: int): int -> a + b;
    x := add(5, 5); // x is int

    struct Cat {
    	name: str,
    	weight: float,
    	age: int
    };
    pew: Cat;
    b := pew; // b is of type Cat and contains the values that the struct `x` contains 

    enum Wew {
    	WOW,
    	MEOW
    }

    mew := Wew::WOW;

    ar : []int = [ 1, 2, 3, 4 ];
    b := ar[2]; // b is of type int


## Type Casting
In Alloy, type casting can be done as follows:

    mut my_val: float = 1.3;
    mut my_val_two: int = int(my_val); // my_val_two is now 1

To do type casting, the type its being casted to must be added, followed by the value being casted in parentheses. 

    val: int = int(4.5); // val is now 4

## Literals

### Numeric Literals
Integer literal:

	1234

You can use underscores in integer literals to improve readability:

	1_234

Binary literal:

	0b10011010010

Hex literal:

	0x4d2

Octal literal:

	0o2322

Floating-point literals contain a period. If a literal lacks a period, or you want to specify the precision, use the `f` suffix for 32 bits (`f32`) or the `d` suffix for 64 bits (`f64`):

	12.34

### Text Literals
A string literal is enclosed with double-quotes:

    "Hello!"

A character literal is enclosed with single-quotes:

    'a'

The following escape sequences are available:

|Sequence|Description|
|--------|-----------|
|\a|alarm/bell|
|\b|backspace|
|\f|formfeed|
|\n|newline|
|\r|carriage return|
|\t|horizontal tab|
|\v|vertical tab|
|\\\\|backslash|
|\'|single quotation mark|
|\"|double quotation mark|
|\?|question mark|
|\oNNN|octal number|
|\xNN|hex number|

### Mutability
When a variable is mutable, it can be mutated or changed. When a variable is immutable,
it cannot be changed. By default, Alloy assumes that a variable you've defined is immutable.
You can however, specify the `mut` keyword before the declaration. This will inform the compiler that you intend
to modify the variable later in the program. For instance:

	x: int = 5;
	x = 10;         // ERROR: x is constant!
	
We can also error it like so:

	x: int;         // ERROR: no value assigned for constant!
	
Why? Because variables are treated as constants unless otherwise specified; therefore they must 
have a value assigned on definition.

## Tuples
A tuple is defined in a manner similar to a variable. However you specify the type and the values it contains within the pipe symbol (`|`). For instance:

	mut my_tuple: |int, int|;

To initialize the tuple with values, we use a similar notation to the tuples signature, and we
wrap the values in pipes. The order should be the same as the order used in signifying the types in the signature.
For instance:

	// this is valid
	mut my_type: |int, double| = |10, 3.4|;
	
	// the following errors, note the order of the types/values. 
	// the int must come before the double in the values
	mut another_type: |int, double| = |3.4, 10|; 

## Functions
A function defines a sequence of statements and an optional return value, along with a name,
and a set of parameters. Functions are declared with the keyword `func`, followed by a name to
identify the function, and then a list of parameters. Finally, an optional colon `:` followed by
a return type, e.g. a struct, data type or `void` can be added. Note that if you do not specify a colon and
a return type, it is assumed that the function returns the `void` type.

An example of a function:

	mut int result;

	// if the function returns void,
	// the return type is optional, thus
	// func add(x: int, y: int) { }
	// is perfectly valid.
	func add(x: int, y: int): void {
		result = x + y;
	}

### Function Return Types
We can simplify this using a return type of `int`, for instance:

	func add(x: int, y: int): mut int {
		return x + y;
	}

### Single-line Functions
If you have a function that consists of a single statement, it's suggested that you use the
`->` operator instead of an entire block. The `return` keyword is omitted.

	func add(a: int, b: int): int -> a + b;

## Structures
A structure is a complex data type that defines a list of variables all grouped under one name
in memory:

	struct Cat {
	
	}
	
A structure contains variable definitions separated by commas. Trailing commas are allowed. For example, we can write
a structure to define a Cat's properties like so:

	struct Cat {
		name: str,
		age: int,
		weight: float
	}
	
While structures are more complex types, they are instantiated just as any ordinary type, such as an integer
or a float:

	mut terry: Cat;
	
Note how the structure declared is mutable. This is because we aren't declaring any of the fields in the
structure. We can define the contents of the structure like so:

	terry: Cat = {
		name: "Terry",
		age: 2,
		weight: 3.12
	};
	
The struct initializer is a statement, therefore it must be terminated with a semi-colon. Note that
the values in the struct initializer do not have to be in order, but we suggest you do to keep things
consistent.

A structure declaration can also be preceeded by the `packed` keyword. The `packed` keyword prevents aligning of structure members according to the platform the user is on, i.e. 32-bit or 64-bit. A good article going over padding and data structure alignment can be found on [this Wikipedia page and can be a good resource for the curious.](http://en.wikipedia.org/wiki/Data_structure_alignment). A packed structure can be declared like so:

    packed struct Cat {
	    name: str,
		age: int,
		weight: float
	}

### Default structure values
A structure's members can also be assigned default values:

	struct Cat {
		name: str = "Terry",
		age: int = 0
	}

In this example, upon creation of an instance of the `Cat` structure, the value of the `name` member will by default be `"Terry"`, provided a custom value has not already been assigned. Below, we create an instance of the structure `Cat`, called `cat`, and *only* give its `age` member a value. This means that the `name` member of the instance already holds the default value "Terry".

	cat: ^Cat = {
		age: 12
	};

Therefore running

	printf("%s is %d years old\n", cat.name, cat.age);

will give us:

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
		func say() {
			println("Hi my name is %s and I'm %d years of age.", self.name, self.age);
		}
	}

To access the structure that the we're implementing, you use the `self` keyword. To call the `say`
function we defined, you need to have an instance of the structure. Functions can then be accessed
via the dot operator:

	func main(): int {
		p: Person;
		p.say();
	}

## Function Prototypes
A function prototype is similar to the syntax for a function declaration, however instead of using
curly braces to start a new block, you end the statement with a semi-colon. A function prototype is a good way of defining all the functions you will be using in your program before actually implementing them. For example, a function prototype
for a function `add` that takes two parameters (both integers), and returns an integer would be as follows:

	func add(a: int, b: int): int;
The function can then be implemented elsewhere in the program. While (in most cases) a trivial move, sometimes adding the function prototype at the start of the program before implementing it elsewhere is considered good practice. 

### Calling C Functions
You can use the function prototypes showcased above to call c functions. Say we wanted to use the `printf`
function in `stdio`, we create a prototype for it. Note that the printf is a variadic function, i.e. it can
take an unspecified amount of arguments. This is denoted with an ellipses in C, and in Alloy, an underscore.
Note that this is mostly for backwards compatibility with C code, and we don't suggest you use it in your code
generally. Once you've created the prototype, it is called like any other function.

Here's an example of printf in Alloy:

	// main.aly
	func printf(format: str, _): int;

	// usage
	func main(): int {
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
	func acos(mut x: double): mut double;
	func asin(mut x: double): mut double;
	func atan(mut x: double): mut double;
	func atan2(mut y: double, mut x: double): mut double;
	func cos(mut x: double): mut double;
	func cosh(mut x: double): mut double;
	func sin(mut x: double): mut double;
	func sinh(mut x: double): mut double;
	func tanh(mut x: double): mut double;
	func exp(mut x: double): mut double;
	func log(mut x: double): mut double;
	func log10(mut x: double): mut double;
	func pow(mut x: double, mut y: double): mut double;
	func sqrt(mut x: double): mut double;
	func ceil(mut x: double): mut double;
	func floor(mut x: double): mut double;

And we can use this library in a file:

	// main.aly
	!use "math" // note we don't use "math.aly", just "math"; makes it cleaner

	func main(): int {
		mut x: double = pow(5, 2);
		printf("%f\n", x);
		return 0;
	}

Note that you need to compile any files that you use, so the code sample above
would be compiled as:

	alloyc math.aly main.aly

The files must also be compiled in order. We plan to fix this soon.

## Pointers
The caret (`^`) is what we use to denote a pointer, i.e something that points to an
address in memory. The ampersand (`&`) symbol is the **address of** operator. For instance:

	x: int = 5;
	y: ^int = &x;

In the above example, we create an integer `x` that stores the value `5` somewhere in memory. The variable `y` is
pointing to the address of `x`. Printing out the value of `y` will give you random gibberish denoting the memory chunk in which `x` is located. This is because
it just stores the address to `x`. Now if we want to access the value at the address of x, we must *dereference* the pointer
that points to it. This is again done with the caret (`^`), for example:

	x: int = 5;     // 5
	y: ^int = &x;   // 0xDEADBEEF       (somewhat arbitrary address)
	z: int = ^y;    // 5                (get the value at our address 0xDEADBEEF)

We've introduced a new variable `z`, that stored the value at the address `y`, in other words, the value of `x`.

## Managing Memory
Alloy is not a garbage collected language, therefore when you allocate memory, you must free it after you are
no longer using it. We felt that, as unsafe as it is to rely on the user to manage the memory being allocated, performance takes a higher precedence. Although garbage collection makes things fool-proof and removes a significant amount of workload from the user, it inhibits the performance we were going for. 

Memory is allocated using the `alloc` keyword and freed using the `free` keyword. The size of a particular type can be found using the `sizeof` operator, much like C. The `realloc` keyword can be used to reallocate a chunk of memory, in case it needs to be of a larger/smaller size.

## Flow Control
### If Statements
If statements are denoted with the `if` keyword followed by a condition. Parenthesis around the expression/condition
are optional:

	if x == 1 {
		....
	}

### Match Statements
Match statements are very similar to C's `switch` statement. Note that by default, a match clause will break instead
of continuing to other clauses. A match is denoted with the `match` keyword, followed by something to match
and then a block:

	match some_var {
		...
	}

Within the match statement are match clauses. Match clauses consist of an expression, followed by a single
statement operator `->`, or a block if you want to do multiple statements:

	match x {
		0 -> ...;
		1 {
			...
		};
	}

Each clause must end with a semi-colon, including blocks.

### For Loops
For loops are a little more different in Alloy. There are no while loops, or do-while loops.

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

#### Indexed For Loop
Finally, if you want to iterate from A to B or vice versa, you write a for loop with two conditions-
the first being the range and the second being the step. For instance:

	for x < 10, x++ {
		...
	}

Also note that `x` is not defined in the for loop; it must be defined outside of the loop. For instance:

	mut x: int = 0;
	for x < 10, x++ {
		...
	}

## Option Types
Option types represent an optional value- they can either be `Some` or `None`, i.e. they can either have
a value, or not have a value -- they are often paired with a `match` statement.
An option type is denoted with an open angular bracket, the type that is optional, and a closing angle
bracket.

	func example(a: ?int) {
		match a {
			Some -> printf("wow!\n");
			None -> printf("Damn\n");
		}
	}

Here's an example with an `Option` type as a function return type. These are especially useful for
cleanly checking for errors in your code. Note that the example below is semi-pseudo code, i.e.
the functions that it calls do not exist, since we haven't written any file IO libraries for Alloy
yet.

	func read_file(name: str): ?str {
		read = non_existent_file_reading_function(str, ...);
		if read {
			return Some(read.contents);
		}
		return None;
	}
	func main() {
		file_name: str = "vident_top_ten_favorite_bread_types.md";
		file_contents: str = read_file(file_name, ...);
		
		match file_contents {
			Some -> printf("file %s contains:\n %s", file_name, file_contents)
			None -> printf("failed to read file!");
		}
		return 0;
	}

## Enums
An enumeration is denoted with the `enum` keyword, followed by a name and a block. The block contains
the enum items, which are identifiers (typically uppercase) with an optional default value. Enum items
must be terminated with a comma (excluding the final item in the enumerion). For example:

	enum DogBreed {
		POODLE,
		GRAYHOUND,
		SHIH_ZU
	}

To refer to the enum item, you need to specify the name of the enumeration, followed by two colons `::`,
and finally the enum item.

	func main(): int {
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

	x: int = points[0]; // x is now 10

### Statically Initializing an array
If you already know what data needs to be stored in the array, you can simply initialize
the array on its declaration:

	some_array: []int = [
		0, 1, 2, 3, 4
	];

Note that I did not specify a size in the block this time, and that there is also a semi-colon `;`
at the end of the initializing block, this is because it's still a statement.

## Generics
We're still thinking about this, want to suggest an idea/have your say? Post an issue, or comment
on an existing one (relevant to the topic) [here](https://github.com/alloy-lang/alloy/issues).

## Macro System
We're still thinking about this, want to suggest an idea/have your say? Post an issue, or comment
on an existing one (relevant to the topic) [here](https://github.com/alloy-lang/alloy/issues).
