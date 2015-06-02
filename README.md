<<<<<<< HEAD
<h1 align="center">Ark</h1>

## Important Notice!!
We've ported to **Go**, we've still got the original C code, which you
can find [here](https://github.com/ark-lang/ark-c)

Due to the name change, we've migrated the subreddit from r/alloy_lang to [r/ark_lang](http://www.reddit.com/r/ark_lang), go subscribe! Finally, our irc has moved from #alloy-lang to #ark-lang
=======
#<a href="http://ark-lang.org">The Ark Programming Language</a>

## Notices
**For the final time, we've changed our name. We're the *Ark* programming language now. We had to get rid of Alloy due to some complications, and finalized on Ark. The entire discussion (which is rather long for a name change) can be found on [issue #195](https://github.com/ark-lang/ark/issues/195).**

**Due to this name change, we've also migrated the subreddit from r/alloy_lang to [r/ark_lang](http://www.reddit.com/r/ark_lang), go subscribe! Finally, our irc has moved from #alloy-lang to #ark-lang**

We're also currently porting the backend from C to **LLVM**, this may take a while and things will be broken. If
you want to check out the progress, have a look at [issue #345](https://github.com/ark-lang/ark/issues/345).
>>>>>>> 7b809376e8b79ef131aaa19be5fa091809348bd4

## Resources
* [Reference](/docs/REFERENCE.md)
* [Tests](/tests/)
* [Libraries](/lib/)
* [Contributing](/CONTRIBUTING.md)
* [Ark Style Guide](/docs/STYLEGUIDE.md)

<<<<<<< HEAD
## Installing
Requires Go to be installed and $GOPATH setup.

Building LLVM bindings (must be done first and may take a while):

    $ go get -d llvm.org/llvm/bindings/go/llvm
    $ cd $GOPATH/src/llvm.org/llvm/bindings/go/
    $ ./build.sh
    $ go install llvm.org/llvm/bindings/go/llvm

Building ark:

    $ go get github.com/ark-lang/ark
    $ go install github.com/ark-lang/ark
    $ ark

Make sure `$GOPATH/bin` is in your `$PATH`.

```
Usage of ark:
  -v:                 enable verbose mode
  --output:           output file
  --version:          show version information
  --codegen=<backend> sets the codegen to <backend> (none, llvm or ark. default: none)
```

## `make gen` and `make fmt`
The target `gen` is included for the convenience of the developers. It runs `go generate` on all the modules in ark.
=======
## Building
If you want to try out Ark yourself, clone the repository, compile it, and add `bin/ark` to your path. You can
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

However, note that the LLVM backend is currently very incomplete, and is only recommended for development purposes. 
We're working as fast as possible to get it as stable as possible.
>>>>>>> 7b809376e8b79ef131aaa19be5fa091809348bd4

### Testing
If Ark is in your `$PATH`:

<<<<<<< HEAD
    $ ./test.py

If Ark is not in your `$PATH`:

    $ PATH=$PATH:"./bin/" ./test.py

## License
Ark is licensed under the [MIT License](/LICENSE).
=======
	$ ./test.py

If Ark is not in your `$PATH`:

	$ PATH=$PATH:"./bin/" ./test.py

## Just show me what it looks like already
Sure, you can either see a small virtual machine implemented in Ark [here](tests/old_tests/vm.ark). 
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
>>>>>>> 7b809376e8b79ef131aaa19be5fa091809348bd4
