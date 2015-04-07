#ifndef __SCOPE_H
#define __SCOPE_H

#include "vector.h"

/**

	This is for a smart pointer implementation. When a scope is entered,
	a Scope object is added to a Stack of Scopes. When a pointer is defined
	and declared, it will be added to the "pointers" list. When a scope is exited,
	the list of pointers will have `free` calls inserted, then the scope is destroyed
	and subsequently popped from the stack.

	This is a compile-time operation, thus there will be no overhead during the execution
	of the program. 

 */

typedef struct {
	Vector *pointers;
} Scope;

Scope *createScope();

void destroyScope(Scope *scope);

#endif // __SCOPE_H