# JAYFOR
JAYFOR is a compiled programming language

## Frontend
The frontend of JAYFOR is fairly simple, we use a hand written Lexer and Parser. 

### Why a hand written Lexer/Parser?
* I've never used a lexer/parser generator 
* I like writing things from scratch
* Fun
* Learning more about lexers/parsers
* Tailoring the lexer/parser specifically to the J4 language
* Less bloat

## Backend
The backend uses LLVM. We use LLVM because
* More optimisations in there that I could do in a lifetime
* I've never written a Bytecode Compiler
* It's super duper fast
* Easy to use

# FELIXS REMEMBER STUFF
for type index:<start, end, step> {
    
}

fn add(int a, int b)[int] {
    
}

// tuples
fn add(int a, int b)[int, double] {
    
}

if (x == 5) {
    
}

// while loop
while (x == 10) {
    
}

// do while
do (x == 5) {
    
}