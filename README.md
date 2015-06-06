<h1 align="center">Ark</h1>

## Important Notice!!
We've ported to **Go**. We've still got the original C code, which you can find [here](https://github.com/ark-lang/ark-c).

Due to the name change, we've migrated the subreddit from /r/alloy_lang to [/r/ark_lang](http://www.reddit.com/r/ark_lang), go subscribe! Finally, our IRC channel has moved from #alloy-lang to [#ark-lang](https://webchat.freenode.net/?channels=%23ark-lang) on freenode.

## Resources
* [Reference](https://github.com/ark-lang/ark-docs/blob/master/REFERENCE.md)
* [Tests](/tests/)
* [Libraries](/lib/)
* [Contributing](/CONTRIBUTING.md)
* [Ark Style Guide](https://github.com/ark-lang/ark-docs/blob/master/STYLEGUIDE.md)

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

## Usage
For detailed usage information, run `ark help`. For information on specific commands, use `ark help <command>`.

### Building
```
ark build tests/big_test.ark
```

##& Docgen
```
ark docgen tests/big_test.ark --dir some_output_dir
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
