# Semantic Analysis
This is a guide for me to see what we should be checking in the semantic
analysis.

## Variables already defined

    int x = 5;

    // error!
    int x = 10;

## Const variable being re-assigned

    int x = 5;

    // error!
    x = 20;

## Mutables being passed to constants

    // error, needs to be mutable!
    fn change(int ^a): void {
        ^a = 10;
    }

    fn main(): void {
        // alloc x and set to 10
        int ^x = ~10;
        change(x);
        free(^x)
    }

## Uninitialized Variables

    // invalid, since it's constnat
    // it must be initialized
    int x;

    // valid, set to zero by default
    mut int x;

    // valid, does not initialize the variable
    mut int x = _;