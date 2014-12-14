# <a name="syntax styles"></a>Syntax

We aim to make Jayfor's syntax different while still keeping things concise and easy to get around. We've tried to get rid of some of the annoying flaws [we felt] affect us most commonly while programming. 

##Functions

Functions are declared like so:

	fn function_name(param_one, param_two): [return_type] { 
		statements;
	}

Tuples are also supported, allowing you to return multiple values from a function in a (somewhat) array-like fashion. So you could also do:

	fn function_name(param_one, param_two): [data_type, data_type, ...] { 
		statements;
	}

---------------------------------
##FOR loops


We felt that the old style of the **for** loop was annoying as hell. So we implemented a new way to use them:

	for data_type variable_name:<start, end, step> {
		statements;
	}

**step** is the value by which we must increment **variable_name**
This is equivalent to:
		
	for(data_type variable_name = start; variable_name < end; variable_name += step) { 
		statements;
	}

If **value_one** happens to be greater than **value_two**, Jayfor can detect that and decrement the value by **step** amounts too. There is no need to specify explicitly. However, it *can* be done if you feel the itch/need to.
So:

	for data_type:variable_name:<start, end, -step> {
		statements;
	}

is valid, but only for when **start** is greater than **end**. If you didn't already figure this out, quit programming.

You could also do:

	for data_type variable_name:<start, end>
		
in which case **step** defaults to 1, and the increments happen in steps of 1 (obviously).

----------------------------------------------
##DO-WHILE loops


Do-while loops are tedious to write. Having to write a do and then a while condition after it is a complete waste of time in our opinion. So we have a new way to declare do-while loops:

	do(condition) { 
		statements;
	}

If you think this is worse than:

	do {
		statements;
	} while(condition);

then you're retarded.

----------------------------------------------
##IF loop

**IF** loops are the same, since they don't really need any change.

	if(condition) {
		statements;
	}

----------------------------------------------
##WHILE loop

**WHILE** loops are the same too, since, again, they don't really need to be changed.

	while(condition) {
		statements;
	}

----------------------------------------------
##Other general stuff

Most of the other syntax is pretty much the same. However, we despise the *variable++* way of incrementing (or decrementing). So please do:

	variable_to_be_incremented += number;

or:

	variable_to_be_decremented -= number;

if you need to increment or decrement by a certain value.
	

We will update this with newer stuff if our great minds think it up. Stay tuned.
