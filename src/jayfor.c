#include "jayfor.h"

bool DEBUG_MODE = false;
bool EXECUTE_BYTECODE = false;
char *EXECUTABLE_FILENAME = "a.j4e";

typedef struct {
	char *argument;
	char *next_argument;
} argument;

static void parseArgument(argument *arg) {
	char argument = arg->argument[0];

	switch (argument) {
		case 'v':
			printf(KGRN "jayfor Version: %s\n" KNRM, jayfor_VERSION);
			exit(1);
			break;
		case 'd':
			DEBUG_MODE = true;
			break;
		case 'h':
			printf(KYEL "jayfor Argument List\n");
			printf("\t-h,\t\t shows a help menu\n");
			printf("\t-v,\t\t shows current version\n");
			printf("\t-d,\t\t logs extra debug information\n");
			printf("\t-r,\t\t will compile and execute code instead of creating an executable\n");
			printf("\t-o [file name],\t\t creates an executable with the given file name and extension\n");
			printf("\n" KNRM);
			exit(1);
			break;
		case 'o':
			EXECUTABLE_FILENAME = arg->next_argument;
			break;
		case 'r':
			EXECUTE_BYTECODE = true;
			break;
		default:
			printf(KRED "error: unrecognized argument %c\n" KNRM, argument);
			exit(1);
			break;
	}
}

jayfor *create_jayfor(int argc, char** argv) {
	// not enough args just throw an error
	if (argc <= 1) {
		printf(KRED "error: no input files\n" KNRM);
		exit(1);
	}

	char *filename = NULL;
	jayfor *self = malloc(sizeof(*self));

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
			filename = argv[i];
		}
		else {
			printf(KRED "error: argument not recognized: %s\n" KNRM, argv[i]);
		}
	}

	// just in case.
	self->scanner = NULL;
	self->lexer = NULL;
	self->parser = NULL;

	// start actual useful shit here
	self->scanner = create_scanner();
	scan_file(self->scanner, filename);

	// pass the scanned file to the lexer to tokenize
	self->lexer = create_lexer(self->scanner->contents);
	self->compiler = NULL;
	self->j4vm = NULL;

	return self;
}

void start_jayfor(jayfor *self) {
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
		start_jayfor_vm(self->j4vm, self->compiler->bytecode, self->compiler->global_count + 1);
	}
	else {
		// output bytecode to file, overwrite existing
		FILE *output = fopen(EXECUTABLE_FILENAME, "wb");
		if (!output) {
			perror("fopen: failed to create executable\n");
			exit(1);
		}

		int i;
		for (i = 0; i < self->compiler->current_instruction; i++) {
			if (i % 8 == 0 && i != 0) {
				fprintf(output, "\n");
			}
			fprintf(output, "%04d ", *(self->compiler->bytecode + i));
		}
		fclose(output);
	}
}

void destroy_jayfor(jayfor *self) {
	destroy_scanner(self->scanner);
	destroy_lexer(self->lexer);
	destroy_parser(self->parser);
	destroy_compiler(self->compiler);
	if (EXECUTE_BYTECODE) {
		destroy_jayfor_vm(self->j4vm);
	}
	free(self);
}
