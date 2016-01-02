## Ark [![Build Status](https://api.travis-ci.org/ark-lang/ark.svg?branch=master)][1] [![license](http://img.shields.io/badge/license-MIT-brightgreen.svg)](https://raw.githubusercontent.com/ark-lang/ark/master/LICENSE)

[Ark](//www.ark-lang.org) is a systems programming language somewhere in-between C and C++.

## Index
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
* [Code of Conduct](#coc)

## <a name="example"></a> Example
For a more complicated example, check out a port of my virtual machine MAC in Ark
[here](//www.github.com/ark-lang/mac-ark). Or if you just want a small example 
program written in Ark.

```rust
[c] func printf(fmt: str, ...);
func main() -> int {
    mut i := 0;
    for i < 5 {
        C::printf("i: %d\n", i);
        i += 1;
    }
    return 0;
}
```

## <a name="resources"></a> Resources
* [#ark-lang](//webchat.freenode.net/?channels=%23ark-lang)
* [Reference Book (WIP)](http://felixangell.gitbooks.io/ark-reference/content/)
* [Reference](//github.com/ark-lang/ark-docs/blob/master/REFERENCE.md)
* [Contributing](/CONTRIBUTING.md)
* [Ark Style Guide](//github.com/ark-lang/ark-docs/blob/master/STYLEGUIDE.md)
* [Tests](/tests/)
* [Libraries (WIP)](/lib/)

## <a name="installing"></a> Installing
Installing Ark is simple, you'll need a few dependencies 
before you get started:

### <a name="dependencies"></a> Dependencies
* Go installed and `$GOPATH` setup - [Instructions on setting up GOPATH](//golang.org/doc/code.html#GOPATH)
* subversion
* LLVM installed, with `llvm-config` and `llc` in your `$PATH`
* a C++ compiler
* `libedit-dev` installed

## <a name="building"></a> Building
Replace `release` to match your llvm release. You can check by running 
the `llvm-config --version` command. If you are on 3.6.1, `release` would
become `RELEASE_361`, or `RELEASE_362` for 3.6.2, and so on.

```bash
$ release=RELEASE_362 # set this to match your llvm-config --version
$ svn co https://llvm.org/svn/llvm-project/llvm/tags/$release/final $GOPATH/src/llvm.org/llvm
$ cd $GOPATH/src/llvm.org/llvm/bindings/go
$ ./build.sh
$ go install llvm.org/llvm/bindings/go/llvm
$ go get github.com/ark-lang/ark/...
```

The `ark` binary will be built in `$GOPATH/bin`. To use the compiler, 
make sure `$GOPATH/bin` is in your `$PATH`.

## <a name="usage"></a> Usage
For detailed usage information, run `ark help`. For information
on specific commands, use `ark help <command>`.

### <a name="compiling-ark-code"></a> Compiling Ark Code
To compile ark code, pass a module to the executable
sub-command `build`:

```bash
$ ark build tests/big_test.ark -o out_name
```

_If the `-o` option is not specified, the binary name will default to `main`._

## <a name="utilities"></a> Utilities
### <a name="make-gen-and-make-fmt"></a> `make gen` and `make fmt`
The targets `gen` and `fmt` are included for the convenience of the developers. 
They run `go generate` and `go fmt` respectively on all the modules in Ark. 
Please run `make fmt` before creating a pull request.

### <a name="testing"></a> Testing
Requires `$GOPATH/bin` to be in your `$PATH` and Python 2.4 or newer.

```bash
$ ./test.py
```

[1]: https://travis-ci.org/ark-lang/ark "Build Status"
