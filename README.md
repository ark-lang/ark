# [Alloy™](http://alloy-lang.org) [![BuildStatus](https://travis-ci.org/felixangell/alloy.svg?branch=master)](https://travis-ci.org/felixangell/alloy)
Alloy is a work in progress programming language.

## Contributing
If you want to help us make this language, check out the [CONTRIBUTING](/CONTRIBUTING.md) file! :)

## General Notes
The language is a work in progress, the developers are mostly students who have jobs, lives, schoolwork and other things to maintain, therefore
development could be slow, and has been! Alloy is written in C, we're trying to get the basics down before we begin optimizing, the code is kind of
messy and needs refactoring. We prefer using clang to compile the code, and GCC to compile the generated code. 

### Status

* Lexer - Completed
* Parser - Mostly complete, some syntax is unimplemented
* Semantic Analysis - TODO, Not a huge priority, currently the compiler assumes code is valid
* Code Generation - In progress
* Bootstrapped - TODO

## IRC
We have an IRC where we discuss Alloy, and other stuff too, come join! If you want to help contribute,
we highly suggest you join the IRC :)

* server: irc.freenode.net
* channel: #alloy-lang

## Building
Disclaimer: This project is constantly in development and is still a work in progress, we are still a long way away from getting everything to work somewhat smoothly, so if it breaks or doesn't build, sorry!
To build you will need:

 - Make 3(.81???)
 - A suitable GNU C compiler (any one of the below will do fine, we aren't sure about other C compilers quite yet):
   - [`clang >= 3.4.0`](http://llvm.org/releases/download.html)
   - [`gcc >= 4.8.1`](https://gcc.gnu.org/) (change `C` in the [makefile](/Makefile) to `gcc`

```bash
# clone the repository
git clone -v https://github.com/felixangell/alloy.git
    
# cd into the repository
cd alloy

# build
make
```

## License
Alloy is licensed under the [MIT License](/LICENSE.md).
