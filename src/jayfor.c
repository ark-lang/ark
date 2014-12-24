#include "jayfor.h"

bool DEBUG_MODE = false;
bool EXECUTE_BYTECODE = false;
bool RUN_VM_EXECUTABLE = false;
char *VM_EXECUTABLE_NAME = "";
char *OUTPUT_EXECUTABLE_NAME = "a.j4e";

static void parseArgument(argument *arg) {
	char argument = arg->argument[0];

	switch (argument) {
		case 'v':
			printf("jayfor Version: %s\n", JAYFOR_VERSION);
			exit(1);
			break;
		case 'd':
			DEBUG_MODE = true;
			break;
		case 'e':
			RUN_VM_EXECUTABLE = true;
			VM_EXECUTABLE_NAME = arg->next_argument;
			break;
		case 'h':
			KYEL();
			printf("jayfor Argument List\n");
			printf("\t-h,\t shows a help menu\n");
			printf("\t-v,\t shows current version\n");
			printf("\t-d,\t logs extra debug information\n");
			printf("\t-r,\t will compile and execute code instead of creating an executable\n");
			printf("\t-o <name>.j4e,\t creates an executable with the given file name and extension\n");
			printf("\n");
			KNRM();
			exit(1);
			break;
		case 'o':
			OUTPUT_EXECUTABLE_NAME = arg->next_argument;
			break;
		case 'r':
			EXECUTE_BYTECODE = true;
			break;
		default:
			error_message("error: unrecognized argument %c\n", argument);
			break;
	}
}

jayfor *create_jayfor(int argc, char** argv) {
	// not enough args just throw an error
	if (argc <= 1) {
		error_message("error: no input files\n");
	}
	jayfor *self = malloc(sizeof(*self));
	self->filename = NULL;

	if (!self) {
		perror("malloc: failed to allocate memory for jayfor");
		exit(1);
	}

	int i;
	// i = 1 to ignore first arg
	for (i = 1; i < argc; i++) {
		if (argv[i][0] == '-') {
			argument arg;

			// remove the -
			size_t len = strlen(argv[i]) - 1;
			char temp[len];
			memcpy(temp, &argv[i][len], 1);
			temp[len] = '\0';

			// set argument stuff
			arg.argument = temp;
			arg.next_argument = argv[i + 1];
			
			// multiple arguments needed for -o, consume twice
			if (!strcmp(arg.argument, "o")) {
				i += 2;
			}

			// parse the argument
			parseArgument(&arg);
		}
		else if (strstr(argv[i], ".j4")) {
			self->filename = argv[i];
		}
		else {
			error_message("error: argument not recognized: %s\n", argv[i]);
		}
	}

	self->scanner = NULL;
	self->lexer = NULL;
	self->parser = NULL;
	self->lexer = NULL;
	self->compiler = NULL;
	self->j4vm = NULL;

	if (RUN_VM_EXECUTABLE) {
		self->j4vm = create_jayfor_vm();
		run_vm_executable(self);
		return self;
	}

	return self;
}

void start_jayfor(jayfor *self) {
	if (RUN_VM_EXECUTABLE) return;

	// start actual useful shit here
	self->scanner = create_scanner();
	scan_file(self->scanner, self->filename);

	self->lexer = create_lexer(self->scanner->contents);

	while (self->lexer->running) {
		get_next_token(self->lexer);
	}

	// initialise parser after we tokenize
	self->parser = create_parser(self->lexer->token_stream);

	start_parsing_token_stream(self->parser);

	self->compiler = create_compiler();
	start_compiler(self->compiler, self->parser->parse_tree);
	
	if (EXECUTE_BYTECODE) {
		self->j4vm = create_jayfor_vm();
		start_jayfor_vm(self->j4vm, self->compiler->bytecode);
	}
	else {
		// output bytecode to file, overwrite existing
		FILE *output = fopen(OUTPUT_EXECUTABLE_NAME, "wb");
		if (!output) {
			perror("fopen: failed to create executable\n");
			exit(1);
		}

		int i;
		for (i = 0; i < self->compiler->current_instruction; i++) {
			// every 8 instructions, newline
			if (i % 8 == 0 && i != 0) {
				fprintf(output, "\n");
			}
			// print instruction as 4 digit number 
			fprintf(output, "%04X ", self->compiler->bytecode[i]);
		}
		fclose(output);
	}
}

void run_vm_executable(jayfor *self) {
	FILE *executable = fopen(VM_EXECUTABLE_NAME, "rb");
	int default_bytecode_size = 32;
	int *bytecode = malloc(sizeof(int) * default_bytecode_size);

	if (!executable) {
		perror("fopen: failed to open executable");
		exit(1);
	}

	int i = 0;
	int current_value = 0;
	while (fscanf(executable, "%04X", &current_value) != EOF) {
		bytecode[i++] = current_value;
		if (i >= default_bytecode_size) {
			default_bytecode_size *= 2;
			int *temp = realloc(bytecode, sizeof(*temp) * default_bytecode_size);
			if (!temp) {
				perror("realloc: failed to reallocate memory for bytecode");
				exit(1);
			}
			else {
				bytecode = temp;
			}
		}
	}

	start_jayfor_vm(self->j4vm, bytecode);
}

void destroy_jayfor(jayfor *self) {
	if (!RUN_VM_EXECUTABLE) {
		destroy_scanner(self->scanner);
		destroy_lexer(self->lexer);
		destroy_parser(self->parser);
		destroy_compiler(self->compiler);
	}
	destroy_jayfor_vm(self->j4vm);
	free(self);
}
