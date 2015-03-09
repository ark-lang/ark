# [Alloy™](http://alloy-lang.github.io) [![BuildStatus](https://travis-ci.org/alloy-lang/alloy.svg?branch=master)](https://travis-ci.org/alloy-lang/alloy)
Alloy is a work in progress programming language.

## IRC
We have an IRC where we discuss Alloy, and other stuff too, come join! If you want to help contribute,
we highly suggest you join the IRC :)

* server: irc.freenode.net
* channel: #alloy-lang

## Building
To build you will need:

 - Make 3(.81???)
 - A suitable GNU C compiler (any one of the below will do fine, we aren't sure about other C compilers quite yet):
   - [`clang >= 3.4.0`](http://llvm.org/releases/download.html)
   - [`gcc >= 4.8.1`](https://gcc.gnu.org/) (change `C` in the [makefile](/Makefile) to `gcc`)

```bash
# clone the repository
git clone -v https://github.com/alloy-lang/alloy.git
    
# cd into the repository
cd alloy

# build
make

# run a program
cd tests

# run the test bash script
# if some of these fail, it might be because
# the developers have changed the document
./alloyctest.sh
```

## License
Alloy is licensed under the [MIT License](/LICENSE.md).
