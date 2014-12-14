# Jayfor
Jayfor is a programming language written in C.

# Table of Contents
* [About](#about)
  * [More details](JAYFOR.md)
* [Informal Specification](SPECIFICATION.md)
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

# <a name="about"></a>About
Jayfor is a programming language written in C. It is still under
heavy development. Jayfor is influenced from languages like Rust,
C, C++, and Java. We like to keep things simple, safe and fast,
but maintain a syntactically beautiful language.
You can view the snippet of code below for a basic 'feel' of the language,
or check out some actual code we use for the library [here](libs/math.j4).

	// add the given values together
	fn add(int a = 0, int b = ((5 + 5) - (5 + 5))) [int] {
		for int x:<0, 10, 2> {
            add(x, 5)
        }
        ret [(a + b)]
	}

	fn main() [int] {
		int a = add(5, 3)
		println(a); // semi colons are optional!
        ret [a]
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
