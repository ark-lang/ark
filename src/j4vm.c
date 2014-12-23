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
	vm->globals = NULL;
	vm->stack = NULL;
	
	vm->instruction_pointer = 0;
	vm->frame_pointer = -1;
	vm->stack_pointer = -1;
	vm->default_global_space = 32;
	vm->default_stack_size = 32;

	vm->running = true;
	return vm;
}

static void print_instruction(int *code, int ip) {
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

static void push_value(jayfor_vm *vm, int value, int index) {
	if (index > vm->default_stack_size) {
		vm->default_stack_size *= 2;
		int *temp = realloc(vm->stack, sizeof(*vm->stack) * vm->default_stack_size);
		if (!temp) {
			perror("realloc: failed to reallocate memory for stack");
			exit(1);
		}
		else {
			vm->stack = temp;
		}
	}
	vm->stack[index] = value;
}

void start_jayfor_vm(jayfor_vm *vm, int *bytecode, int bytecode_size) {
	vm->bytecode = bytecode;
	vm->globals = malloc(sizeof(*vm->globals) * vm->default_global_space);
	vm->stack = malloc(sizeof(*vm->stack) * vm->default_stack_size);

	while (vm->running) {
		if (DEBUG_MODE) print_instruction(vm->bytecode, vm->instruction_pointer);
		int op = vm->bytecode[vm->instruction_pointer++];

		switch (op) {
			case ADD: {
				int a = vm->stack[vm->stack_pointer--];
				int b = vm->stack[vm->stack_pointer--];
				push_value(vm, a + b, ++vm->stack_pointer);
				break;
			}
			case SUB: {
				int a = vm->stack[vm->stack_pointer--];
				int b = vm->stack[vm->stack_pointer--];
				push_value(vm, a - b, ++vm->stack_pointer);
				break;
			}	
			case MUL: {
				int a = vm->stack[vm->stack_pointer--];
				int b = vm->stack[vm->stack_pointer--];
				push_value(vm, a * b, ++vm->stack_pointer);
				break;
			}
			case DIV: {
				int a = vm->stack[vm->stack_pointer--];
				int b = vm->stack[vm->stack_pointer--];
				push_value(vm, a / b, ++vm->stack_pointer);
				break;
			}
			case MOD: {
				int a = vm->stack[vm->stack_pointer--];
				int b = vm->stack[vm->stack_pointer--];
				push_value(vm, a % b, ++vm->stack_pointer);
				break;
			}
			case POW: {
				int a = vm->stack[vm->stack_pointer--];
				int b = vm->stack[vm->stack_pointer--];
				push_value(vm, pow(a, b), ++vm->stack_pointer);
				break;
			}
			case RET: {
				int return_value = vm->stack[vm->stack_pointer--];
				vm->stack_pointer = vm->frame_pointer;
				vm->instruction_pointer = vm->stack[vm->stack_pointer--];
				vm->frame_pointer = vm->stack[vm->stack_pointer--];
				int number_of_args = vm->stack[vm->stack_pointer--];
				vm->stack_pointer -= number_of_args;
				push_value(vm, return_value, ++vm->stack_pointer);
				break;
			}
			case PRINT: {
				int x = vm->stack[vm->stack_pointer--];
				printf("%d, %04X\n", x, x);
				break;
			}
			case CALL: {
				int address = vm->bytecode[vm->instruction_pointer++];
				int number_of_args = vm->bytecode[vm->instruction_pointer++];
				push_value(vm, number_of_args, ++vm->stack_pointer);
				push_value(vm, vm->frame_pointer, ++vm->stack_pointer);
				push_value(vm, vm->instruction_pointer, ++vm->stack_pointer);
				vm->frame_pointer = vm->stack_pointer;
				vm->instruction_pointer = address;
				break;
			}
			case ICONST: {
				push_value(vm, vm->bytecode[vm->instruction_pointer++], ++vm->stack_pointer);
				break;
			}
			case LOAD: {
				int offset = vm->bytecode[vm->instruction_pointer++];
				push_value(vm, vm->stack[vm->frame_pointer + offset], ++vm->stack_pointer);
				break;
			}
			case GLOAD: {
				int address = vm->bytecode[vm->instruction_pointer++];
				push_value(vm, vm->globals[address], ++vm->stack_pointer);
				break;
			}
			case STORE: {
				int offset = vm->bytecode[vm->instruction_pointer++];
				push_value(vm, vm->stack[vm->stack_pointer--], vm->frame_pointer + offset);
				break;
			}
			case GSTORE: {
				int address = vm->bytecode[vm->instruction_pointer++];
				if (vm->default_global_space >= address) {
					vm->default_global_space *= 2;
					int *temp = realloc(vm->globals, sizeof(*vm->globals) * vm->default_global_space);
					if (!temp) {
						perror("realloc: failed to allocate memory for globals");
						exit(1);
					}
					else {
						vm->globals = temp;
					}
				}
				vm->globals[address] = vm->stack[vm->stack_pointer--];
				break;
			}
			case POP: {
				vm->stack_pointer--;
				break;
			}
			case HALT: {
				vm->running = false;
				break;
			}
			default: {
				printf("error: unknown instruction %d. Halting...\n", op);
				exit(1);
				break;
			}
		}
	}
}

void destroy_jayfor_vm(jayfor_vm *vm) {
	if (vm != NULL) {
		free(vm);
		vm = NULL;
	}
}