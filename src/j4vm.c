#include "j4vm.h"

JayforVM *createJayforVM() {
	JayforVM *vm = malloc(sizeof(*vm));
	vm->bytecode = NULL;
	vm->instructionPointer = 0;
	vm->framePointer = -1;
	vm->stack = createStack();
	vm->running = true;
	vm->globals = NULL;
	return vm;
}

void startJayforVM(JayforVM *vm, int *bytecode, int globalCount) {
	vm->bytecode = bytecode;
	vm->globals = malloc(sizeof(*vm->globals) * globalCount);

	// for arithmetic operations
	int *a = NULL;
	int *b = NULL;
	int c = 0;
	int offset = 0;
	int address = 0;

	while (vm->running) {
		int op = vm->bytecode[vm->instructionPointer++];

		switch (op) {
			case ADD:
				a = popStack(vm->stack);
				b = popStack(vm->stack);
				c = *a + *b;
				pushToStack(vm->stack, &c);
				break;
			case SUB:
				a = popStack(vm->stack);
				b = popStack(vm->stack);
				c = *a - *b;
				pushToStack(vm->stack, &c);
				break;
			case MUL:
				a = popStack(vm->stack);
				b = popStack(vm->stack);
				c = *a * *b; // ****************
				pushToStack(vm->stack, &c);
				break;
			case ICONST:
				pushToStack(vm->stack, &vm->bytecode[vm->instructionPointer++]);
				break;
			case LOAD:
				offset = vm->bytecode[vm->instructionPointer++];
				pushToStack(vm->stack, getValueFromStack(vm->stack, vm->framePointer + offset));
				break;
			case GLOAD:
				address = vm->bytecode[vm->instructionPointer++];
				pushToStack(vm->stack, &vm->globals[address]);				
				break;
			case STORE:
				offset = vm->bytecode[vm->instructionPointer++];
				pushToStackAtIndex(vm->stack, popStack(vm->stack), vm->framePointer + offset);
				break;
			case GSTORE:
				address = vm->bytecode[vm->instructionPointer++];
				vm->globals[address] = (int) popStack(vm->stack);
				break;
			case POP:
				popStack(vm->stack);
				break;
			case HALT:
				vm->running = false;
				break;
			default:
				printf("error: unknown instruction %d. Halting...\n", op);
				exit(1);
				break;
		}
	}
}

void destroyJayforVM(JayforVM *vm) {
	if (vm != NULL) {
		free(vm);
		vm = NULL;
	}
}