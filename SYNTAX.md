# <a name="syntax styles"></a>Syntax

We aim to make Jayfor different while still keeping things concise and easy to get around. We've tried to get rid of some of the annoying flaws [we felt] affect us most commonly while coding. 

##Functions

Functions are delared like so:

		fn function_name(param_one, param_two): [return_type] { }

We've added the tuples functionality, allowing you to return multiple values from a function in a (somewhat) array-like fashion. So you could also do:

		fn function_name(param_one, param_two): [tuple_value_one, tuple_value_two, ...] { }

---------------------------------
##FOR loops


We felt that the old style of the *for* loop was annoying as hell. So we implemented a new way to use them:

		for data_type variable_name:<start, end, step>

*step* is the value by which we must increment *variable_name*
This is equivalent to:
		
		for(data_type variable_name = start; variable_name < end; variable_name += step) { }

If *value_one* happens to be greater than *value_two*, Jayfor can detect that and decrement the value by *step* amounts too. There is no need to specify explicitly.

You could also do:

		for data_type variable_name:<start, end>
		
in which case *step* defaults to 1, and the increments happen in steps of 1 (obviously).

----------------------------------------------
##DO-WHILE loops


Do-while loops are tedious to write. Having to write a do and then a while condition after it is a complete waste of time in our opinion. So we have a new way to declare do-while loops:

		do(condition) { }

If you think this is worse than:

		do {
		
		} while(condition);

then you're retarded.

We will update this with newer stuff if our great minds think it up. Stay tuned.
