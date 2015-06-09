<h1 align="center">Ark Programming Language</h1>

## Important Notice!!
We've ported to *Go*. We've still got the original C code, which you can find [here](https://github.com/ark-lang/ark-c).

Due to the name change, we've migrated the subreddit from /r/alloy_lang to [/r/ark_lang](http://www.reddit.com/r/ark_lang), go subscribe! Finally, our IRC channel has moved from #alloy-lang to [#ark-lang](https://webchat.freenode.net/?channels=%23ark-lang) on freenode.

## Resources
* [Reference](https://github.com/ark-lang/ark-docs/blob/master/REFERENCE.md)
* [Tests](/tests/)
* [Libraries](/lib/)
* [Contributing](/CONTRIBUTING.md)
* [Ark Style Guide](https://github.com/ark-lang/ark-docs/blob/master/STYLEGUIDE.md)

## Dependencies
* Go installed and `$GOPATH` setup
* subversion
* LLVM installed, with `llvm-config` and `llc` in your `$PATH`
* a C++ compiler

## Installing

Replace `RELEASE_360` with the version of LLVM you have installed. For example, version 3.6.1 becomes `RELEASE_361`. You can find out your version of llvm by running `llvm-config --version`.

    $ svn co https://llvm.org/svn/llvm-project/llvm/tags/RELEASE_360/final $GOPATH/src/llvm.org/llvm
    $ export CGO_CPPFLAGS="`llvm-config --cppflags`"
    $ export CGO_LDFLAGS="`llvm-config --ldflags --libs --system-libs all`"
    $ export CGO_CXXFLAGS=-std=c++11
    $ go install -tags byollvm llvm.org/llvm/bindings/go/llvm
    $ go get github.com/ark-lang/ark
    $ go install github.com/ark-lang/ark

Make sure `$GOPATH/bin` is in your `$PATH`.

## Usage
For detailed usage information, run `ark help`. For information on specific commands, use `ark help <command>`.

### Building
```
ark build tests/big_test.ark -o out_name
```
If the `-o` option is not specified, the outputted binary will be names `main`.

### Docgen
```
ark docgen tests/big_test.ark --dir some_output_dir
```
If the `--dir` option is not specified, the generated documentation will be placed in `docgen`.

## `make gen` and `make fmt`
The targets `gen` and `fmt` are included for the convenience of the developers. They run `go generate` and `go fmt` respectively on all the modules in Ark. Please run `make fmt` before creating a pull request.

### Testing
Requires `$GOPATH/bin` to be in your `$PATH`.

    $ ./test.py

## License
Ark is licensed under the [MIT License](/LICENSE).
