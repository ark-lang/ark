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
	vm->default_global_space = 32;
	vm->running = true;
	vm->globals = NULL;
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

void start_jayfor_vm(jayfor_vm *vm, int *bytecode) {
	vm->bytecode = bytecode;
	vm->globals = malloc(sizeof(*vm->globals) * vm->default_global_space);

	while (vm->running) {
		if (DEBUG_MODE) {
			print_instruction(vm->bytecode, vm->instruction_pointer);
		}
		int op = vm->bytecode[vm->instruction_pointer++];

		switch (op) {
			case ADD: {
				int *a = pop_stack(vm->stack);
				int *b = pop_stack(vm->stack);
				int c = *a + *b;
				push_to_stack(vm->stack, &c);
				break;
			}
			case SUB: {
				int *a = pop_stack(vm->stack);
				int *b = pop_stack(vm->stack);
				int c = *a - *b;
				push_to_stack(vm->stack, &c);
				break;
			}	
			case MUL: {
				int *a = pop_stack(vm->stack);
				int *b = pop_stack(vm->stack);
				int c = *a * *b;
				push_to_stack(vm->stack, &c);
				break;
			}
			case DIV: {
				int *a = pop_stack(vm->stack);
				int *b = pop_stack(vm->stack);
				int c = *a / *b;
				push_to_stack(vm->stack, &c);	
				break;
			}
			case MOD: {
				int *a = pop_stack(vm->stack);
				int *b = pop_stack(vm->stack);
				int c = *a % *b;
				push_to_stack(vm->stack, &c);
				break;
			}
			case POW: {
				int *a = pop_stack(vm->stack);
				int *b = pop_stack(vm->stack);
				int c = *a ^ *b;
				push_to_stack(vm->stack, &c);
				break;
			}
			case RET: {
				int *ret = (int*) pop_stack(vm->stack);
				vm->stack->stack_pointer = vm->frame_pointer;
					
				int *ip = pop_stack(vm->stack);
				int *fp = pop_stack(vm->stack);
				vm->instruction_pointer = *ip;
				vm->frame_pointer = *fp;

				int *numOfArgs = pop_stack(vm->stack);
				vm->stack->stack_pointer -= *numOfArgs;
				push_to_stack(vm->stack, ret);
				break;
			}
			case PRINT: {
				int *x = pop_stack(vm->stack);
				printf("%d\n", *x);
				break;
			}
			case CALL: {
				int address = vm->bytecode[vm->instruction_pointer++];
				int numOfArgs = vm->bytecode[vm->instruction_pointer++];
				push_to_stack(vm->stack, &numOfArgs);
				push_to_stack(vm->stack, &vm->frame_pointer);
				push_to_stack(vm->stack, &vm->instruction_pointer);
				vm->frame_pointer = vm->stack->stack_pointer;
				vm->instruction_pointer = address;
				break;
			}
			case ICONST: {
				int x = vm->bytecode[vm->instruction_pointer++];
				push_to_stack(vm->stack, &x);
				break;
			}
			case LOAD: {
				int offset = vm->bytecode[vm->instruction_pointer++];
				push_to_stack(vm->stack, get_value_from_stack(vm->stack, vm->frame_pointer + offset));
				break;
			}
			case GLOAD: {
				int address = vm->bytecode[vm->instruction_pointer++];
				push_to_stack(vm->stack, &vm->globals[address]);				
				break;
			}
			case STORE: {
				int offset = vm->bytecode[vm->instruction_pointer++];
				push_to_stack_at_index(vm->stack, pop_stack(vm->stack), vm->frame_pointer + offset);
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
				int *value = pop_stack(vm->stack);
				vm->globals[address] = *value;
				break;
			}
			case POP: {
				pop_stack(vm->stack);
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