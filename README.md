<h1 align="center"><a href="http://alloy-lang.org">The Alloy Programming Language</a></h1>

[![BuildStatus](https://travis-ci.org/alloy-lang/alloy.svg?branch=master)](https://travis-ci.org/alloy-lang/alloy)

## Notice
We're currently porting the backend from C to LLVM, this may take a while and things will be broken. If
you want to checkout the progress, have a look at [issue #345](https://github.com/alloy-lang/alloy/issues/345).
We're also looking to change our name, so if you have any suggestions you would like to make, please head over to [issue #195](https://github.com/alloy-lang/alloy/issues/195) and do so.

## Resources

* [Reference](/docs/REFERENCE.md)
* [Tests](/tests/)
* [Libraries](/lib/)
* [Contributing](/CONTRIBUTING.md)
* [Alloy Style Guide](/docs/STYLEGUIDE.md)

## Example
<p align="center">
<img src="http://alloy-lang.org/example.gif" width="312px" height="312px" />
</p>
<p align="center">
Here's a 3d cube being rendered using SDL and OpenGL in Alloy. You can
check out the source code for this <a href="https://www.github.com/alloy-lang/space-invaders">here</a>
</p>

## Building
If you want to try out Alloy yourself, clone the repository, compile it, and add `bin/alloyc` to your path. You can
also run the test script (you'll need python) `test.py` to see if the tests work, though we can't guarantee they
will all run successfully if you're building the nightly.

### Requirements:
You will need:

* GNU Make
* Clang/GCC (we're not sure about other compilers' support)
* Python 2.4 or above (optional, for running the tests)

### Compiling

	$ git clone http://www.github.com/alloy-lang/alloy
	$ cd alloy
	$ make

We're using LLVM by default now, so you can just run `make` to create the executable.

However, note that the LLVM backend is currently very incomplete, and is only recommended for development purposes. We're working as fast as possible to get it as stable as possible.

### Testing
Note that python is required.
If alloyc is in your `$PATH`:

	$ ./test.py

If alloyc is not in your `$PATH`:

	$ PATH=$PATH:"./bin/" ./test.py

Pass the `--llvm` argument to test.py to use the LLVM backend instead of the C backend.

## Status
Alloy is still constantly being worked on. At the moment it compiles,
however some aspects of the language are unimplemented or broken.

## Just show me what it looks like already
Sure, you can either see a small virtual machine implemented in Alloy [here](tests/vm.aly). 
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
Alloy is licensed under the [MIT License](/LICENSE.md).
