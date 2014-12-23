#include <stdio.h>
#include <time.h>

#include "jayfor.h"

int main(int argc, char** argv) {
	// start the timer
	clock_t timer = clock();

	// jayfor stuff
	jayfor *jayfor = create_jayfor(argc, argv);
	start_jayfor(jayfor);
	destroy_jayfor(jayfor);

	// finished timer
	timer = clock() - timer;	// calculate time taken
	
	// calculate seconds and milliseconds taken
	double secondsTaken = ((double) timer) / CLOCKS_PER_SEC;
	double msTaken = secondsTaken * 1000;

	printf(KGRN "Finished in %.3f/s (%.0f/ms)\n" KNRM, secondsTaken, msTaken);

	return 0;
}
