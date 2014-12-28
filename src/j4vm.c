#include "j4vm.h"

object *create_object() {
	object *obj = calloc(1, sizeof(object));
	if (!obj) {
		perror("calloc: failed to calloc object");
		exit(1);
	}
	obj->type = '\0';
	obj->ref_count = 0;
	return obj;
}

object *retain_object(object *obj) {
	obj->ref_count++;
	return obj;
}

void release_object(object *obj) {
	if (!obj) {
		debug_message("failed to release object, invalid object specified", false); 
		obj->ref_count--;
		return;
	}
	if (obj->ref_count <= 0) {
		free(obj);
	}
}

object *get_current_stack_item(jayfor_vm *vm) {
	object *current = vm->stack_pointer;
	return current;
}

object *vm_pop_stack(jayfor_vm *vm) {
	vm->stack_pointer--;
	return *vm->stack_pointer;
}

object *get_local(jayfor_vm *vm) {
	return get_local_at_index(vm, *vm->instructions);
}

object *get_local_at_index(jayfor_vm *vm, int index) {
	return vm->locals[index];
}

void set_local_at_index(jayfor_vm *vm, object *obj, int index) {
	vm->locals[index] = obj;
}

void set_local(jayfor_vm *vm, object *obj) {
	set_local_at_index(vm, obj, *vm->instructions);
}

void vm_push_object(jayfor_vm *vm, object *obj) {
	assert(vm->stack_pointer - vm->stack < MAX_STACK_COUNT);
	*(++vm->stack_pointer) = VALUE;		
	retain_object(*vm->stack_pointer);
}

jayfor_vm *create_jayfor_vm() {
	jayfor_vm *vm = malloc(sizeof(*vm));
	true_object = retain_object(create_object());
	false_object = retain_object(create_object());
	null_object = retain_object(create_object());
	return vm;
}

void start_jayfor_vm(jayfor_vm *jvm, int *instructions) {
	jvm->instructions = instructions;
	jvm->self = create_object();
	jvm->stack_pointer = jvm->stack;
	retain_object(jvm->self);

	while (true) {
		switch (*jvm->instructions) {
			case PUSH_NUMBER: {
				jvm->instructions++;
				STACK_PUSH(jvm, create_number((long) *jvm->instructions));
				break;
			}
			case PUSH_STRING: {
				// todo
				break;
			}
			case PUSH_SELF: {
				vm_push_object(jvm, jvm->self);
				break;
			}
			case PUSH_NULL: {
				vm_push_object(jvm, null_object);
				break;
			}
			case PUSH_BOOL: {
				jvm->instructions++;
				vm_push_object(jvm, *jvm->instructions ? true_object : false_object);
				break;
			}
			case CALL: {
				// todo
				break;
			}
			case HALT: {
				// not a very semantic opcode, TODO improve
				goto cleanup;
				break;
			}
			case GET_LOCAL: {
				jvm->instructions++;
				vm_push_object(jvm, jvm->locals[*jvm->instructions]);
				break;
			}
			case SET_LOCAL: {
				jvm->instructions++;
				jvm->locals[*jvm->instructions] = STACK_POP(jvm);
				break;
			}
			case ADD: {
				object *a = STACK_POP(jvm);
				object *b = STACK_POP(jvm);
				STACK_PUSH(jvm, create_number(number_value(a) + number_value(b)));
				release_object(a);
				release_object(b);
				break;
			}
			case JUMP_UNLESS: {
				jvm->instructions++;
				int offset = *jvm->instructions;
				object *condition = STACK_POP(jvm);
				if (!object_is_true(condition)) jvm->instructions += offset;
				release_object(condition);
				break;
			}
			case JUMP: {
				jvm->instructions++;
				int offset = *jvm->instructions;
				jvm->instructions += offset;
				break;
			}
			default: {
				printf("unknown instruction\n");
				exit(1);
			}
		}
		jvm->instructions++;
	}

	cleanup:
	debug_message("cleaning up virtual machine", false);
	release_object(jvm->self);
	int i;
	for (i = 0; i < MAX_LOCAL_COUNT; i++) {
		if (jvm->locals[i]) {
			debug_message("releasing local object", false);
			release_object(jvm->locals[i]);
		}
	}
	debug_message("clearing stack", false);
	while (jvm->stack_pointer > jvm->stack) {
		release_object(STACK_POP(jvm));
	}
}

void destroy_jayfor_vm(jayfor_vm *jvm) {
	free(jvm);
	jvm = NULL;
}

bool object_is_true(object *obj) {
	if (obj == false_object || obj == null_object) { 
		return false;
	}
	return true;
}

int number_value(object *obj) {
	assert(obj->type == type_number);
	return obj->value.number;
}

object *create_number(int value) {
	object *obj = create_object();
	obj->type = type_number;
	obj->value.number = value;
	return obj;
}

object *create_string(char *value) {
	object *obj = create_object();
	obj->type = type_string;
	obj->value.string = value;
	return obj;
}
