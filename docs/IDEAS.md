# Ideas
Please don't take this document seriously, just some random ideas I throw around.

------

## Direct memory allocation
A tilde could perhaps be shorthand for allocating memory and setting it in
one swoop. For example the following:
    
    int ^x = alloc
    ^x = 21

Can be simplfied to:

    int ^x = ~21

Perhaps arrays can be

    int ^a = ~[10, 11, 12, 13]

Which is the same as 

    int ^a = alloc(int * 4)
    a[0] = 10
    a[1] = 11
    a[2] = 12
    a[3] = 13

I may scratch the tilde, however you should be able to explicitly malloc something. The grammar would probably be something like:

	(literal | ["~" literal])

------

## Memory Model
The memory model is pretty simple, the compiler just inserts free's.