# Polymorphism
TODO, some kind of generics would be nice.

# Macro System
TODO

# Data Types

	int -> 32/64 bit (architechture dependant)
	float -> 32/64 bit (architechture dependant)

	u8, u16, u32, u64 -> unsigned, X bits, not guaranteed

	i8, i16, i32, i64 -> signed, X bits, not guaranteed

	byte -> u8
	bool -> i8

# Option Types
No NULL's, option types are better.

	func doStuff(int a, int b): <int> {
		return Some(a / b);
	}

# Tuples

	func swap(int a, int b): (int, int) {
		return (b, a);
	}

	(int, int) x = swap(5, 10);
	// something like that for access
	printf("%d %d\n", x.0, x.1);

# Void shorthand?

	int z = 0;

	func add(int a, int b) { // no colon = void?
		z = a + b;
	}

# Cleaner memory allocations
This is really weird

	struct Entity {
		int x = 0;
		int y = 0;
		int time_alive = 0;
		bool alive = true;
	}

	// constructor-y thing?
	// if you provide params
	// it will set x from struct to
	// x given, names must match
	// the struct member
	// can change self to this, entity, e, etc..
	impl Entity(int x, int y) as self {

		func render(): void {
			gl.fillRect(self.x, self.y);
		}

		func update(): void {
			self.time_alive += 1;
			if self.time_alive > 2000 {
				self.alive = false;
			}
		}
	}

	// "constructor" call, not directly
	// setting struct values.
	// if it's a pointer, will malloc memory
	// if passed to function expecting pointer, will malloc memory
	Entity ^e = Entity(10, 10);

	loop {
		e.render();
		e.update();
		if !e.alive {
			break;
		}
	}

	some_vector.push_back(Entity(10, 10));

## Match
Match is better and safer than a switch

	switch (x) {
		case 0:
			printf("hey\n"); // no break! common pit fall
		case 1:
			printf("hi\n");
			break;
	}

	// doesn't fall through
	match x {
		0 -> printf("hey\n");
		1 -> printf("hi\n");
	}

	// what if we want to "emulate" a fall through?
	// it's cleaner!
	// defaults to break
	// you can specify continue or return, but only
	// through a block, instead of a ->

	match x {
		0, 1 -> printf("hey hi\n");
		2 {
			printf("more hi\n");
			printf("wow!\n");
			return 5;
		}
	}

## Enforcing braces
Error in C

	if (x == 1)
		if (y == 2)
			printf("hi");
	else
		printf("swag");

Is evaluated as:

	if (x == 1)
		if (y == 2)
			printf("hi");
		else
			printf("swag");

Fixed with a block:

	if (x == 1) {
		if (y == 2)
			printf("hi");
	}
	else 
		printf("swag");

In alloy, we enforce braces:

	if x == 1 {
		if y == 2 {

		}
	}
	else {

	}

# Loops

	// while loop
	for xyz {
		...
	}

	// infinite loop
	for {
		...
	}

	// traditional loop
	// inferred type?
	for i = 0; i < 10; i++ {
		...
	}

