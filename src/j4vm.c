#include "j4vm.h"

instruction debug_instructions[] = {
    { "add",    0 },
    { "sub",    0 },
    { "mul",    0 },
    { "div",    0 },
    { "mod",    0 },
    { "ret",    0 },
    { "print",  0 },
    { "call",   2 },
    { "iconst", 1 },
    { "fconst", 1 },
    { "load",   1 },
    { "gload",  1 },
    { "store",  1 },
    { "gstore", 1 },
    { "ilt",    0 },
    { "ieq",    0 },
    { "brf",    1 },
    { "brt",    1 },
    { "pop",    0 },
    { "halt",   0 }
};

int float_to_int_bits(float x) {
	union {
		float f; // 32 bit IEEE 754 single-precision
		int i;   // 32 bit 2's complement int
	} u;
	if (isnan(x)) {
		return 0x7fc00000;
	}
	else {
		u.f = x;
		return u.i;
	}
}

float int_bits_to_float(int x) {
	union {
		float f; // 32 bit IEEE 754 single-precision
		int i;   // 32 bit 2's complement int
	} u;
	u.i = x;
	return u.f;
}

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
    switch (instruction->number_of_args) {
    	case 0:
	        printf("%04d:  %-20s\n", ip, instruction->name);
	        break;
    	case 1:
	        printf("%04d:  %-10s%-10d\n", ip, instruction->name, code[ip + 1]);
	        break;
    }
}

static void push_value(jayfor_vm *vm, object value, int index) {
	if (index > vm->default_stack_size) {
		vm->default_stack_size *= 2;
		object *temp = realloc(vm->stack, sizeof(*vm->stack) * vm->default_stack_size);
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

static object pop_stack(jayfor_vm *vm) {
	return vm->stack[vm->stack_pointer--];
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
				object b = pop_stack(vm);
				object a = pop_stack(vm);
				
				object result;
				if (b.type == TYPE_FLOAT || a.type == TYPE_FLOAT)
					result.type = TYPE_FLOAT;
				else
					result.type = TYPE_INT;
				result.data = b.data + a.data;
				
				push_value(vm, result, ++vm->stack_pointer);
				break;
			}
			case SUB: {
				object b = pop_stack(vm);
				object a = pop_stack(vm);

				object result;
				if (b.type == TYPE_FLOAT || a.type == TYPE_FLOAT)
					result.type = TYPE_FLOAT;
				else
					result.type = TYPE_INT;
				result.data = b.data - a.data;

				push_value(vm, result, ++vm->stack_pointer);
				break;
			}	
			case MUL: {
				object b = pop_stack(vm);
				object a = pop_stack(vm);

				object result;
				if (b.type == TYPE_FLOAT || a.type == TYPE_FLOAT)
					result.type = TYPE_FLOAT;
				else
					result.type = TYPE_INT;
				result.data = b.data * a.data;

				push_value(vm, result, ++vm->stack_pointer);
				break;
			}
			case DIV: {
				object b = pop_stack(vm);
				object a = pop_stack(vm);

				object result;
				if (b.type == TYPE_FLOAT || a.type == TYPE_FLOAT)
					result.type = TYPE_FLOAT;
				else
					result.type = TYPE_INT;
				result.data = b.data / a.data;

				push_value(vm, result, ++vm->stack_pointer);
				break;
			}
			case MOD: {
				object b = pop_stack(vm);
				object a = pop_stack(vm);

				object result;
				if (b.type == TYPE_FLOAT || a.type == TYPE_FLOAT)
					result.type = TYPE_FLOAT;
				else
					result.type = TYPE_INT;
				result.data = b.data % a.data;

				push_value(vm, result, ++vm->stack_pointer);
				break;
			}
			case RET: {
				object return_value = pop_stack(vm);
				vm->stack_pointer = vm->frame_pointer;
				vm->instruction_pointer = pop_stack(vm).data;
				vm->frame_pointer = pop_stack(vm).data;
				object number_of_args = pop_stack(vm);
				vm->stack_pointer -= number_of_args.data;
				push_value(vm, return_value, ++vm->stack_pointer);
				break;
			}
			case PRINT: {
				object x = pop_stack(vm);
				if (x.type == TYPE_FLOAT) {
					printf("float: %f\n", int_bits_to_float(x.data));
				}
				else {
					printf("int: %d at %d\n", x.data, vm->stack_pointer);
				}
				break;
			}
			case CALL: {
				object address;
				address.type = TYPE_INT;
				address.data = vm->bytecode[vm->instruction_pointer++];

				object number_of_args;
				number_of_args.type = TYPE_INT;
				number_of_args.data = vm->bytecode[vm->instruction_pointer++];

				push_value(vm, number_of_args, ++vm->stack_pointer);

				object frame_pointer_store;
				frame_pointer_store.type = TYPE_INT;
				frame_pointer_store.data = vm->frame_pointer;
				push_value(vm, frame_pointer_store, ++vm->stack_pointer);

				object instruction_pointer_store;
				instruction_pointer_store.type = TYPE_INT;
				instruction_pointer_store.data = vm->instruction_pointer;
				push_value(vm, instruction_pointer_store, ++vm->stack_pointer);

				vm->frame_pointer = vm->stack_pointer;
				vm->instruction_pointer = address.data;
				break;
			}
			case ICONST: {
				object value;
				value.type = TYPE_INT;
				value.data = vm->bytecode[vm->instruction_pointer++];
				push_value(vm, value, ++vm->stack_pointer);
				break;
			}
			case FCONST: {
				object value;
				value.type = TYPE_FLOAT;
				value.data = vm->bytecode[vm->instruction_pointer++];
				push_value(vm, value, ++vm->stack_pointer);
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
				object original = pop_stack(vm);
				push_value(vm, original, vm->frame_pointer + vm->bytecode[vm->instruction_pointer++]);
				break;
			}
			case GSTORE: {
				int address = vm->bytecode[vm->instruction_pointer++];
				if (vm->default_global_space >= address) {
					vm->default_global_space *= 2;
					object *temp = realloc(vm->globals, sizeof(*vm->globals) * vm->default_global_space);
					if (!temp) {
						perror("realloc: failed to allocate memory for globals");
						exit(1);
					}
					else {
						vm->globals = temp;
					}
				}
				vm->globals[address] = pop_stack(vm);
				break;
			}
			case ILT: {
				object b = pop_stack(vm);
				object a = pop_stack(vm);

				// store it anyway
				float bf = b.data;
				float af = a.data;

				// check for floats and convert accordingly
				if (b.type == TYPE_FLOAT) {
					bf = int_bits_to_float(b.data);
				}
				if (a.type == TYPE_FLOAT) {
					af = int_bits_to_float(a.data);
				}

				object result;
				result.type = TYPE_INT;
				result.data = af < bf ? true : false;

				push_value(vm, result, ++vm->stack_pointer);
			}
			case IEQ: {
				object b = pop_stack(vm);
				object a = pop_stack(vm);

				// store it anyway
				float bf = b.data;
				float af = a.data;

				// check for floats and convert accordingly
				if (b.type == TYPE_FLOAT) {
					bf = int_bits_to_float(b.data);
				}
				if (a.type == TYPE_FLOAT) {
					af = int_bits_to_float(a.data);
				}

				object result;
				result.type = TYPE_INT;
				result.data = af == bf ? true : false;

				push_value(vm, result, ++vm->stack_pointer);
			}
			case BRT: {
				int address = vm->bytecode[vm->instruction_pointer++];
				if (pop_stack(vm).data) {
					vm->instruction_pointer = address;
				}
			}
			case BRF: {
				int address = vm->bytecode[vm->instruction_pointer++];
				if (!pop_stack(vm).data) {
					vm->instruction_pointer = address;
				}
			}
			case POP: {
				pop_stack(vm);
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