#include <stdio.h>
#include <time.h>

#include "jayfor.h"

int main(int argc, char** argv) {

	int i;
	for (i = 0; i < argc; i++) {
		printf("%s\n", argv[i]);
	}

	// for calculating time taken
	clock_t timer = clock();

	// jayfor stuff
	Jayfor *jayfor = jayforCreate(argc, argv);
	jayforStart(jayfor);

	printf("THIS IS THE FILE CONTENT BITCH:\n");
	printf("%s\n\n\n", jayfor->scanner->content);e

	jayforDestroy(jayfor);

	// calculate time
	timer = clock() - timer;
	double timeTaken = ((double) timer) / CLOCKS_PER_SEC;
	printf("Finished in %f/s\n", timeTaken);

	return 0;
}
