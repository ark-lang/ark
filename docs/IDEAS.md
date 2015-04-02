# Polymorphism

# Macro System

# Data Types

	int -> 32 bit 
	float -> 32 bit

	u8, u16, u32, u64
	i8, i16, i32, i64

	byte -> u8
	bool -> i8

# Option Types
No NULL's, option types are better.

	func doStuff(int a, int b): <int> {
		return Some(a / b);
	}

# Tuples

	func readFile(string path): <string> {
		string fileContents = ...
		if (success) {
			return Some(fileContents);
		}
		return None;
	}

# Void shorthand?

	int z = 0;

	func add(int a, int b) { // no colon = void?
		z = a + b;
	}

# Cleaner memory allocations

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
