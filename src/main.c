#include <stdio.h>
#include <time.h>

#include "alloyc.h"

int main(int argc, char** argv) {
	// start the timer
	clock_t timer = clock();

	// alloyc stuff
	alloyc *alloyc = create_alloyc(argc, argv);
	start_alloyc(alloyc);
	destroy_alloyc(alloyc);

	// finished timer
	timer = clock() - timer;	// calculate time taken
	
	// calculate seconds and milliseconds taken
	double secondsTaken = ((double) timer) / CLOCKS_PER_SEC;
	double msTaken = secondsTaken * 1000;

	primary_message("Finished in %.6f/s (%.0f/ms)", secondsTaken, msTaken);

	return 0;
}
