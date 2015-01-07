# Jayfor - [![Build Status](https://travis-ci.org/jayfor-lang/jayfor.svg?branch=master)](https://travis-ci.org/jayfor-lang/jayfor)
Jayfor is a programming language written in C.

# Table of Contents/Resources
* [Proposals](#proposals)
* [Requirements](#requirements)
* [Contributing](#contributing)
* [IRC](#IRC)
* [About](#about)
* [Reference](misc/REFERENCE.md)
* [License](#license)

# <a name="proposals"></a>Proposing Ideas
If you have an idea for the language, post an issue with the format of `Proposal: proposal summary` into the [issues](http://github.com/jayfor-lang/jayfor/issues) section of the repository, for example:

    Proposal: add X to this language because Y
    Proposal: add X to this language to replace Y

Please tag any proposals with the "proposal" label!

# <a name="requirements"></a>Requirements
* Clang 3.5
* LLVM 3.5 

# <a name="contributing"></a>Contributing
## Why should I contribute?
Jayfor is a programming language in the early stages, so it's a lot easier to work on than
say a more developed language like Rust.
As we are a long way from hitting that stable 1.0, we have more time to experiment with the
language, which means you can have a say in new syntax ideas, and so on.

Send a pull request to [http://github.com/jayfor-lang/jayfor](http://github.com/jayfor-lang/jayfor). Use [http://github.com/jayfor-lang/jayfor/issues](http://github.com/jayfor-lang/jayfor/issues) for discussion. Please note that we consider that you have granted non-exclusive right to your contributed code under the MIT License.
**Please conform to the programming rules and write short, but meaningful commit messages.**

# <a name="IRC"></a>IRC
We have an IRC channel, this is mostly where we discuss development related
issues, syntax styles, and so on. Interested in contributing, come over
to our IRC channel:

    server:     irc.freenode.net
    port:       8001
    channel:    ##jayfor

Since there aren't many of us, we're mostly active around 4pm GMT on weekdays.

# <a name="about"></a>About
Jayfor is a programming language written in C. It is still under
heavy development. Jayfor is influenced from languages like Rust,
C, C++, and Java. We like to keep things simple, safe and fast,
but maintain a syntactically beautiful language.
You can view the snippet of code below for a basic 'feel' of the language.

    // this is no longer a test for while loops
    // just some random code for testing

    int a = (5 + 5);
    int z = a;
    int a;
    char b = 'a';

    fn add(int a, int b): int {
        return (a + b);
    }

    fn test(int a, int b, str x): <int, float, str> {
        for int whatever:(10, 0) {
            add(a, b);
        }

        return <a, 3.2, "test">;
    }

    fn main(int a, int b): void {
        test(5, 5, "rand");
    }

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
