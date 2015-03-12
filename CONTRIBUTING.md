# Contributing
We'd love your contributions, but we do have a few guidelines
that we go by here.

## Community
Having any problems? [Open an Issue](https://github.com/alloy-lang/alloy/issues). There is also an IRC room over on freenode: [#alloy-lang](http://webchat.freenode.net/?channels=%23alloy-lang).

It's recommended that you open an issue than head over to the IRC since we rarely go there given our tight schedules. Issues will however be responded to ASAP. 

## Pull Requests
All pull requests are welcome. Please ensure that you rebase your commits
against remote master prior to submitting a pull request, in order to
prevent edit collisions. Pull Requests will be judged by the following 2 
developers:

* felixangell
* vnev

who will merge it if it's up to standard.

## Proposing a feature or change
If you want to have a say in the language, feel free to
[open an issue](https://github.com/alloy-lang/alloy/issues). We use proposals
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

Here is a good example (obviously fake):

> [Feature] Require Line Numbers
> ==============================
>
> Alloy should require line numbers to be prepended at the start of each
> line in Alloy files. This will make it much easier to debug Alloy programs,
> as it will be easier to find the line where a problem occurs, as opposed
> to having to use special functionality in a text editor. Here is the
> example hello world file with the changes implemented:
>
> [`helloworld.ay`](/examples/helloworld.ay):
> ```
> 	 !use stdio
>
>    // main entry point of the program
>    fn main(): int {
> 	    // this function is inside the standard io library
> 	 	println("hello, world")
> 	 	return 0
> 	 }
> ```

Please ensure that you check back on the issue often and actively participate
in the ensuing discussion until a consensus is reached.
