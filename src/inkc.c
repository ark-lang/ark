#include "inkc.h"

bool DEBUG_MODE = false;
char *OUTPUT_EXECUTABLE_NAME = "a";

static void parse_argument(argument *arg) {
	char argument = arg->argument[0];

	switch (argument) {
		case 'v':
			printf("inkc version: %s\n", INKC_VERSION);
			exit(1);
			break;
		case 'd':
			DEBUG_MODE = true;
			break;
		case 'h':
			printf("Ink-Lang Argument List\n");
			printf("\t-h,\t shows a help menu\n");
			printf("\t-v,\t shows current version\n");
			printf("\t-d,\t logs extra debug information\n");
			printf("\n");
			exit(1);
			break;
		case 'o':
			if (!arg->next_argument) {
				error_message("error: missing filename after '-o'");
			}
			OUTPUT_EXECUTABLE_NAME = arg->next_argument;
			break;
		default:
			error_message("error: unrecognized command line option '-%s'\n", arg->argument);
			break;
	}
}

inkc *create_inkc(int argc, char** argv) {
	inkc *self = safe_malloc(sizeof(*self));
	self->filename = NULL;
	self->scanner = NULL;
	self->lexer = NULL;
	self->parser = NULL;
	self->compiler = NULL;
	self->pproc = NULL;
	self->semantic = NULL;

	// not enough args just throw an error
	if (argc <= 1) {
		error_message("no input files");
		return self;
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
			arg.next_argument = NULL;
			if (argv[i + 1] != NULL) {
				arg.next_argument = argv[i + 1];
			}

			// multiple arguments needed for -o, consume twice
			// todo make this cleaner for when we expand
			if (!strcmp(arg.argument, "o") || !strcmp(arg.argument, "e")) {
				i += 2;
			}

			// parse the argument
			parse_argument(&arg);
		}
		else if (strstr(argv[i], ".ink")) {
			self->filename = argv[i];
		}
		else {
			error_message("argument not recognized: %s\n", argv[i]);
		}
	}

	return self;
}

void start_inkc(inkc *self) {
	// filename is null, so we should exit
	// out of here
	if (self->filename == NULL) {
		return;
	}

	// start actual useful shit here
	self->scanner = create_scanner();
	scan_file(self->scanner, self->filename);

	// lex file
	self->lexer = create_lexer(self->scanner->contents);
	while (self->lexer->running) {
		get_next_token(self->lexer);
	}

	// initialise parser after we tokenize
	self->parser = create_parser(self->lexer->token_stream);
	start_parsing_token_stream(self->parser);

	self->semantic = create_semantic_analyser(self->parser->parse_tree);
	start_analysis(self->semantic);
}

void destroy_inkc(inkc *self) {
	destroy_scanner(self->scanner);
	destroy_lexer(self->lexer);
	destroy_parser(self->parser);
	destroy_semantic_analyser(self->semantic);
	free(self);
}
