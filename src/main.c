#include <stdio.h>
#include <time.h>

#include "jayfor.h"

int main(int argc, char** argv) {
	// start the timer
	clock_t timer = clock();

	// jayfor stuff
	Jayfor *jayfor = jayforCreate(argc, argv);
	jayforStart(jayfor);
	jayforDestroy(jayfor);

	// finished timer
	timer = clock() - timer;	// calculate time taken
	double timeTaken = ((double) timer) / CLOCKS_PER_SEC;	// in seconds
	timeTaken *= 1000;	// convert to milliseconds

	printf(KGRN "Finished in %.3f/ms\n" KNRM, timeTaken);

	return 0;
}
