#include <stdio.h>
#include <time.h>

#include "jayfor.h"

int main(int argc, char** argv) {

	// for calculating time taken
	clock_t timer = clock();

	// jayfor stuff
	Jayfor *jayfor = jayforCreate(argc, argv);
	jayforStart(jayfor);
	jayforDestroy(jayfor);

	// calculate time
	timer = clock() - timer;
	double timeTaken = ((double) timer) / CLOCKS_PER_SEC;
	printf("Finished in %f/s\n", timeTaken);

	return 0;
}
