#include "j4vm.h"

JayforVM *createJ4VM() {
	JayforVM *vm = malloc(sizeof(*vm));
	vm->bytecode = NULL;
	vm->instructionPointer = 0;
	vm->framePointer = -1;
	vm->stack = createStack();
	vm->running = true;
	return vm;
}

void startJ4VM(JayforVM *vm, int *bytecode) {
	vm->bytecode = bytecode;

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
			case RET:

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

void destroyJ4VM(JayforVM *vm) {
	if (vm != NULL) {
		free(vm);
		vm = NULL;
	}
}