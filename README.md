<h1 align="center"><a href="http://ark-lang.org">The Ark Programming Language</a></h1>


## Notice
**For the final time, we've changed our name. We're the *Ark* programming language now. We had to get rid of Alloy due to some complications, and finalized on Ark. The entire discussion (which is rather long for a name change) can be found on [issue #195](https://github.com/ark-lang/ark/issues/195).**

## Notice 2
We're also currently porting the backend from C to LLVM, this may take a while and things will be broken. If
you want to checkout the progress, have a look at [issue #345](https://github.com/ark-lang/ark/issues/345).


## Resources

* [Reference](/docs/REFERENCE.md)
* [Tests](/tests/)
* [Libraries](/lib/)
* [Contributing](/CONTRIBUTING.md)
* [Ark Style Guide](/docs/STYLEGUIDE.md)

## Building
If you want to try out Ark yourself, clone the repository, compile it, and add `bin/arkc` to your path. You can
also run the test script (you'll need python) `test.py` to see if the tests work, though we can't guarantee they
will all run successfully if you're building the nightly.

### Requirements:
You will need:

* GNU Make
* Clang/GCC (we're not sure about other compilers' support)
* Python 2.4 or above (optional, for running the tests)

### Compiling

	$ git clone http://www.github.com/ark-lang/ark
	$ cd ark
	$ make

We're using LLVM by default now, so you can just run `make` to create the executable.

However, note that the LLVM backend is currently very incomplete, and is only recommended for development purposes. We're working as fast as possible to get it as stable as possible.

### Testing
Note that python is required.
If arkc is in your `$PATH`:

	$ ./test.py

If arkc is not in your `$PATH`:

	$ PATH=$PATH:"./bin/" ./test.py

## Status
Ark is still constantly being worked on. At the moment it compiles,
however some aspects of the language are unimplemented or broken.

## Just show me what it looks like already
Sure, you can either see a small virtual machine implemented in Ark [here](tests/vm.aly). 
Or you can just see a small Hello World example below:

```rust
// c binding for printf
// you have to do this for now, but soon
// we'll have a standard library for this
func printf(format: str, _): int;

func main(): int {
    printf("Hello, World");
    return 0;
}
```

## License
Ark is licensed under the [MIT License](/LICENSE.md).
