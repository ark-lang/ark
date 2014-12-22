#include "jayfor.h"

bool DEBUG_MODE = false;

static void parseArgument(char *arg) {
	char argument = arg[0];
	switch (argument) {
		case 'v':
			printf(KGRN "Jayfor Version: %s\n" KNRM, JAYFOR_VERSION);
			exit(1);
			break;
		case 'd':
			DEBUG_MODE = true;
			break;
		case 'h':
			printf(KYEL "JAYFOR Argument List\n");
			printf("\t-h,\t shows a help menu\n");
			printf("\t-v,\t shows current version\n");
			printf("\t-d,\t logs extra debug information\n");
			printf("\n" KNRM);
			exit(1);
			break;
		default:
			printf(KRED "error: unrecognized argument %c\n" KNRM, argument);
			exit(1);
			break;
	}
}

Jayfor *create_jayfor(int argc, char** argv) {
	// not enough args just throw an error
	if (argc <= 1) {
		printf(KRED "error: no input files\n" KNRM);
		exit(1);
	}

	char *filename = NULL;
	Jayfor *jayfor = malloc(sizeof(*jayfor));

	if (!jayfor) {
		perror("malloc: failed to allocate memory for JAYFOR");
		exit(1);
	}

	int i;
	// i = 1 to ignore first arg
	for (i = 1; i < argc; i++) {
		if (argv[i][0] == '-') {
			size_t len = strlen(argv[i]) - 1;
			char temp[len];
			memcpy(temp, &argv[i][len], 1);
			temp[len] = '\0';
			parseArgument(temp);
		}
		else if (strstr(argv[i], ".j4")) {
			filename = argv[i];
		}
		else {
			printf(KRED "error: argument not recognized: %s\n" KNRM, argv[i]);
		}
	}

	// just in case.
	jayfor->scanner = NULL;
	jayfor->lexer = NULL;
	jayfor->parser = NULL;

	// start actual useful shit here
	jayfor->scanner = create_scanner();
	scan_file(jayfor->scanner, filename);

	// pass the scanned file to the lexer to tokenize
	jayfor->lexer = create_lexer(jayfor->scanner->contents);
	jayfor->compiler = NULL;
	jayfor->j4vm = NULL;

	return jayfor;
}

void start_jayfor(Jayfor *jayfor) {
	while (jayfor->lexer->running) {
		get_next_token(jayfor->lexer);
	}

	// initialise parser after we tokenize
	jayfor->parser = create_parser(jayfor->lexer->token_stream);

	start_parsing_token_stream(jayfor->parser);

	jayfor->compiler = create_compiler();
	start_compiler(jayfor->compiler, jayfor->parser->parse_tree);

	jayfor->j4vm = create_jayfor_vm();
	start_jayfor_vm(jayfor->j4vm, jayfor->compiler->bytecode, jayfor->compiler->global_count + 1);
}

void destroy_jayfor(Jayfor *jayfor) {
	destroy_scanner(jayfor->scanner);
	destroy_lexer(jayfor->lexer);
	destroy_parser(jayfor->parser);
	destroy_compiler(jayfor->compiler);
	free(jayfor);
}
