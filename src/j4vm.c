#include "j4vm.h"

static Instruction debugInstructions[] = {
    { "add",    0 },
    { "sub",    0 },
    { "mul",    0 },
    { "div",    0 },
    { "mod",    0 },
    { "pow",    0 },
    { "ret",    0 },
    { "call",   2 },
    { "iconst", 1 },
    { "load",   1 },
    { "gload",  1 },
    { "store",  1 },
    { "gstore", 1 },
    { "pop",    0 },
    { "halt",   0 }
};

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

static void printInstruction(int *code, int ip) {
    int opcode = code[ip];
    if (opcode > HALT) return;
    Instruction *instruction = &debugInstructions[opcode];
    switch (instruction->numOfArgs) {
    	case 0:
	        printf("%04d:  %-20s\n", ip, instruction->name);
	        break;
    	case 1:
	        printf("%04d:  %-10s%-10d\n", ip, instruction->name, code[ip + 1]);
	        break;
    }
}

void startJayforVM(JayforVM *vm, int *bytecode, int globalCount, int entryPoint) {
	vm->bytecode = bytecode;
	vm->globals = malloc(sizeof(*vm->globals) * globalCount);
	vm->instructionPointer = entryPoint;

	// for arithmetic operations
	int *a = NULL;
	int *b = NULL;
	int c = 0;
	int offset = 0;
	int address = 0;
	int numOfArgs = 0;
	int ret = 0;

	while (vm->running) {
		printInstruction(vm->bytecode, vm->instructionPointer);
		
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
			case DIV:
				a = popStack(vm->stack);
				b = popStack(vm->stack);
				c = *a / *b; // ****************
				pushToStack(vm->stack, &c);
				break;
			case MOD:
				a = popStack(vm->stack);
				b = popStack(vm->stack);
				c = *a % *b; // ****************
				pushToStack(vm->stack, &c);
				break;
			case POW:
				a = popStack(vm->stack);
				b = popStack(vm->stack);
				c = *a ^ *b; // ****************
				pushToStack(vm->stack, &c);
				break;
			case RET:
				ret = (int) popStack(vm->stack);
				vm->stack->stackPointer = vm->framePointer;
				vm->instructionPointer = (int) popStack(vm->stack);
				vm->framePointer =(int) popStack(vm->stack);
				numOfArgs = (int) popStack(vm->stack);
				vm->stack->stackPointer -= numOfArgs;
				pushToStack(vm->stack, &ret);
				break;
			case CALL:
				address = vm->bytecode[vm->instructionPointer++];
				numOfArgs = vm->bytecode[vm->instructionPointer++];
				pushToStack(vm->stack, &numOfArgs);
				pushToStack(vm->stack, &vm->framePointer);
				pushToStack(vm->stack, &vm->instructionPointer);
				vm->framePointer = vm->stack->stackPointer;
				vm->instructionPointer = address;
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