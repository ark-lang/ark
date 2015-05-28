# ark-go
Experimental Ark compiler in Go.

You can find the language reference and other stuff over [here](https://github.com/ark-lang/ark).

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

## `make gen`
The target `gen` is included for the convenience of the developers. It runs `go generate` on all the modules in alloy-go.

# Developers Notes
### Wed 27 May
#### _Notes on recent implementations_
A list is N amount of variable decl's surrounded by parenthesis.

	List = "(" { "," VariableDecl } ")"

A function declaration uses this list, for instance:

	FuncDecl = "func" iden list [ Type ] ( "{" Block "}" | "->" Statement )

However, since a list can be expression, this makes the following legal:

	func add(a: int = 5) {

	}

Therefore in a later stage (i.e. semantic analysis) you would have
to say that in the context of a list, assigning a variable is illegal. Another option
is that we create a new node for parameter values, which would be similar to a variable
decl, however assignment would not be valid. Currently, I'm assuming we're doing the former,
therefore I have made it so that semi-colons are enforced in the DeclNode for a VariableDecl
Node, as opposed to doing it at the end of the List, because of this change the parser
does not expect you to do this:

	func add(a: int,; b: int;)