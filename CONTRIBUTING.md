Contributing
============

Community
---------

Having any problems? [Open an Issue](https://github.com/ink-lang/ink/issues).
We will be creating an IRC room shortly.

Pull Requests
-------------

All pull requests are welcome. Please ensure that you rebase your commits
against remote master prior to submitting a pull request, in order to
prevent edit collisions.

Proposing a feature or change
-----------------------------

If you want to have a say in the language, feel free to
[open an issue](https://github.com/ink-lang/ink/issues). We use proposals
so that people can vote on the idea, since the smallest change can put Ink
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

Here is a good example (obviously fake):

> [Feature] Require Line Numbers
> ==============================
>
> Ink should require line numbers to be prepended at the start of each
> line in Ink files. This will make it much easier to debug Ink programs,
> as it will be easier to find the line where a problem occurs, as opposed
> to having to use special functionality in a text editor. Here is the
> example hello world file with the changes implemented:
>
> [`helloworld.inks`](/examples/helloworld.inks):
> ```ink
> 	// use stdio; -- doesn't work yet, no preprocessor
>
>  // main entry point of the program
>  fn main(): int {
> 	// this function is inside the standard io library
> 	println("hello, world")
> 	return 0
> }
> ```

Please ensure that you check back on the issue often and actively participate
in the ensuing discussion until a consensus is reached.
