#include <stdio.h>

#include "compiler.h"

int main(int argc, char** argv) {
	Compiler *compiler = createCompiler(argc, argv);
	if (compiler) {
		// start the timer
		clock_t timer = clock();

		startCompiler(compiler);
		destroyCompiler(compiler);

		// finished timer
		timer = clock() - timer;

		// calculate seconds and milliseconds taken
		double secondsTaken = ((double) timer) / CLOCKS_PER_SEC;
		double msTaken = secondsTaken * 1000;

		primaryMessage("Finished in %.6f/s (%.0f/ms)", secondsTaken, msTaken);
	}

	return 0;
}
