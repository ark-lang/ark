#include "lexer.h"

// this is just for debugging
static const char* TOKEN_NAMES[] = {
	"END_OF_FILE", "IDENTIFIER", "NUMBER",
	"OPERATOR", "SEPARATOR", "ERRORNEOUS",
	"STRING", "CHARACTER", "UNKNOWN"
};

token *create_token() {
	token *tok = safe_malloc(sizeof(*tok));
	tok->type = UNKNOWN;
	tok->content = NULL;
	return tok;
}

const char* get_token_name(token *tok) {
	return TOKEN_NAMES[tok->type];
}

void destroy_token(token *token) {
	if (token) {
		free(token);
		token = NULL;
	}
}

lexer *create_lexer(char* input) {
	lexer *lexer = safe_malloc(sizeof(*lexer));
	lexer->input = input;
	lexer->pos = 0;
	lexer->current_char = input[lexer->pos];
	lexer->token_stream = create_vector();
	lexer->running = true;
	lexer->line_number = 0;
	lexer->char_number = 0;
	return lexer;
}

void consume_character(lexer *lexer) {
	if (lexer->current_char == '\n' || is_end_of_input(lexer->current_char)) {
		lexer->char_number = 0;
		lexer->line_number++;
		printf("new line %d\n", lexer->line_number);	
	}
	lexer->current_char = lexer->input[++lexer->pos];
	lexer->char_number++;
}

char* extract_token(lexer *lexer, int start, int length) {
	char* result = safe_malloc(sizeof(char) * (length + 1));
	strncpy(result, &lexer->input[start], length);
	result[length] = '\0';
	return result;
}

void skip_layout_and_comments(lexer *lexer) {
	while (is_layout(lexer->current_char)) {
		consume_character(lexer);
	}

	// consume a block comment and its contents
	if (lexer->current_char == '/' && peek_ahead(lexer, 1) == '*') {
		// consume new comment symbols
		consume_character(lexer);
		consume_character(lexer);

		while (true) {
			consume_character(lexer);
			if (lexer->current_char == '*' && peek_ahead(lexer, 1) == '/') {
				// consume the comment symbols
				consume_character(lexer);
				consume_character(lexer);

				// eat layout stuff like space etc
				while (is_layout(lexer->current_char)) {
					consume_character(lexer);
				}
				break;
			}
		}
	}

	// consume a single line comment
	while ((lexer->current_char == '/' && peek_ahead(lexer, 1) == '/')) {
		consume_character(lexer);	// eat the /
		consume_character(lexer);	// eat the /

		while (!is_comment_closer(lexer->current_char)) {
			if (is_end_of_input(lexer->current_char)) return;
			consume_character(lexer);
		}
		
		while (is_layout(lexer->current_char)) {
			consume_character(lexer);
		}
	}
}

void expect_character(lexer *lexer, char c) {
	if (lexer->current_char == c) {
		consume_character(lexer);
	}
	else {
		printf("error: expected `%c` but found `%c`\n", c, lexer->current_char);
		exit(1);
	}
}

void recognize_end_of_input_token(lexer *lexer) {
	push_token_c(lexer, END_OF_FILE, "<END_OF_FILE>");
}

void recognize_identifier_token(lexer *lexer) {
	consume_character(lexer);

	while (is_letter_or_digit(lexer->current_char)) {
		consume_character(lexer);
	}
	while (is_underscore(lexer->current_char) && is_letter_or_digit(peek_ahead(lexer, 1))) {
		consume_character(lexer);
		while (is_letter_or_digit(lexer->current_char)) {
			consume_character(lexer);
		}
	}

	push_token(lexer, IDENTIFIER);
}

void recognize_number_token(lexer *lexer) {
	consume_character(lexer);
	if (lexer->current_char == '.') {
		consume_character(lexer); // consume dot
		while (is_digit(lexer->current_char)) {
			consume_character(lexer);
		}
	}

	while (is_digit(lexer->current_char)) {
		if (peek_ahead(lexer, 1) == '.') {
			consume_character(lexer);
			while (is_digit(lexer->current_char)) {
				consume_character(lexer);
			}
		}
		consume_character(lexer);
	}

	push_token(lexer, NUMBER);
}

void recognize_string_token(lexer *lexer) {
	expect_character(lexer, '"');

	// just consume everthing
	while (!is_string(lexer->current_char)) {
		consume_character(lexer);
	}

	expect_character(lexer, '"');

	push_token(lexer, STRING);
}

void recognize_character_token(lexer *lexer) {
	expect_character(lexer, '\'');

	if (is_letter_or_digit(lexer->current_char)) {
		consume_character(lexer); // consume character		
	}
	else {
		printf("error: empty character constant\n");
		exit(1);
	}

	expect_character(lexer, '\'');

	push_token(lexer, CHARACTER);
}

void recognize_operator_token(lexer *lexer) {
	consume_character(lexer);
	push_token(lexer, OPERATOR);
}

void recognize_end_of_line_token(lexer *lexer) {
	consume_character(lexer);
}

void recognize_separator_token(lexer *lexer) {
	consume_character(lexer);
	push_token(lexer, SEPARATOR);
}

void recognize_errorneous_token(lexer *lexer) {
	consume_character(lexer);
	push_token(lexer, ERRORNEOUS);
}

void push_token(lexer *lexer, int type) {
	token *tok = create_token();
	tok->type = type;
	tok->content = extract_token(lexer, lexer->start_pos, lexer->pos - lexer->start_pos);
	push_back_item(lexer->token_stream, tok);
}

void push_token_c(lexer *lexer, int type, char *content) {
	token *tok = create_token();
	tok->type = type;
	tok->content = content;
	push_back_item(lexer->token_stream, tok);
}

char peek_ahead(lexer *lexer, int ahead) {
	return lexer->input[lexer->pos + ahead];
}

void get_next_token(lexer *lexer) {
	lexer->start_pos = 0;
	skip_layout_and_comments(lexer);
	lexer->start_pos = lexer->pos;

	if (is_end_of_input(lexer->current_char)) {
		recognize_end_of_input_token(lexer);
		lexer->running = false;	// stop lexing
		return;
	}
	if (is_letter(lexer->current_char) || is_digit(lexer->current_char) || lexer->current_char == '_') {
		// ident
		recognize_identifier_token(lexer);
	}
	else if (is_digit(lexer->current_char) || lexer->current_char == '.') {
		// number
		recognize_number_token(lexer);
	}
	else if (is_string(lexer->current_char)) {
		// string
		recognize_string_token(lexer);
	}
	else if (is_character(lexer->current_char)) {
		// character
		recognize_character_token(lexer);
	}
	else if (is_operator(lexer->current_char)) {
		// operator
		recognize_operator_token(lexer);
	}
	else if (is_end_of_line(lexer->current_char)) {
		recognize_end_of_line_token(lexer);
	}
	else if (is_separator(lexer->current_char)) {
		// separator
		recognize_separator_token(lexer);
	}
	else {
		// errorneous
		recognize_errorneous_token(lexer);
	}
}

void destroy_lexer(lexer *lexer) {
	if (lexer) {
		int i;
		for (i = 0; i < lexer->token_stream->size; i++) {
			token *tok = get_vector_item(lexer->token_stream, i);
			// eof isnt malloc'd for content
			if (tok->type != END_OF_FILE) {
				free(tok->content);
			}
			destroy_token(tok);
		}
		free(lexer);
	}
}
