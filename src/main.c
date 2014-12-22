#include <stdio.h>
#include <time.h>

#include "jayfor.h"

int main(int argc, char** argv) {
	// start the timer
	clock_t timer = clock();

	// jayfor stuff
	Jayfor *jayfor = create_jayfor(argc, argv);
	start_jayfor(jayfor);
	destroy_jayfor(jayfor);

	// finished timer
	timer = clock() - timer;	// calculate time taken
	double timeTaken = ((double) timer) / CLOCKS_PER_SEC;	// in seconds

	printf(KGRN "Finished in %f/s\n" KNRM, timeTaken);

	return 0;
}
