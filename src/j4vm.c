#include "j4vm.h"

static Instruction debugInstructions[] = {
    { "add",    0 },
    { "sub",    0 },
    { "mul",    0 },
    { "div",    0 },
    { "mod",    0 },
    { "pow",    0 },
    { "ret",    0 },
    { "print",    0 },
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

static void programDump(JayforVM *vm) {
	printf(KYEL "\nDUMPING PROGRAM CONTENTS\n");
	int i;
	for (i = 0; i< vm->instructionPointer; i++) {
		printInstruction(vm->bytecode, i);
	}
	printf("\n" KNRM);
}

// this is a utility to make it easier for the VM
static int popValueFromStack(JayforVM *vm, Stack *stk) {
	int *x = popStack(stk);
	if (!x) {
		if (DEBUG_MODE) {
			printInstruction(vm->bytecode, vm->instructionPointer);
			programDump(vm);
		}
		exit(1);
	}
	return *x;
}

void startJayforVM(JayforVM *vm, int *bytecode, int globalCount) {
	vm->bytecode = bytecode;
	vm->globals = malloc(sizeof(*vm->globals) * globalCount);

	// for arithmetic operations
	int a = 0;
	int b = 0;
	int c = 0;
	int offset = 0;
	int address = 0;
	int numOfArgs = 0;
	int ret = 0;

	while (vm->running) {
		if (DEBUG_MODE) printInstruction(vm->bytecode, vm->instructionPointer);
		
		int op = vm->bytecode[vm->instructionPointer++];

		switch (op) {
			case ADD:
				a = popValueFromStack(vm, vm->stack);
				b = popValueFromStack(vm, vm->stack);
				c = a + b;
				pushToStack(vm->stack, &c);
				break;
			case SUB:
				a = popValueFromStack(vm, vm->stack);
				b = popValueFromStack(vm, vm->stack);
				c = a - b;
				pushToStack(vm->stack, &c);
				break;
			case MUL:
				a = popValueFromStack(vm, vm->stack);
				b = popValueFromStack(vm, vm->stack);
				c = a * b;
				pushToStack(vm->stack, &c);
				break;
			case DIV:
				a = popValueFromStack(vm, vm->stack);
				b = popValueFromStack(vm, vm->stack);
				c = a / b;
				pushToStack(vm->stack, &c);
				break;
			case MOD:
				a = popValueFromStack(vm, vm->stack);
				b = popValueFromStack(vm, vm->stack);
				c = a % b;
				pushToStack(vm->stack, &c);
				break;
			case POW:
				a = popValueFromStack(vm, vm->stack);
				b = popValueFromStack(vm, vm->stack);
				c = a ^ b;
				pushToStack(vm->stack, &c);
				break;
			case RET:
				ret = popValueFromStack(vm, vm->stack);
				vm->stack->stackPointer = vm->framePointer;
				vm->instructionPointer = popValueFromStack(vm, vm->stack);
				vm->framePointer = popValueFromStack(vm, vm->stack);
				numOfArgs = popValueFromStack(vm, vm->stack);
				vm->stack->stackPointer -= numOfArgs;
				pushToStack(vm->stack, &ret);
				break;
			case PRINT:
				printf("%d\n", popValueFromStack(vm, vm->stack));
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
				vm->globals[address] = popValueFromStack(vm, vm->stack);
				break;
			case POP:
				popValueFromStack(vm, vm->stack);
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