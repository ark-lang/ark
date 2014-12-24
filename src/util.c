#include "util.h"

extern void debug_message(char *msg, bool exit_on_error) {
	if (DEBUG_MODE) {
		printf("%s %s %s\n", KYEL, msg, KNRM);
		if (exit_on_error) exit(1);
	}
}