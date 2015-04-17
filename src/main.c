#include <stdio.h>

#include "alloyc.h"

int main(int argc, char** argv) {
	// start the timer
	clock_t timer = clock();

	// alloy compiler stuff
	AlloyCompiler *alloyc = createAlloyCompiler(argc, argv);
	if (alloyc) {
		startAlloyCompiler(alloyc);
		destroyAlloyCompiler(alloyc);
	}

	// finished timer
	timer = clock() - timer;	// calculate time taken

	if (alloyc) {
		// calculate seconds and milliseconds taken
		double secondsTaken = ((double) timer) / CLOCKS_PER_SEC;
		double msTaken = secondsTaken * 1000;

		primaryMessage("Finished in %.6f/s (%.0f/ms)", secondsTaken, msTaken);
	}
	
	return 0;
}
