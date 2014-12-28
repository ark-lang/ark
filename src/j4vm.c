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

object *vm_peek(jayfor_vm *vm) {
	object *current = *vm->stack_pointer;
	return current;
}

void vm_set(jayfor_vm *vm, object *obj) {
	*vm->stack_pointer = obj;
}

object *vm_pop(jayfor_vm *vm) {
	vm->stack_pointer--;
	return vm_peek(vm);
}

void vm_push(jayfor_vm *vm, object *obj) {
	if (vm->stack_pointer - vm->stack > MAX_STACK_COUNT) {
		error_message("stack overflow");
	}
	++vm->stack_pointer;
	vm_set(vm, obj);		
	retain_object(vm_peek(vm));
}

object *get_local(jayfor_vm *vm) {
	return get_local_at_index(vm, *vm->instructions);
}

object *get_local_at_index(jayfor_vm *vm, int index) {
	if (index < 0 || index > MAX_LOCAL_COUNT) {
		error_message("error: local index %d is out of bounds", index);
	}
	return vm->locals[index];
}

void set_local_at_index(jayfor_vm *vm, object *obj, int index) {
	if (index < 0 || index > MAX_LOCAL_COUNT) {
		error_message("error: local index %d is out of bounds", index);
	}
	vm->locals[index] = obj;
}

void set_local(jayfor_vm *vm, object *obj) {
	set_local_at_index(vm, obj, *vm->instructions);
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
				vm_push(jvm, create_number((long) *jvm->instructions));
				break;
			}
			case PUSH_STRING: {
				// todo
				break;
			}
			case PUSH_SELF: {
				vm_push(jvm, jvm->self);
				break;
			}
			case PUSH_NULL: {
				vm_push(jvm, null_object);
				break;
			}
			case PUSH_BOOL: {
				jvm->instructions++;
				vm_push(jvm, *jvm->instructions ? true_object : false_object);
				break;
			}
			case CALL: {
				// todo
				break;
			}
			case HALT: {
				goto cleanup;
				break;
			}
			case GET_LOCAL: {
				jvm->instructions++;
				vm_push(jvm, get_local(jvm));
				break;
			}
			case SET_LOCAL: {
				jvm->instructions++;
				set_local(jvm, vm_pop(jvm));
				break;
			}
			case ADD: {
				object *a = vm_pop(jvm);
				object *b = vm_pop(jvm);
				vm_push(jvm, create_number(number_value(a) + number_value(b)));
				release_object(a);
				release_object(b);
				break;
			}
			case JUMP_UNLESS: {
				jvm->instructions++;
				int offset = *jvm->instructions;
				object *condition = vm_pop(jvm);
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
	debug_message("cleaning up virtual machine\n");
	release_object(jvm->self);
	int i;
	for (i = 0; i < MAX_LOCAL_COUNT; i++) {
		if (!jvm->locals[i]) {
			debug_message("releasing local object\n");
			release_object(jvm->locals[i]);
		}
	}
	debug_message("clearing stack\n");
	while (jvm->stack_pointer > jvm->stack) {
		release_object(vm_pop(jvm));
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
