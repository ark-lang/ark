# JAYFOR SPECIFCATION
We aim to make Jayfor's syntax unique, concise, and easy to get around. We've tried to get rid of some of the annoying flaws [we felt] affect us most commonly while programming. 

## Functions

Functions are declared like so:

	fn function_name(data_type arg_name, data_type arg_name): void { 
		statements;
	}

The above is a function which returns void (nothing).

### Returning Values
We can also return values from functions, like so:

	fn function_name(data_type arg_name, data_type arg_name): void { 
		statements;
	}

### Tuples
Tuples are also supported, allowing you to return multiple values from a function in a (somewhat) array-like fashion. So you could also do:

	fn function_name(param_one, param_two): [data_type, data_type, ...] { 
		statements;
	}

---------------------------------
##FOR loops


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
## DO-WHILE loops
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

----------------------------------------------
## Misc

Most of the other syntax is pretty much the same. However, we despise the *variable++* way of incrementing (or decrementing). So please do:

	variable_to_be_incremented += number;

or:

	variable_to_be_decremented -= number;

if you need to increment or decrement by a certain value.
	

We will update this with newer stuff if our great minds think it up. Stay tuned.
