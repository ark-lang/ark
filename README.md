# <a href="http://jayfor-lang.github.io">Jayfor</a> - [![Build Status](https://travis-ci.org/jayfor-lang/jayfor.svg?branch=master)](https://travis-ci.org/jayfor-lang/jayfor)
Jayfor is a programming language written in C. The goals of this language are:

* Speed
* Simplicity
* Clean

## Why?
We created Jayfor because we love writing C, and love the simplicity of C; we also wanted to evolve C
into something more modern, and easier to use.

## Status
* The language is still in development
* The compiler is written in C
* LLVM is used for the backend

## Contributing
### Community
We have an IRC where we discuss Jayfor things, and other stuff too:

    channel: ##jayfor
    server: irc.freenode.net

### Pull Requests
All pull requests are welcome:

* Fork the project
* Create your branch for the feature `git checkout -b toast`
* Commit your changes                `git commit -am 'Makes toast'`
* Push to the branch                 `git push origin toast`
* Submit a Pull Request

### Proposing a feature/change
If you want to have a say in the language, feel free to post an Issue in the [Issue Handler](https://github.com/jayfor-lang/jayfor/issues). We
use proposals so that people can vote on the idea, since the smallest change can put Jayfor in a completely
different direction. To help out the developers, please include the following in your Issue:

    Title:
    [PROPOSAL] - Proposal Summary.

    Contents:
    * What you are proposing
    * Why you think it's a good idea
    * If it's a syntax related proposal, post any ideas of syntax you can. This isn't mandatory, but will help.

## Building
To build you will need **LLVM 3.4 or above**, and **clang**. If you don't have clang,
you can change the [Makefile](Makefile) to use **GCC** or another compiler. Open your
terminal and run the following:

    # clone the repository
    git clone https://github.com/jayfor-lang/jayfor.git
    
    # cd into the repository
    cd jayfor

    # build
    make

    # run a program
    ./j4 tests/simple.j4

If something doesn't work, feel free to post in the [Issues Handler](https://github.com/jayfor-lang/jayfor/issues).

## License
Jayfor is licensed under The MIT License. I have no idea
what this means, I just randomly chose it. Read it [here](misc/LICENSE.md)
