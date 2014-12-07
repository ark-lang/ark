#include <stdio.h>
#include <stdlib.h>

#include "jayfor.h"

int main(int argc, char** argv) {
	// pass arguments through so the jayfor backend
	// can deal with it
	Jayfor *jayfor = jayforCreate(argc, argv);
	jayforStart(jayfor);
	jayforDestroy(jayfor);

	return 0;
}
