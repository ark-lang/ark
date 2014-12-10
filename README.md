# Jayfor
Jayfor is a programming language written in C.

# License
Jayfor is licensed under The MIT License. I have no idea
what this means, I just randomly chose it. Read it [here](LICENSE.md)

# Building
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

# Goals
* Self hosting
* No Garbage Collection
* Simplicity
* Safety
* Helpful Error Messages
