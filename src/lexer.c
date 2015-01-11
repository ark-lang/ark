#include "lexer.h"

static const char* token_NAMES[] = {
	"END_OF_FILE", "IDENTIFIER", "NUMBER",
	"OPERATOR", "SEPARATOR", "ERRORNEOUS",
	"STRING", "CHARACTER", "UNKNOWN"
};

token *create_token() {
	token *token = safe_malloc(sizeof(*token));
	return token;
}

const char* get_token_name(token *tok) {
	return token_NAMES[tok->type];
}

void destroy_token(token *token) {
	if (token) {
		free(token);
		token = NULL;
	}
}

Lexer *create_lexer(char* input) {
	Lexer *lexer = safe_malloc(sizeof(*lexer));
	
	lexer->input = input;
	lexer->pos = 0;
	lexer->current_char = input[lexer->pos];
	lexer->token_stream = create_vector();
	lexer->running = true;
	lexer->line_number = 0;

	return lexer;
}

void consume_character(Lexer *lexer) {
	lexer->current_char = lexer->input[++lexer->pos];
}

char* extract_token(Lexer *lexer, int start, int length) {
	char* result = safe_malloc(sizeof(char) * (length + 1));
	strncpy(result, &lexer->input[start], length);
	result[length] = '\0';
	return result;
}

void skip_layout_and_comments(Lexer *lexer) {
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

void expect_character(Lexer *lexer, char c) {
	if (lexer->current_char == c) {
		consume_character(lexer);
	}
	else {
		printf("error: expected `%c` but found `%c`\n", c, lexer->current_char);
		exit(1);
	}
}

void recognize_identifier_token(Lexer *lexer) {
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
}

void recognize_number_token(Lexer *lexer) {
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
}

void recognize_string_token(Lexer *lexer) {
	expect_character(lexer, '"');

	// just consume everthing
	while (!is_string(lexer->current_char)) {
		consume_character(lexer);
	}

	expect_character(lexer, '"');
}

void recognize_character_token(Lexer *lexer) {
	expect_character(lexer, '\'');

	if (is_letter_or_digit(lexer->current_char)) {
		consume_character(lexer); // consume character		
	}
	else {
		printf("error: empty character constant\n");
		exit(1);
	}

	expect_character(lexer, '\'');
}

char peek_ahead(Lexer *lexer, int ahead) {
	return lexer->input[lexer->pos + ahead];
}

void get_next_token(Lexer *lexer) {
	int startPos;
	skip_layout_and_comments(lexer);
	startPos = lexer->pos;

	lexer->current_token = create_token();

	if (is_end_of_input(lexer->current_char)) {
		lexer->current_token->type = END_OF_FILE;
		lexer->current_token->content = "<END_OF_FILE>";

		lexer->running = false;	// stop lexing

		// push last item onto token stream
		push_back_item(lexer->token_stream, lexer->current_token);
		return;
	}
	if (is_letter(lexer->current_char) || is_digit(lexer->current_char) || lexer->current_char == '_') {
		lexer->current_token->type = IDENTIFIER;
		recognize_identifier_token(lexer);
	}
	else if (is_digit(lexer->current_char) || lexer->current_char == '.') {
		lexer->current_token->type = NUMBER;
		recognize_number_token(lexer);
	}
	else if (is_string(lexer->current_char)) {
		lexer->current_token->type = STRING;
		recognize_string_token(lexer);
	}
	else if (is_character(lexer->current_char)) {
		lexer->current_token->type = CHARACTER;
		recognize_character_token(lexer);
	}
	else if (is_operator(lexer->current_char)) {
		lexer->current_token->type = OPERATOR;
		consume_character(lexer);
	}
	else if (is_end_of_line(lexer->current_char)) {
		consume_character(lexer);
	}
	else if (is_separator(lexer->current_char)) {
		lexer->current_token->type = SEPARATOR;
		consume_character(lexer);
	}
	else {
		lexer->current_token->type = ERRORNEOUS;
		consume_character(lexer);
	}

	lexer->current_token->content = extract_token(lexer, startPos, lexer->pos - startPos);
	printf("%s\n", lexer->current_token->content);
	push_back_item(lexer->token_stream, lexer->current_token);
}

void destroy_lexer(Lexer *lexer) {
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
