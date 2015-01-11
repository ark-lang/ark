# Jayfor - [![Build Status](https://travis-ci.org/jayfor-lang/jayfor.svg?branch=master)](https://travis-ci.org/jayfor-lang/jayfor)
Jayfor is a programming language written in C. The goals of this language are:

* Speed
* Simplicity
* Clean

## Why?
I created Jayfor as I love writing C, and I love the simplicity of C, yet I wanted to evolve C
into something more modern, and easier to use.

## Status
* The language is still in development
* The compiler is written in C
* LLVM is used for the backend

## Proposals
If you want to have a say in the language, feel free to post an Issue in the [Issue Handler](issues). We
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
you can change the [Makefile](Makefile) to use GCC or another compiler. Open your
terminal and run the following:

    # clone the repository
    git clone https://github.com/jayfor-lang/jayfor.git
    
    # cd into the repository
    cd jayfor

    # build
    make

    # run a program
    ./j4 tests/simple.j4

If something doesn't work, feel free to post in the [Issues Handler](issues).

## License
Jayfor is licensed under The MIT License. I have no idea
what this means, I just randomly chose it. Read it [here](misc/LICENSE.md)
