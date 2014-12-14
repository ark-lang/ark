# Jayfor
Jayfor is a programming language written in C.

# Table of Contents
* [About](#about)
  * [More details](JAYFOR.md)
* [Informal Specification](SPECIFICATION.md)
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

    fn add(int a, int b) [void] {
        ret [(a + b)]
    }

    fn test(int a, int b, str x) [int, float, str] {
        for int x:<10, 0> {
            add(x, x)
        }

        ret [0, 3.2, "test"]
    }

    fn main(int a, int b) [void] {
        test(5, 5)
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

# <a name="license"></a>License
Jayfor is licensed under The MIT License. I have no idea
what this means, I just randomly chose it. Read it [here](LICENSE.md)
