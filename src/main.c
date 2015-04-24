#include <stdio.h>

#include "alloyc.h"

int main(int argc, char** argv) {

	// alloy compiler stuff
	AlloyCompiler *alloyc = createAlloyCompiler(argc, argv);
	if (alloyc) {
		// start the timer
		clock_t timer = clock();
		startAlloyCompiler(alloyc);
		destroyAlloyCompiler(alloyc);

		// finished timer
		timer = clock() - timer;	// calculate time taken

		// calculate seconds and milliseconds taken
		double secondsTaken = ((double) timer) / CLOCKS_PER_SEC;
		double msTaken = secondsTaken * 1000;

		// write the message
		primaryMessage("Finished in %.6f/s (%.0f/ms)", secondsTaken, msTaken);
	}

	return 0;
}
