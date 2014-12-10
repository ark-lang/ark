# Jayfor
Jayfor is a programming language written in C.

# Table of Contents
* [About](#about)
* [License](#license)
* [Building](#building)
* [Goals](#goals)

# <a name="about"></a>About
Jayfor is a programming language written in C. It is still under
heavy development. Jayfor is influenced from languages like Rust,
C, C++, and Java. We like to keep things simple, safe and fast,
but maintain a syntactically beautiful language.

## Sample
You can view the snippet of code below for a basic 'feel' of the language,
or check out some actual code we use for the library [here](libs/math.j4).

	# add the given values together
	fn add(int a, int b): int {
		ret a + b;
	}

	fn main(int argc, str[] argv): void {
		int a = add(5, 3);
		println(a);
	}

# <a name="license"></a>License
Jayfor is licensed under The MIT License. I have no idea
what this means, I just randomly chose it. Read it [here](LICENSE.md)

# <a name="building"></a>Building
Building is very simple

## Requirements
Only GCC or Clang is required to build the repository. You
can also use another compiler, but you will have to change
the [Makefile](Makefile).

## Clone the repository

	git clone git@github.com:freefouran/jayfor.git
	cd jayfor

## Building with Make
If you are using MinGW or GCC:

	make j4mingw

If you are using Clang

	make

This will create an executable, which can be run like
so:

	./j4 somefile.j4

# <a name="goals"></a>Goals
* Self hosting
* No Garbage Collection
* Simplicity
* Safety
* Helpful Error Messages
