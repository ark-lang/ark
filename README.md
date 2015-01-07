# Jayfor - [![Build Status](https://travis-ci.org/jayfor-lang/jayfor.svg?branch=master)](https://travis-ci.org/jayfor-lang/jayfor)
Jayfor is a programming language written in C.

# <a name="proposals"></a>Proposing Ideas
If you have an idea for the language, post an issue with the format: `Proposal: proposal summary` into the [issues](http://github.com/jayfor-lang/jayfor/issues) section of the repository, for example:

    Proposal: add X to this language because Y
    Proposal: add X to this language to replace Y

Please tag any proposals with the "proposal" label!

# <a name="requirements"></a>Requirements
* Clang 3.5
* LLVM 3.5 

# <a name="IRC"></a>IRC
We have an IRC channel, this is mostly where we discuss development related
issues, syntax styles, and so on. Interested in contributing, come over
to our IRC channel:

    server:     irc.freenode.net
    port:       8001
    channel:    ##jayfor

Since there aren't many of us, we're mostly active around 4pm GMT on weekdays.

## <a name="building"></a>Building
Building is easy:

    // note, you atleast need LLVM 3.5!

    cd where_you_want_jayfor_to_be
    git clone https://www.github.com/jayfor-lang/jayfor.git
    cd jayfor
    make

Then to run a j4 program:

    ./j4 tests/simple.j4

# <a name="syntax"></a>Syntax
Check the REFERENCE.md file in misc/.

# <a name="license"></a>License
Jayfor is licensed under The MIT License. I have no idea
what this means, I just randomly chose it. Read it [here](misc/LICENSE.md)
