# ark-go
Experimental Ark compiler in Go.

* [#ark-lang](http://webchat.freenode.net/?channels=%23ark-lang) on freenode
* [Language reference](https://github.com/ark-lang/ark-docs/blob/master/REFERENCE.md)
* [Ark Style Guide](https://github.com/ark-lang/ark-docs/blob/master/STYLEGUIDE.md)

	Usage of ark-go:
	  -input:   input file
	  -output:  output file
	  -v:       enable verbose mode
	  -version: show version information

## Installing
Requires Go to be installed and $GOPATH setup.

Building LLVM bindings (must be done first and may take a while):

	$ go get -d llvm.org/llvm/bindings/go/llvm
	$ cd $GOPATH/src/llvm.org/llvm/bindings/go/
	$ ./build.sh
	$ go install llvm.org/llvm/bindings/go/llvm

Building ark-go:

	$ go get github.com/ark-lang/ark-go
	$ go install github.com/ark-lang/ark-go
	$ ark-go

Make sure `$GOPATH/bin` is in your `$PATH`.

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
