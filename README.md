## Ark [![Build Status](https://api.travis-ci.org/ark-lang/ark.svg?branch=master)][1] [![license](http://img.shields.io/badge/license-MIT-brightgreen.svg)](https://raw.githubusercontent.com/ark-lang/ark/master/LICENSE)
[1]: https://travis-ci.org/ark-lang/ark "Build Status"

[Ark](//www.ark-lang.org) is a systems programming language is a systems
programming language somewhere inbetween C and C++. It's goals are to
modernize the C language, yet remove the cruft that is present in C++
due to backwards compatibility.

## Index
* [Getting Involved](#getting-involved)
* [Example](#example)
* [Installing](#installing)
    * [Dependencies](#dependencies)
    * [Building](#building)

## <a name="getting-involed"></a> Getting Involved
Check out the [contributing guide](/CONTRIBUTING.md), there's a lot of information
there to give you ideas of how you can help out.

## <a name="example"></a> Example
Ark is still a work in progress, this code sample reflects what Ark can
do *currently*, though the way you write the following will likely change
in the near future.

More examples can be found [here](/examples).

```rust
// binding to printf
[c] func printf(fmt: str, ...);

func main(args: []str) -> int {
    // mutable i, type inferred
    mut i := 0;

    // #args gets the length of an array
    for i < #args {
        // accessed via the C module
        C::printf("%s\n", args[i]);

        i += 1;
    }
    return 0;
}
```

## <a name="installing"></a> Installing
Installing Ark is simple, you'll need a few dependencies 
before you get started:

### <a name="dependencies"></a> Dependencies
* Go installed and `$GOPATH` setup - [Instructions on setting up GOPATH](//golang.org/doc/code.html#GOPATH)
* subversion
* LLVM installed, with `llvm-config` and `llc` in your `$PATH`
* a C++ compiler
* `libedit-dev` installed

### <a name="building"></a> Building
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
