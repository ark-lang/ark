# Contributing
We'd love your contributions to make Ark a better programming language.

## Interested in contributing?
If you want to contribute to Ark, that's great! We encourage all 
developers to actively discuss Ark and its features/implementations 
on the IRC ([#ark-lang](http://webchat.freenode.net/?channels=%23ark-lang)).

### What can I do to help? 
There are plenty of things you can do to help:

* Comment on issues, your voice helps us make the language better!
* Fix bugs/implement features in the issue tracker
* Create/fix/modify documentation
* Work on the standard library
* Work on the compiler

## Pull Requests
We love Pull Requests! Please make sure you work on a separate branch, and
that you rebase with the master before submitting. If you are implementing
a feature of some kind, please add a test if there is not an existing one
already! Please add a summary or brief description of your PR, however the
more detail the better.

## Fixing Bugs
If you've found a bug when implementing a feature, or want to
fix an existing bug; here are what we consider must do steps.

* Run it through a debugger, e.g. `lldb` or `gdb`.
  Typically, if theres a segmentation fault, or other bug
  that occurrs on run-time, g/lldb is a good start.
* Use valgrind to track down memory leaks
* Run the ark code with the `-v` flag! This is the verbose
  flag and can help you locate what *stage the bug occurrs.
  *during lexical analysis, semantic analysis, parsing, etc...

## Writing Ark code
Be it personal code, or you're contributing to the standard libraries,
check out the [style guide](/STYLEGUIDE.md). Not only will this keep everything
consistent (if you're contributing to the std libraries that is), but it will
also make the code a lot more readable and easier to maintain.

## Community
Please do not post questions for help in the Issue Tracker _unless_ you think
it's a bug related to the ark compiler, or a feature. Instead, check out
the following communities:

### IRC
You'll probably find one of the core super contributors on the IRC. If you have
any questions, or just want to come talk to us; join the [#ark-lang](http://webchat.freenode.net/?channels=%23ark-lang)
channel on freenode.

* ip: `irc.freenode.net`
* channel: `#ark-lang`

### Subreddit
We also have a subreddit, [here's the link](http://www.reddit.com/r/ark_lang). Feel free to ask us questions on that,
a lot of the developers are also Reddit addicts.

### Mailing List
We have no idea how this works, but apparently we should have one, [so here it is](https://groups.google.com/forum/#!forum/ark-lang).
It's recommended that you open an issue than head over to the IRC since we dont check the mailing list too often.
