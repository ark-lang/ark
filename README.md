# alloy-go
Experimental Alloy compiler in Go

You cand find the language reference and other stuff over [here](https://github.com/alloy-lang/alloy).

## Installing
Requires Go to be installed and $GOPATH setup.

Building LLVM bindings (must be done first and may take a while):

	$ go get -d llvm.org/llvm/bindings/go/llvm
	$ cd $GOPATH/src/llvm.org/llvm/bindings/go/
	$ ./build.sh
	$ go install llvm.org/llvm/bindings/go/llvm

Building alloy-go:

	$ go get github.com/alloy-lang/alloy-go
	$ go install github.com/alloy-lang/alloy-go
	$ alloy-go

Make sure `$GOPATH/bin` is in your `$PATH`.

## Styleguide
* Use tabs for indenting, spaces for alignment.
* Use `v` as the name of the struct a method is called on
* Use American English (unfortunately)

### Abbreviations
* Statement -> Stat
* Expression -> Expr
* Declaration -> Decl
