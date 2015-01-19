[Ink](http://ink-lang.github.io) [![BuildStatus](https://travis-ci.org/ink-lang/ink.svg?branch=master)](https://travis-ci.org/ink-lang/ink)
===

Ink is a programming language written in C. The goals of this language are:

* Speed
* Simplicity
* Cleanliness

We created Ink because we love writing C, and love the simplicity of C; we also wanted to evolve C
into something more modern, and easier to use.

Diclaimer
------
I [freefouran] personally started this project with a friend [CaptainChloride] for fun, we're not
trying to revolutionize anything, it's a small project we're doing for fun and learning. The project
is still in it's early stages of development, and is far from complete. We're currently trying to
implement the core of the language, so it will be pretty much identical to C, with a few quirks here
and there. Once we've got that down we're going to implement all the funky stuff, this is where **you**
can decide what you would like to see in the language, it could be anything, and you can try and implement
it yourself, or just post an issue and see if someone likes the idea and is willing to implement it for you.
Since we're both young (16 and 17), we have loads of exams, school work, jobs, which means we can't 
work on the language 24/7, but the more contributors we get, the quick the language gets done. If you're
an active contributor, there's also a good chance that you will be added to the **official developer team**,
ooooooooooooooohhhh, exciting.

Status
------

The language is still in development. More information/specifics
can be found in [`/docs/`](/docs/). The compiler is written in C
and LLVM is used for the backend.

Building
--------

To build you will need:

 - [`LLVM >= 3.4`](http://llvm.org/releases/download.html)
 - A suitable GNU C compiler (any one of the below will do fine):
   - [`clang >= 3.4.0`](http://llvm.org/releases/download.html)
   - [`gcc >= 4.8.1`](https://gcc.gnu.org/) (change `LCC` and 
     `LCXX` in the [makefile](/Makefile) to `gcc` and `g++`, respectively)

```bash
# clone the repository
git clone -v https://github.com/ink-lang/ink.git
    
# cd into the repository
cd ink

# build
make

# run a program
./inkc tests/parser-tests/const.ink
```

Contributing
------------

Something not working? Open an [Issue](https://github.com/ink-lang/ink/issues)
or send us a [Pull Request](https://github.com/ink-lang/ink/pulls)
on GitHub.

See [CONTRIBUTING.md](/CONTRIBUTING.md).

License
-------

Ink is licensed under the [MIT License](/LICENSE.md).
