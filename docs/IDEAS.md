# Ideas
Please don't take this document seriously, just some random ideas I throw around.

------

## Allocation as a compile time operator
If we have allocation as a compile time operator, this means we can infer
the size of the the data type, however we can also have an option on giving
the size of data to allocate. The following example allocates *just* enough
data for the integer.

    int ^x = alloc

We can optionally give it an expression too, this will be how many bytes to allocate.

    int ^x = alloc 5
    int ^y = alloc sizeof int

    // with parenthesis
    int ^z = alloc(5)
    int ^d = alloc(sizeof(int))
    int ^a = alloc(sizeof int)

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

------

## Memory Model
Goals:

* No Garbage Collection