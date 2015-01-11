# Jayfor - [![Build Status](https://travis-ci.org/jayfor-lang/jayfor.svg?branch=master)](https://travis-ci.org/jayfor-lang/jayfor)
Jayfor is a programming language written in C.

# <a name="proposals"></a>Proposing Ideas
If you have an idea for the language, post an issue with the format: `Proposal: proposal summary` into the [issues](http://github.com/jayfor-lang/jayfor/issues) section of the repository, for example:

    Proposal: add X to this language because Y
    Proposal: add X to this language to replace Y

Please tag any proposals with the "proposal" label!

# Examples
There are some code examples in the `examples` folder.

# Requirements
* Clang 3.5
* LLVM 3.5 

# IRC
We have an IRC channel, this is mostly where we discuss development related
issues, syntax styles, and so on. Interested in contributing to Jayfor, come over
to our IRC channel:

    server:     irc.freenode.net
    port:       8001
    channel:    ##jayfor

## Building
Building is easy. **Please note that this langauge is still
in design phase, so we can't promise anything will work!**

    // note, you atleast need LLVM 3.4 or above
    // if you build with 3.5 you may have to add
    // additional linker options

    cd where_you_want_jayfor_to_be
    git clone https://www.github.com/jayfor-lang/jayfor.git
    cd jayfor
    make

Then to run a j4 program:

    ./j4 exampes/helloworld.j4

# Syntax
Check the REFERENCE.md file in misc/.

# License
Jayfor is licensed under The MIT License. I have no idea
what this means, I just randomly chose it. Read it [here](misc/LICENSE.md)
