Going to keep a little log for developer notes, what I'm working etc. Kind of like a shitty log. These are just ideas
and a way for me to keep track of the shit I do. 

## Friday, 27 March 2015
### 10:18pm
Haven't done much, ended up spending the rest of the day on Reddit, lol. I moved the old alloy server onto my 
own digital ocean for now since I forgot the details to Elliotts' server. I also moved the repo onto it's own organization,
since I decided to put the web site back up on GitHub, then I'll setup dploy.io to deploy it to mine/elliotts' server.
Currently re-considering the expression parsing at the moment.

### 5:48pm
Going to start working on the new parser, I decided to merge it before it fully worked, but it's now kind of working.
I might need to adjust the grammar, as it expects a semi-colon after everything, even nested functions, i.e:

	fn add(): void {
		fn add(): void {
		
		};
	}
	
Also, I need to fix expression parsing, since that seems to be causing **a lot** of the errors, which also means
I need to re-write operator precedence parsing, fun... As for an idea I just had, perhaps every C header file should include
a set of C libraries that I feel are only necessary, particularly just `stdio` and `stdlib`. I don't see why you would need
anything other than these two libraries, especially since we can write the rest of the libraries that you may need in Alloy,
i.e string manipulation, etc. One issue I can think of is that some libraries for C require c style strings, i.e null-terminated.
And before I get to work, I need to remember to do `loop` statements! These should be super easy, just syntactic sugar for `while (true)`.