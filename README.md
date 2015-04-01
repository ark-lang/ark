# [Alloy](http://alloy-lang.org) [![BuildStatus](https://travis-ci.org/alloy-lang/alloy.svg?branch=master)](https://travis-ci.org/alloy-lang/alloy)
Alloy is a work in progress programming language, read the [reference](docs/REFERENCE.md) to find out more about the plans, syntax, etc. Please note that the language is constantly being changed, so the Reference may be outdated, or a little bit behind the master.

## Building
Alloy is a work in progress, if you would like to contribute to the creation of the language, or just give it a go, you will
need a few things:

* GCC/Clang (another compiler may work, but we test it works on these mostly)
* Make

```bash
$ git clone http://www.github.com/alloy-lang/alloy
$ cd alloy
$ make
```

After you make the project, the alloyc executable will be located in `bin/alloyc`. For a list of commands,
run the `-h` flag. There are test files under the `tests/` directory, however the one in the example may not exist
if you are reading this in the future! We also can't guarantee that it will still work either!

**The compilers generated code will be compiled with GCC by default, you can use a flag to change this however, see the code snippet!**
```bash
bin/alloyc tests/factorial.ay

// IF YOU DO NOT HAVE GCC OR WANT TO USE CLANG FOR EXAMPLE, DO THIS!
bin/alloyc somefile.ay -compiler clang

// an executable will be created in your current directory called `main`
// run it like so...
./main
```

## Community
If you are looking for some help, want to ask a question, or just want to talk to us for whatever reason, we have a few
places where you can find us.

### Subreddit
We have a subreddit, this is mostly to get it before someone else does, but a lot of the developers are active Redditors. So feel
free to ask questions on there, etc.
[Link to the subreddit](http://www.reddit.com/r/alloy_lang)

### Mailing List
We have no idea how this works, but apparently we should have one. You can find it here:
[Mailing List](https://groups.google.com/forum/#!forum/alloy-lang)

### IRC
We have an IRC where we discuss Alloy, and other stuff too. Come join! If you want to help contribute,
we highly suggest you join the IRC (although we might not always be available due to time zone differences) :)

* server: irc.freenode.net
* channel: #alloy-lang

## Contributing
If you want to contribute, check out the [CONTRIBUTING](CONTRIBUTING.md) file.

## Contributor List
Here are the people who are working, or have worked on Alloy. Feel free to add yourself
to this list when you commit! :)

* [felixangell](https://github.com/felixangell)
* [vnev](https://github.com/vnev)
* [haneefmubarak](https://github.com/haneefmubarak)
* [xhjo](https://github.com/xhjo)
* [IanMurray](https://github.com/IanMurray)
* [requimrar](https://github.com/requimrar)
* [ianhedoesit](https://github.com/ianhedoesit)
* [FelixWittmann](https://github.com/FelixWittmann)
* [paraxor](https://github.com/paraxor)
* [andars](https://github.com/andars)
* [theunamedguy](https://github.com/theunamedguy)

## Additional Credits/Libraries

* LLVM
* SDS

## License
Alloy is licensed under the [MIT License](/LICENSE.md).
