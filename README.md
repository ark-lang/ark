## Ark Programming Language
[![Build Status](https://travis-ci.org/ark-lang/ark.png?branch=master)][1]

[Ark](//www.ark-lang.org) is a systems programming language somewhere inbetween C and C++.

## Example
A small example program written in Ark.

```rust
[c] func printf(fmt: str, ...);
func main(): int {
    mut i := 0;
    for i < 5 {
        C::printf("i: %d\n", i);
        i = i + 1;
    }
    return 0;
}
```

## Resources
* [Reference](https://github.com/ark-lang/ark-docs/blob/master/REFERENCE.md)
* [Tests](/tests/)
* [Libraries](/lib/)
* [Contributing](/CONTRIBUTING.md)
* [Ark Style Guide](https://github.com/ark-lang/ark-docs/blob/master/STYLEGUIDE.md)

## Installing
Installing Ark should be relatively easy, you'll need a few dependencies before
you get started:

### Dependencies
* Go installed and `$GOPATH` setup
* subversion
* LLVM installed, with `llvm-config` and `llc` in your `$PATH`
* a C++ compiler

## Building
Replace `RELEASE_360` with the version of LLVM you have installed. For example, version 3.6.1 becomes `RELEASE_361`. You can find out your version of llvm by running `llvm-config --version`.

    $ svn co https://llvm.org/svn/llvm-project/llvm/tags/RELEASE_360/final $GOPATH/src/llvm.org/llvm
    $ export CGO_CPPFLAGS="`llvm-config --cppflags`"
    $ export CGO_LDFLAGS="`llvm-config --ldflags --libs --system-libs all`"
    $ export CGO_CXXFLAGS=-std=c++11
    $ go install -tags byollvm llvm.org/llvm/bindings/go/llvm
    $ go get github.com/ark-lang/ark
    $ go install github.com/ark-lang/ark

Make sure `$GOPATH/bin` is in your `$PATH`.

To see the current state of the compiler, try running the [test script](#testing).

## Usage
For detailed usage information, run `ark help`. For information on specific commands, use `ark help <command>`.

### Compiling Ark Code
```
ark build tests/big_test.ark -o out_name
```
If the `-o` option is not specified, the outputted binary will be names `main`.

### Docgen
```
ark docgen tests/big_test.ark --dir some_output_dir
```
If the `--dir` option is not specified, the generated documentation will be placed in `docgen`.

## Utilities
### `make gen` and `make fmt`
The targets `gen` and `fmt` are included for the convenience of the developers. They run `go generate` and `go fmt` respectively on all the modules in Ark. Please run `make fmt` before creating a pull request.

### Testing
Requires `$GOPATH/bin` to be in your `$PATH` and Python 2.4 or newer.

    $ ./test.py

## License
Ark is licensed under the [MIT License](/LICENSE).

[1]: https://travis-ci.org/ark-lang/ark "Build Status"
