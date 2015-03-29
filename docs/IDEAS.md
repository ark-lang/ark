# Ideas
Please don't take this document seriously, just some random ideas I throw around.

------

## C Bindings
A c binding is created as follows:

    binding {
        'file.h' -> fn add(int a, int b): int;
    }

    And then it can be used in your code i.e: 

        int x = add(5, 5);

------

## Memory Management
I could implement a static analyzer for memory? For instance, the compiler would basically do
ARC **once** before code generation. So it would work as follows:

    int ^x = malloc(sizeof(^x));        // alocate some memory
                                        // the compiler will take this and create a reference to it

    ^x = 10;

    // it has zero references so the compiler inserts free(x) at the end of the scope.

Now a more complex example:

    struct Dog {
        int y;
    }

    struct Cat {
        int x;
        Dog ^dog;
    }

    fn swag(Cat ^cat): void {
        cat.dog = malloc(sizeof(^cat.dog));
        cat.dog.y = 10;
    }

    fn main(): void {
        Cat ^cat = malloc(sizeof(^cat));
        cat.x = 10;
        swag(cat);
    }

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
