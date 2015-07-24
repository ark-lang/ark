# Contributing
We'd love your contributions to make Ark a better programming language.

***All code changes to the Go codebase must have `go fmt` run on them***

## Interested in contributing?
If you want to contribute to Ark, that's great! We encourage all
developers to actively discuss Ark and its features/implementations
on the IRC ([#ark-lang](http://webchat.freenode.net/?channels=%23ark-lang)).

### What can I do to help?
There are plenty of things you can do to help:

* **Report bugs/other issues (big help!)**
* Comment on issues, your voice helps us make the language better!
* Fix bugs/implement features in the issue tracker
* Create/fix/modify documentation
* Work on the compiler
* RFC/Specs

## Issue Label Documentation
To keep things organized, we have introduced various labels that contributors with sufficient
permissions can add labels to issues. 

Labels are prefixed with a letter, the letter denotes the category of label:

* E: A label prefixed with E is how much experience is required to take on the issue
* M: These are miscellaneous issues, typically ones too broad to categorize
* O: These are the operating system/platform that an issue is relevant to
* P: The priority of the issue
* S: The stage in which the issue occurs (note not all issues have stages)
* A: An area in which the issue is relevant to, e.g. A-tests is for issues related to tests

## Pull Requests
We love Pull Requests! If you are planning on contributing, please read this list:

* Run `go fmt` on your code
* Work on a separate branch
* Make sure your patch is up-to-date with the master branch, if not, please rebase
* If you are implementing a feature of some kind, please add a test if there is not an existing one already
* Please add a summary or brief description of your PR, however, the more detail the better

## Fixing Bugs
If you've found a bug when implementing a feature, or want to
fix an existing bug; these steps will help you out.

* Run it through a debugger, e.g. `lldb`/`gdb`.
  Typically, if theres a segmentation fault, or other bug
  that occurs on run-time, g/lldb is a good start.
* Use Valgrind to track down memory leaks
* Run the Ark code with the `-v` flag! This is the verbose
  flag and can help you locate what stage the bug occurs (during lexical analysis, semantic analysis, parsing, etc).

## Writing Ark code
Be it personal code, or you're contributing to the standard libraries,
check out the [style guide](https://github.com/ark-lang/ark-docs/blob/master/STYLEGUIDE.md). Not only will this keep everything
consistent (if you're contributing to the std libraries that is), but it will
also make the code a lot more readable and easier to maintain.

## Ark RFCs/Specs
If you want to request a feature, or improve upon an existing feature, keep reading. If you want to help decide how to integrate an unimplemented feature, or give ideas on how it would work, keep reading. 

We now have a repository for RFC's and Specifications, this is to keep everything separate from issues related to the compiler itself (in terms of implementations of features, bugs, etc). If you want to request a feature, or you want to help use decide how to integrate a feature into the language, or just change something in the language: please go visit [this repository](//github.com/ark-lang/ark-rfcs).

## Community
Please do not post questions for help in the Issue Tracker _unless_ you think
it's a bug related to the Ark compiler, or a feature. Instead, check out
the following communities:

### IRC
You'll probably find one of the core super contributors on the IRC. If you have
any questions, or just want to come talk to us; join the [#ark-lang](http://webchat.freenode.net/?channels=%23ark-lang)
channel on freenode.

* ip: `irc.freenode.net`
* channel: `#ark-lang`

### Reddit
We also have a subreddit, [here's the link](http://www.reddit.com/r/ark_lang). Feel free to ask us questions on that,
a lot of the developers are also Reddit addicts.
