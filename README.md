#<a href="http://ark-lang.org">The Ark Programming Language</a>

## Notices
**For the final time, we've changed our name. We're the *Ark* programming language now. We had to get rid of Alloy due to some complications, and finalized on Ark. The entire discussion (which is rather long for a name change) can be found on [issue #195](https://github.com/ark-lang/ark/issues/195).**

**Due to this name change, we've also migrated the subreddit from r/alloy_lang to [r/ark_lang](http://www.reddit.com/r/ark_lang), go subscribe! Finally, our irc has moved from #alloy-lang to #ark-lang**

We're also currently porting the backend from C to **LLVM**, this may take a while and things will be broken. If
you want to check out the progress, have a look at [issue #345](https://github.com/ark-lang/ark/issues/345).

## Resources
* [Reference](/docs/REFERENCE.md)
* [Tests](/tests/)
* [Libraries](/lib/)
* [Contributing](/CONTRIBUTING.md)
* [Ark Style Guide](/docs/STYLEGUIDE.md)

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

## Styleguide
* Use tabs for indenting, spaces for alignment
* Use `v` as the name of the struct a method is called on
* Use American English (unfortunately)

### Abbreviations
* Statement -> Stat
* Expression -> Expr
* Declaration -> Decl

## `make gen` and `make fmt`
The target `gen` is included for the convenience of the developers. It runs `go generate` on all the modules in alloy-go.

### Testing
If Ark is in your `$PATH`:

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