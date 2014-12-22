#include "j4vm.h"

instruction debug_instructions[] = {
    { "add",    0 },
    { "sub",    0 },
    { "mul",    0 },
    { "div",    0 },
    { "mod",    0 },
    { "pow",    0 },
    { "ret",    0 },
    { "print",  0 },
    { "call",   2 },
    { "iconst", 1 },
    { "load",   1 },
    { "gload",  1 },
    { "store",  1 },
    { "gstore", 1 },
    { "pop",    0 },
    { "halt",   0 }
};

jayfor_vm *create_jayfor_vm() {
	jayfor_vm *vm = malloc(sizeof(*vm));
	vm->bytecode = NULL;
	vm->instruction_pointer = 0;
	vm->frame_pointer = -1;
	vm->stack = create_stack();
	vm->running = true;
	vm->globals = NULL;
	return vm;
}

static void printinstruction(int *code, int ip) {
    int opcode = code[ip];
    if (opcode > HALT) return;
    instruction *instruction = &debug_instructions[opcode];
    switch (instruction->numOfArgs) {
    	case 0:
	        printf("%04d:  %-20s\n", ip, instruction->name);
	        break;
    	case 1:
	        printf("%04d:  %-10s%-10d\n", ip, instruction->name, code[ip + 1]);
	        break;
    }
}

static void programDump(jayfor_vm *vm) {
	printf(KYEL "\nDUMPING PROGRAM CONTENTS\n");
	int i;
	for (i = 0; i< vm->instruction_pointer; i++) {
		printinstruction(vm->bytecode, i);
	}
	printf("\n" KNRM);
}

// this is a utility to make it easier for the VM
static int popValueFromStack(jayfor_vm *vm, Stack *stk) {
	int *x = pop_stack(stk);
	if (!x) {
		if (DEBUG_MODE) {
			printinstruction(vm->bytecode, vm->instruction_pointer);
			programDump(vm);
		}
		exit(1);
	}
	return *x;
}

void start_jayfor_vm(jayfor_vm *vm, int *bytecode, int global_count) {
	vm->bytecode = bytecode;
	vm->globals = malloc(sizeof(*vm->globals) * global_count);

	// for arithmetic operations
	int a = 0;
	int b = 0;
	int c = 0;
	int offset = 0;
	int address = 0;
	int numOfArgs = 0;
	int ret = 0;

	while (vm->running) {
		if (DEBUG_MODE) printinstruction(vm->bytecode, vm->instruction_pointer);
		
		int op = vm->bytecode[vm->instruction_pointer++];

		switch (op) {
			case ADD:
				a = popValueFromStack(vm, vm->stack);
				b = popValueFromStack(vm, vm->stack);
				c = a + b;
				push_to_stack(vm->stack, &c);
				break;
			case SUB:
				a = popValueFromStack(vm, vm->stack);
				b = popValueFromStack(vm, vm->stack);
				c = a - b;
				push_to_stack(vm->stack, &c);
				break;
			case MUL:
				a = popValueFromStack(vm, vm->stack);
				b = popValueFromStack(vm, vm->stack);
				c = a * b;
				push_to_stack(vm->stack, &c);
				break;
			case DIV:
				a = popValueFromStack(vm, vm->stack);
				b = popValueFromStack(vm, vm->stack);
				c = a / b;
				push_to_stack(vm->stack, &c);
				break;
			case MOD:
				a = popValueFromStack(vm, vm->stack);
				b = popValueFromStack(vm, vm->stack);
				c = a % b;
				push_to_stack(vm->stack, &c);
				break;
			case POW:
				a = popValueFromStack(vm, vm->stack);
				b = popValueFromStack(vm, vm->stack);
				c = a ^ b;
				push_to_stack(vm->stack, &c);
				break;
			case RET:
				ret = popValueFromStack(vm, vm->stack);
				vm->stack->stackPointer = vm->frame_pointer;
				vm->instruction_pointer = popValueFromStack(vm, vm->stack);
				vm->frame_pointer = popValueFromStack(vm, vm->stack);
				numOfArgs = popValueFromStack(vm, vm->stack);
				vm->stack->stackPointer -= numOfArgs;
				push_to_stack(vm->stack, &ret);
				break;
			case PRINT:
				printf("%d\n", popValueFromStack(vm, vm->stack));
				break;
			case CALL:
				address = vm->bytecode[vm->instruction_pointer++];
				numOfArgs = vm->bytecode[vm->instruction_pointer++];
				push_to_stack(vm->stack, &numOfArgs);
				push_to_stack(vm->stack, &vm->frame_pointer);
				push_to_stack(vm->stack, &vm->instruction_pointer);
				vm->frame_pointer = vm->stack->stackPointer;
				vm->instruction_pointer = address;
				break;
			case ICONST:
				push_to_stack(vm->stack, &vm->bytecode[vm->instruction_pointer++]);
				break;
			case LOAD:
				offset = vm->bytecode[vm->instruction_pointer++];
				push_to_stack(vm->stack, get_value_from_stack(vm->stack, vm->frame_pointer + offset));
				break;
			case GLOAD:
				address = vm->bytecode[vm->instruction_pointer++];
				push_to_stack(vm->stack, &vm->globals[address]);				
				break;
			case STORE:
				offset = vm->bytecode[vm->instruction_pointer++];
				push_to_stack_at_index(vm->stack, pop_stack(vm->stack), vm->frame_pointer + offset);
				break;
			case GSTORE:
				address = vm->bytecode[vm->instruction_pointer++];
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

void destroy_jayfor_vm(jayfor_vm *vm) {
	if (vm != NULL) {
		free(vm);
		vm = NULL;
	}
}