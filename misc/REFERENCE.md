# JAYFOR REFERENCE
This is not an exact specification, but an effort to describe the langauge in
as much detail as possible. This document does not serve as an introduction to
the language, but as a reference. 

*DISCLAIMER: THE CONTENTS OF THIS REFERENCE ARE A GUIDE TO DESCRIBE THE LANGUAGE
IN AS MUCH DETAIL AS POSSIBLE. ANYTHING SPECIFIED IN THIS DOCUMENT MAY BE SUBJECT
TO CHANGE. NOTHING IS FINAL*

# Table Of Contents
* [Lexical Structure](#lexical_structure)
  * [Semi Colons](#semi_colons)
  * [Comments](#comments)
  * [Keywords](#keywords)
  * [Expressions](#expressions)
  * [Functions](#functions)
  	* [Tuples](#tuples)
  	* [Single Line Functions](#single_line_functions)
  	* [Unsafe Functions](#unsafe_functions)
  * [Conditionals](#conditionals)
  	* [If Statements](#if)
  	* [While Loops](#if)
  	* [Do-While Loops](#do)
  	* [Infinite Loops](#loop)
  	* [For Loop](#for)
  	* [Matches](#matches)
  * [Data Structures](#data_structures)
  	* [Enumeration](#enumeration)


# <a name="lexical_structure"></a>Lexical Structure
## <a name="semi_colons"></a>Semi Colons
It's important to note that semi colons are optional in the Jayfor programming language. This feature
means that you are less likely to slip up on a "missing semi-colon" error. However, you should keep it
consistent and stick with either no semi-colons or semi-colons.

## <a name="comments"></a>Comments
Comments in Jayfor code follow the general C style of line and block comments. Nested
comments are supported.
Line comments begin with exactly two forward slashes, and block comments begin with
exactly one forward slash and repeated asterisks, and are closed with exactly one asterisks
and one forward slash:

	/*
		This is a block comment.
	 */
	
## <a name="keywords"></a>Keywords

	int 	bool	float	str 	void
	enum	struct 	return 	const 	true
	false 	match 	while 	for 	do
	if 		else 	unsafe 	fn 		tup
	char	break	continue

## <a name="expressions"></a>Expressions
The following arithmetic operations are currently supported:

	+	-	/	*	%

Jayfor does not have operator precedence parsing as of writing this document, this is subject
to change, but for now expressions precedence must be explicitly defined with brackets `()`.
For example:

	5 + 5 / 10 * 2

Would be expressed as:

	((5 + 5) / (10 * 2))

## <a name="functions"></a>Functions
A function defines a sequence of statements, and an *optional* final expression,
along with a name and a set of parameters. Functions are declared with the keyword `fn`.
Functions declare a set of parameters, which the called passed arguments to the function, and
the function passes the results back to the caller.

A function must also have a return type, which defines what value the function will return to
the caller. This is specified after the colon operator, like so:

	fn add(int a, int b): int {
		return (a + b);
	}

### <a name="tuples"></a>Tuples
Functions may also return tuples. Tuples are denoted with a less than symbol `<`, a list
of data types, and a greater than symbol `>`. A tuple is an explicit data type, through
which they are declared as follows:

	tup my_tuple<int, double, str> = <5, 5.3, "string">;

A function returning a tuple is defined similarily to a function with a single return type. A
colon must be specified, however, instead of a single data type; you must provide a tuple
signature, denoted with an opening `<` and closing `>` angle bracket. The functions return 
statement must also follow this pattern:

	fn get_population(str location): <int, int, int> {
		if location == "New York" {
			return <5, 5, 5>;
		}
		return <0, 0, 0>;
	}

### <a name="single_line_functions"></a>Single Line Functions
If a function only has one return statement, you may simplify it with the `=>` operator. For
example, a function that adds two integers can be simplified to:

	fn add(int a, int b): int => return (a + b);

### <a name="unsafe_functions"></a>Unsafe Functions
If a function is later on deemed unsafe, you may use the `unsafe` keyword to give a compile
time warning if an unsafe warning were to be called, i.e:

	unsafe fn allocate_memory(int size): void {
		// some dangerous memory allocation
	}

If a developer were to call this function, a warning would be printed to the console on compile
time:

	""filename.j4":5:5: warning: use of unsafe function 'allocate_memory'!"

## <a name="conditionals"></a>Conditionals
### <a name="if"></a>If Statements
In Jayfor, an if statement is denoted with the `if` keyword, a condition, and a pair of 
curly `{}` braces. Within the braces, a list of statements are specified, which will execute
if the aforementioned condition is true:

	if condition {

	}

### <a name="while"></a>While Loops
While Loops are specified with the `while` keyword, a condition and a pair of curly `{}` braces.
Within the braces are a list of statements, which will execute if the aforementioned condition
is true.

	while condition {

	}

### <a name="do"></a>Do While Loops
Do-While loops have been simplified to a single keyword, while it is not as semantic as the
traditional do while loop, it is consistent, and a lot easier to type. A do-while loop is
specified with the `do` keyword, a condition and a pair of curly `{}` braces:

	do condition {

	}

The statements within the curly braces (block) will be executed at least once, and will continue
executing if the condition is true (and stays true).

### <a name="loop"></a>Infinite Loops
A loop is syntactic sugar for a while loop, it will keep executing the given statements until
the loop is broken out of with the `break` keyword. A loop is specified with the `loop` keyword,
and a pair of curly `{}` braces:

	loop {

	}

### <a name="for"></a>For Loops
The for loop has been simplified from its traditional syntax. A for loop is specified with the
`for` keyword, a data type (int, float, etc...), and an index to keep track of the current
iteration. If you do not care about what iteration you are on, this can be replaced with an
underscore `_`. The for loop must also be given a list of 2 - 3 parameters, the start of the loop,
the end of the loop, and the step. The loops step is how much the `index` will be incremented by
every iteration. It is an optional parameter, if no step is supplied it will default to either positive
or negative 1 -- this depends on what the step and end arguments are. Here are some examples:

	for int index: (0, 10, 1) {

	}

	// same as the above
	for int index: (0, 10) {

	} 

	// start > end, which means that step
	// wil decrement instead of increment
	for int index: (10, 0) {

	}

	for int _: (0, 10) {
		// we don't care about
		// the index, just loop
	}

### <a name="matches"></a>Matches
We feel that the switch syntax is tedious, ugly, and not as semantic as it could be. Therefore we implemented
a Rust like match:

	int value_to_match = 23;
	int value = 5;
	int another_value = 23;

	match value_to_match {
		value {
			// this will be skipped
		}
		another_value {
			// this is the result
		}
		_ {
			// this is a "default" in case value_to_match was for example 64
		}
	}

## <a name="data_structures"></a>Data Structures
### Structs
A `struct` is a data structure defined with the keyword `struct`. Here is an
example of a struct:

	struct vector {
		float x
		float y
	}

	// structs can be defined like so
	vector vec = {
		10, 10
	}

	// which is short hand for
	vector vec;
	vec.x = 10
	vec.y = 10

You can also define a struct with default values, like so:

	struct vector {
		float x = 0
		float y = 0
	}

### <a name="enumeration"></a>Enumeration
Enumerations are defined with the `enum` keyword, like so:

	enum TRAFFIC_LIGHT {
		RED
		ORANGE
		GREEN
	}

By default, the values will always start from zero, and the next value in
the enumeration will be incremented by 1. You can also change the default value with
an equal sign, like so:

	enum TRAFFIC_LIGHT {
		RED = 10
		ORANGE	// 11
		GREEN	// 12
	}

	// or

	enum PET_TYPE {
		DOG = 10
		CAT = 54
		LIZARD = 61
		DRAGON // 61
	}

Enumerations are accessed with the double-colon operator, similar to C++:

	PET_TYPE::DOG

You can match enumerations or use them in if statements:

	// set it to dog
	int x = PET_TYPE::DOG;

	if x == PET_TYPE::DOG {
		/// do stuff
	}

	match x to PET_TYPE {
		DOG {
			// im a dog
		}
		CAT {	
			// im a cat
		}
		LIZARD {
			// im a lizard
		}
		_ {
			// no idea
		}
	}