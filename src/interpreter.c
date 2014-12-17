#include "interpreter.h"

Interpreter *createInterpreter() {
	Interpreter *interpreter = malloc(sizeof(*interpreter));

	return interpreter;
}

void startInterpreter(Vector *ast) {

}

void destroyInterpreter(Interpreter *interpreter) {
	if (interpreter != NULL) {
		free(interpreter);
		interpreter = NULL;
	}
}
