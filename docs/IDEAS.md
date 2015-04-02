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
This is tricky. A C++ `new` and `delete` would be nice, seems to object oriented,
like I'm calling a constructor:

	Entity ^e = new Entity;

Perhaps 