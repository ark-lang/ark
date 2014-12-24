# JAYFOR REFERENCE
This is not an exact specification, but an effort to describe the langauge in
as much detail as possible. This document does not serve as an introduction to
the language, but as a reference. 

*DISCLAIMER: THE CONTENTS OF THIS REFERENCE ARE A GUIDE TO DESCRIBE THE LANGUAGE
IN AS MUCH DETAIL AS POSSIBLE. ANYTHING SPECIFIED IN THIS DOCUMENT MAY BE SUBJECT
TO CHANGE. NOTHING IS FINAL*

# Table Of Contents
* [Lexical Structure](#lexical_structure)
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


# <a name="lexical_structure"></a>Lexical Structure
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
	enum	struct 	ret 	const 	true
	false 	match 	while 	for 	do
	if 		else 	unsafe 	fn 		tup
	char	break

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
		ret (a + b);
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
			ret <5, 5, 5>;
		}
		ret <0, 0, 0>;
	}

### <a name="single_line_functions"></a>Single Line Functions
If a function only has one return statement, you may simplify it with the `=>` operator. For
example, a function that adds two integers can be simplified to:

	fn add(int a, int b): int => ret (a + b);

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
underscore `_`. The for loop must also be given a list of 2 - 3 parameters, the start... todo

	for type index: (0, 10, 1) {

	}