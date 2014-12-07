#include <stdio.h>
#include <stdlib.h>

#include "jayfor.h"

int main(int argc, char** argv) {
	Jayfor *jayfor = jayforCreate(argc, argv);
	jayforStart(jayfor);
	jayforDestroy(jayfor);
	return 0;
}
