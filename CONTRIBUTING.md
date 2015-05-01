# Contributing
We'd love your contributions to make Alloy a better programming language.

## Code Style
Alloy has a fairly strict code-style to keep things consistent:

* Indent with four spaces, no tabs! (Your IDE's/Editors should do this anyway)
* Provide spaces around operators for readability, i.e. (1 + 1), over (1+1)
* Constants and defines are uppercase and snake case, e.g. `THIS_IS_A_CONSTANT`
* Structures are Pascal case, e.g. `SomeStruct`
* Instances and similar are Camel case, e.g. `SomeStruct *someStruct;`
* C code should mostly be compliant with C99 standards
* Code in "C with classes" style
  - Methods should have the "class" as the first parameter, e.g. `void createXYZ(XYZ *...)`
  - The parameter must be named `self`, e.g. `void createXYZ(XYZ *self)`
  - A class must have two methods, namely: `createClass`, and `destroyClass`. Class
    would be replaced with the name of the class, e.g. `createLexer`, `destroyLexer`.

## Community
Having any problems? [Open an Issue](https://github.com/felixangell/alloy/issues). 

### IRC
There is also an IRC room over on freenode: [#alloy-lang](http://webchat.freenode.net/?channels=%23alloy-lang), here's the info if you want to enter it into your own client:

* ip: `irc.freenode.net`
* channel: `#alloy-lang`

### Subreddit
We also have a subreddit, [here's the link](http://www.reddit.com/r/alloy_lang). Feel free to ask us questions on that,
a lot of the developers are also Reddit addicts.

### Mailing List
We have no idea how this works, but apparently we should have one, [so here it is](https://groups.google.com/forum/#!forum/alloy-lang).
It's recommended that you open an issue than head over to the IRC since we dont check the mailing list too often.

## Pull Requests
Please ensure that you rebase your commits against remote master prior to submitting a pull request, in order to prevent edit collisions. Pull Requests will be reviewed by the contributors, and other members of the community.

## Proposing a feature or change
If you want to have a say in the language, feel free to
[open an issue](https://github.com/felixangell/alloy/issues). We use proposals
so that people can vote on the idea, since the smallest change can put Alloy
in a completely different direction. To help us all out, please include
the following when submitting your issue:

> Title: [Feature/Change] Proposal Summary
> =====================================
>
> Contents:
>
> 1. Your proposition, described with as much detail as possible.
> 2. Why time and work should be put into developing out this feature.
> 3. An(y) example(s) of how your proposition would work/be used.
> 4. Any additonal data.

Please ensure that you check back on the issue often and actively participate
in the ensuing discussion until a consensus is reached.
