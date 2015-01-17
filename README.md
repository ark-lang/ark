### ALERT: **We changed the name from Jayfor to Ink. Its the same project, just a different name. [Here's why.](https://github.com/ink-lang/ink/issues/68)**

----

# <a href="http://ink-lang.github.io">Ink</a> - [![Build Status](https://travis-ci.org/ink-lang/ink.svg?branch=master)](https://travis-ci.org/ink-lang/ink)
Ink is a programming language written in C. The goals of this language are:

* Speed
* Simplicity
* Clean

## Why?
We created Ink because we love writing C, and love the simplicity of C; we also wanted to evolve C
into something more modern, and easier to use.

## Status
* The language is still in development
* The compiler is written in C
* LLVM is used for the backend

## Contributing
### Community
The IRC will be up soon.

### Pull Requests
All pull requests are welcome:

* Fork the project
* Clone the repository from your account `git clone git@github.com:your_user/ink.git`
* Create your branch for the fix `git checkout -b ink-toast-fix`
* Make your changes
* Submit a Pull Request
* **IMPORTANT** Please rebase on master everytime before you push to check for conflicts

### Proposing a feature/change
If you want to have a say in the language, feel free to post an Issue in the [Issue Handler](https://github.com/ink-lang/ink/issues). We
use proposals so that people can vote on the idea, since the smallest change can put Ink in a completely
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
    git clone https://github.com/ink-lang/ink.git
    
    # cd into the repository
    cd ink

    # build
    make

    # run a program
    ./inkc tests/simple.ink

If something doesn't work, feel free to post in the [Issues Handler](https://github.com/ink-lang/ink/issues).

## License
Ink is licensed under The MIT License. I have no idea
what this means, I just randomly chose it. Read it [here](misc/LICENSE.md)
