### ALERT: **We changed the name from Jayfor to Ink. Its the same project, just a different name. [Here's why](https://github.com/ink-lang/ink/issues/68).**

----

[Ink](http://ink-lang.github.io) [![BuildStatus](https://travis-ci.org/ink-lang/ink.svg?branch=master)](https://travis-ci.org/ink-lang/ink)
===

is a programming language written in C. The goals of this language are:

* Speed
* Simplicity
* Cleanliness

We created Ink because we love writing C, and love the simplicity of C; we also wanted to evolve C
into something more modern, and easier to use.

IMPORTANT
---------

Ink started off and will, for the moment, **be an educational
project**. We'd love to see it go mainstream, but, as it stands,
there are a lot of stable languages already out there. Its a
tough challenge. So, for all those asking how good is this
language/will this language be when compared to Go or Rust or
`[insert language here]`, we want to clarify (again) that this
project was in no manner started for the sole purpose of competing
with languages already out there. Those that are interested in the
project and like the idea are more than welcome to contribute to its
growth (into something big). Being in a tender stage of development,
we still have a lot of features that we want to implement. But a
majority of the suggestions come from **you**. So please don't
hesitate to propose a change to the language (or better yet,
a pull request!), following the instructions mentioned somewhere
below. Please do not expect too much from the language (at least not
at the moment). Our priority is to get the language to a stage where
it can compute and produce results. Any thoughts about competition
or stance when compared to other languages **is the least of our
priorities right now**. 

Status
------

The language is still in development. More information/specifics
can be found in [`/misc/`](/misc/). The compiler is written in C
and LLVM is used for the backend.

Building
--------

To build you will need:

 - [`LLVM >= 3.4`](http://llvm.org/releases/download.html)
 - A suitable GNU C compiler (any one of the below will do fine):
   - [`clang >= 3.4.0`](http://llvm.org/releases/download.html)
   - [`gcc >= 4.8.1`](https://gcc.gnu.org/)

```bash
# clone the repository
git clone -v https://github.com/ink-lang/ink.git
    
# cd into the repository
cd ink

# build
make -j4

# run a program
./inkc tests/simple.ink
```

Something not working? Open an [Issue](https://github.com/ink-lang/ink/issues)
or send us a [Pull Request](https://github.com/ink-lang/ink/pulls)
on GitHub.

License
-------

Ink is licensed under the [MIT License](/LICENSE.md).
