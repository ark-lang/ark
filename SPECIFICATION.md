# JAYFOR SPECIFCATION
We aim to make Jayfor's syntax unique, concise, and easy to get around. We've tried to get rid of some of the annoying flaws [we felt] affect us most commonly while programming. This is a very informal specification of the JAYFOR programming language.

**PLEASE NOTE THAT THIS LANGUAGE IS STILL IN ITS EARLY STAGES, AND ANYTHING IN THIS DOCUMENT IS STILL SUBJECT TO CHANGE. NOTHING IS GUARANTEED AS FINAL.**

## Expressions
Jayfor is still under construction, this affects expressions as they aren't fully implemented yet! An expression like this:

	int x = 5 + 5 + 5 + 5;

is invalid and will not parse! It should be written like this:

	int x = ((5 + 5) + (5 + 5));

We realize that this is a pain, and we will fix this.

## Comments
### Single Line Comments

	// this is a comment, pretty standard

### Block Comments

	/**
	 * These are buggy and we suggest to not use
	 * them for the time being!
	 */

## Data Structures
Structures are declared as follows:

	struct Name {
		int x,
		int y,
		int z
	};

## Functions

Functions are declared like so:

	fn function_name(data_type arg_name, data_type arg_name) [void] { 
		statements;
	}

The above is a function which returns void (nothing).

### Returning Values
We can also return values from functions, like so:

	fn function_name(data_type arg_name, data_type arg_name) [data_type] { 
		statements;
		ret_statement;
	}

For example, an adding function, could be written like so:

	fn add_values(int a, int b) [int] {
		ret (a + b);
	}

### Tuples
Tuples are also supported, allowing you to return multiple values from a function in a (somewhat) array-like fashion:

	fn function_name(param_one, param_two) [data_type, data_type, ...] { 
		statements;
		ret [5, 6];
	}

---------------------------------
## For Loops


We felt that the old style of the **for** loop was annoying as hell. So we added our own little prettifying sauce to it:

	for data_type variable_name:<start, end, step> {
		statements;
	}

**step** is the value by which we must increment **variable_name**
This is equivalent to:
		
	for(data_type variable_name = start; variable_name < end; variable_name += step) { 
		statements;
	}

Jayfor can detect the difference between the start and end value and increment, or decrement accordingly. You do not need to explicitly define
if you want to increment or decrement, but you can if you feel the itching need to, you can prefix your step with a negative sign to decrement.

	for data_type variable_name:<start, end, -step> {
		statements;
	}

You do not not have to specify a step like so:

	for data_type variable_name:<start, end> {
		statements;
	}

If you do not specify step, it will default at 1.

----------------------------------------------
## Do While Loops
We think that Do-while loops are tedious to write and outdated. Having to write a do and then a while condition after is tedious and a waste of time. With Jayfor, you just have to use the do keyword in replace of while, which
means that Jayfor will ensure that the loop is executed atleast once before checking the condition.

	do (condition) { 
		statements;
	}

----------------------------------------------
## If Statement

We think that **if** statements are perfect as they are:

	if (condition) {
		statements;
	}

----------------------------------------------
## While Loop

**While** loops are also perfect as they traditionally are.

	while(condition) {
		statements;
	}

## Pointers
Still not sure if I should implement pointers or not. If I would they would be like this:
	
	int y = 5;
	
	// x is a pointer to the address of y
	int ~x = &y;
	int ~z = x;

	~x = 10;
	// or 
	x -> 10;

----------------------------------------------
## Common Syntax We've Removed
There are some language features we feel are useless, error prone, unsafe, or we just disagree with it.

### If Statements
In most languages, if your if statement only has a single statement in it's block, you do not need braces, i.e:

	if (x == 5)
		printf("whatever\n");

We feel this feature is unsafe and can be prone to errors, so it has not been implemented. Therefore you must use braces, even in single statement ifs, whiles, etc.

	if (x == 5) {
		statement;
	}

### Incrementing and Decrementing
To make our lives a bit easier, we have not implemented the `++` and `--` operators. While it is a nice shorthand, we feel it is not a great priority to implement this feature. If you wish to increment/decrement, you will have to use the older syntax:

	x += 1;
	y -= 1;
	x += 5; // etc

We will update this with newer stuff if our great minds think it up. Stay tuned.
