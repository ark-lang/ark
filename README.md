# Jayfor - [![Build Status](https://travis-ci.org/jayfor-lang/jayfor.svg?branch=master)](https://travis-ci.org/jayfor-lang/jayfor)
Jayfor is a programming language written in C.

# Table of Contents/Resources
* [About](#about)
* [Reference](misc/REFERENCE.md)
* [IRC](#IRC)
* [Note](#note)
* [Requirements](#requirements)
* [Contributing](#contributing)
* [License](#license)

# Unique(-ish) features
* Semi-colons are optional
* Tuples
* Default argument values
* Reimagined syntax for annoying [traditional] features
  * do whiles, for loops, etc

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
You can view the snippet of code below for a basic 'feel' of the language,
or check out some actual code we use for the library [here](libs/math.j4).

    // this is no longer a test for while loops
    // just some random code for testing

    int a = (5 + 5)
    int z = a
    int a
    char b = 'a'

    fn add(int a, int b): int {
        ret (a + b)
    }

    fn test(int a, int b, str x): <int, float, str> {
        for int x:(10, 0) {
            add(a, b)
        }

        ret <a, 3.2, "test">
    }

    fn main(int a, int b): void {
        test(5, 5, "rand")
    }

# <a name="note"></a>Note
JAYFOR is still in design stage. It's not quite working yet, stay tuned.

# <a name="requirements"></a>Requirements
* GCC/Clang
* LLVM

# <a name="syntax"></a>Syntax
Check the SPECIFICATION.md file.

# <a name="contributing"></a>Contributing
Send a pull request to [http://github.com/freefouran/jayfor](http://github.com/freefouran/jayfor). Use [http://github.com/freefouran/jayfor/issues](http://github.com/freefouran/jayfor/issues) for discussion. Please note that we consider that you have granted non-exclusive right to your contributed code under the MIT License.
**Please conform to the coding rules and write short, but meaningful commit messages.**

# <a name="license"></a>License
Jayfor is licensed under The MIT License. I have no idea
what this means, I just randomly chose it. Read it [here](misc/LICENSE.md)
