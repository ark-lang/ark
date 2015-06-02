# <span align="center">Ark</span>

## Important Notice!!
We've ported to **Go**, we've still got the original C code, which you
can find [here](https://github.com/ark-lang/ark-c)

Due to the name change, we've migrated the subreddit from r/alloy_lang to [r/ark_lang](http://www.reddit.com/r/ark_lang), go subscribe! Finally, our irc has moved from #alloy-lang to #ark-lang

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

## `make gen` and `make fmt`
The target `gen` is included for the convenience of the developers. It runs `go generate` on all the modules in ark.

### Testing
If Ark is in your `$PATH`:

    $ ./test.py

If Ark is not in your `$PATH`:

    $ PATH=$PATH:"./bin/" ./test.py

## License
Ark is licensed under the [MIT License](/LICENSE).