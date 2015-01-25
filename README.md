[Alloy](http://alloy-lang.github.io) [![BuildStatus](https://travis-ci.org/alloy-lang/alloy.svg?branch=master)](https://travis-ci.org/alloy-lang/alloy)
===

Alloy is a programming language written in C. The goals of this language are:

* Speed
* Simplicity
* Cleanliness

We created Alloy because we love writing C, and love the simplicity of C; we also wanted to evolve C
into something more modern, and easier to use.

IRC
------
We have an IRC where we discuss Alloy, and other stuff too, come join! If you want to help contribute,
we highly suggest you join the IRC :)

* server: irc.freenode.net
* channel: #alloy-lang

A taste of Alloy
------
Warning, this is still (potentially) subject to change. Also a disclaimer,
the syntax below is very incosistent and is just to show what's possible in the 
language.
```rust
struct Whatever {
	int x = 10
	int y
}

fn do_stuff(int a, int b): int {
	int x = 5
	while ((a + b) > 10) -> x = (x + 1)

	match x {
		(x == 12) {
			loop {
				if ((a + x) > 21) -> break
			}
			return x
		}
		(x < 2) {
			for _ :(a, b) {
				a = (x + 1)
			}
		}
		(x == 29) -> return a
		_ {
			return 1337
		}
	}

	if (x > 21) {
		return (x - 21)
	}
}

fn main() {
	do_stuff(125, 25)
}
```
Status
------

The language is still in development. More information/specifics
can be found in [`/docs/`](/docs/). The compiler is written in C
and uses LLVM as its backend.

Building
--------

To build you will need:

 - [`LLVM >= 3.4`](http://llvm.org/releases/download.html)
 - A suitable GNU C compiler (any one of the below will do fine):
   - [`clang >= 3.4.0`](http://llvm.org/releases/download.html)
   - [`gcc >= 4.8.1`](https://gcc.gnu.org/) (change `LCC` and 
     `LCXX` in the [makefile](/Makefile) to `gcc` and `g++`, respectively)

```bash
# clone the repository
git clone -v https://github.com/alloy-lang/alloy.git
    
# cd into the repository
cd alloy

# build
make

# run a program
./alloyc tests/parser-tests/const.alloy
```

Contributing
------------

Something not working? Open an [Issue](https://github.com/alloy-lang/alloy/issues)
or send us a [Pull Request](https://github.com/alloy-lang/alloy/pulls)
on GitHub.

See [CONTRIBUTING.md](/CONTRIBUTING.md).

License
-------

Alloy is licensed under the [MIT License](/LICENSE.md).
