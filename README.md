## Ark [![Build Status](https://travis-ci.org/ark-lang/ark.png?branch=master)][1] [![license](http://img.shields.io/badge/license-MIT-green.svg)](https://raw.githubusercontent.com/ark-lang/ark/master/LICENSE) [![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/ark-lang/ark?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

[Ark](//www.ark-lang.org) is a systems programming language somewhere in-between C and C++.

## Table Of Contents
* [Example](#example)
* [Resources](#resources)
* [Installing](#installing)
    * [Dependencies](#dependencies)
* [Building](#building)
* [Usage](#usage)
    * [Compiling Ark Code](#compiling-ark-code)
    * [Generating Documentation](#docgen)
* [Utilities](#utilities)
    * [`make gen` and `make fmt`](#make-gen-and-make-fmt)
    * [Testing](#testing)
* [License](#license)

## <a name="example"></a> Example
For a more complicated example, check out a port of my virtual machine MAC in Ark [here](//www.github.com/ark-lang/mac-ark). Or if you just want a small example program written in Ark.

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

## <a name="resources"></a> Resources
* [Reference](//github.com/ark-lang/ark-docs/blob/master/REFERENCE.md)
* [Tests](/tests/)
* [Libraries](/lib/)
* [Contributing](/CONTRIBUTING.md)
* [Ark Style Guide](//github.com/ark-lang/ark-docs/blob/master/STYLEGUIDE.md)

## <a name="installing"></a> Installing
Installing Ark should be relatively easy, you'll need a few dependencies before
you get started:

### <a name="dependencies"></a> Dependencies
* Go installed and `$GOPATH` setup - [Instructions on setting up GOPATH](//golang.org/doc/code.html#GOPATH)
* subversion
* LLVM installed, with `llvm-config` and `llc` in your `$PATH`
* a C++ compiler

## <a name="building"></a> Building
Replace `RELEASE_360` with the version of LLVM you have installed. For example, version 3.6.1 becomes `RELEASE_361`. You can find out your version of llvm by running `llvm-config --version`.

    $ svn co https://llvm.org/svn/llvm-project/llvm/tags/RELEASE_360/final \
        $GOPATH/src/llvm.org/llvm
    $ export CGO_CPPFLAGS="`llvm-config --cppflags`"
    $ export CGO_LDFLAGS="`llvm-config --ldflags --libs --system-libs all`"
    $ export CGO_CXXFLAGS=-std=c++11
    $ go install -tags byollvm llvm.org/llvm/bindings/go/llvm
    $ go get github.com/ark-lang/ark/...
    $ cd $GOPATH/github.com/ark-lang/ark
    $ make

The `ark` binary will be built in `$GOPATH/bin`. To use the compiler, make sure `$GOPATH/bin` is in your `$PATH`.

To see the current state of the compiler, try running the [test script](#testing).

## <a name="usage"></a> Usage
For detailed usage information, run `ark help`. For information on specific commands, use `ark help <command>`.

### <a name="compiling-ark-code"></a> Compiling Ark Code
```
ark build tests/big_test.ark -o out_name
```
If the `-o` option is not specified, the outputted binary will be names `main`.

### <a name="docgen"></a> Docgen
```
ark docgen tests/big_test.ark --dir some_output_dir
```
If the `--dir` option is not specified, the generated documentation will be placed in `docgen`.

## <a name="utilities"></a> Utilities
### <a name="make-gen-and-make-fmt"></a> `make gen` and `make fmt`
The targets `gen` and `fmt` are included for the convenience of the developers. They run `go generate` and `go fmt` respectively on all the modules in Ark. Please run `make fmt` before creating a pull request.

### <a name="testing"></a> Testing
Requires `$GOPATH/bin` to be in your `$PATH` and Python 2.4 or newer.

    $ ./test.py

[1]: https://travis-ci.org/ark-lang/ark "Build Status"
