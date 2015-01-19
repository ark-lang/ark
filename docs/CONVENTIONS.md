#Conventions 

Please conform to the following conventions when submitting code to sustain readability.

## Snake Case
We use snake case

    like_this

## Comments
Single line comments can be either before the statement or after, like so:

    // hi
    int x = 5;

    int x = 5; // hi

The first option is preferred, though.

Multi-line comments must be as follows:

    /* 
     * this comment is to demonstrate
     * multi-line commenting conventions
     * for the Ink project.
     */

### Documentation Comments
Documentation comments with two asteriks, we use
doxygen -- so conform to their system please :)

    /**
     * This is my documentation
     * @param param1 this is doxygen
     */

All documentation for functions should go in the
header files!

===

##Functions

Functions must be declared as per usual, and the opening brace must come on the same line as the 
function after a single space.

    void test(int x, float y) { // <- the brace must be on the same line, after 1 space
        return x + y;
    }

===

##If conditionals

If the `if` conditional must execute a single statement, use the braces anyway.

    if (condition) {
        x = 1 + 2;
    } 
    else {
        x = 1 + 3;
    }

Instead of doing this:

    if(condition)
        x = 1 + 2;

This is to maintain consistency, and avoid potential errors.

**This rule applies to all other loops/conditionals, not just for the *if* conditional**.

    for (int x = 1; x < 10; x++) {
        x = x + 1;
    }

and so on...


===

Also, please do configure your text editors such that the `tab` button adds 4 spaces, as some text editors default it to 2, which
we don't want. All tabs must add 4 spaces. After every curly brace `{`, the next line must move 4 spaces in before the code is added.

Example:

    if(condition) {
        // 4 spaces
        x = 1 + 2;
    } // close right below where the loop started

Using IDEs is not recommended, considering that we have a Makefile and all of the work was done on Vim and Sublime Text. Please 
refrain from using any IDEs (it can cause problems later on and add unnecessary files). Use the command line for building the 
project and text editor to edit the programs. If you insist on using an IDE, please add any files that it creates to the gitignore.
