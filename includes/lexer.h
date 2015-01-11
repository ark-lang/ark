#ifndef LEXER_H
#define LEXER_H

/**
 * This is C code Linguist, come on...
 * 
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "util.h"
#include "vector.h"

#define WEIRD_CHARACTER_ASCII_THRESHOLD 128

/** types of token */
typedef enum {
	END_OF_FILE, IDENTIFIER, NUMBER,
	OPERATOR, SEPARATOR, ERRORNEOUS,
	STRING, CHARACTER, UNKNOWN, SPECIAL_CHAR
} token_type;

/**
 * Token properties:
 * type 		- the token type
 * content 		- the token content
 * line_number 	- the line number of the token
 * char_number 	- the number of the char of the token
 */
typedef struct {
	int type;
	char* content;
	int line_number;
	int char_number;
} token;

/**
 * Create an empty token
 * 
 * @return allocate memory for token
 */
token *create_token(lexer *lexer);

/**
 * Get the name of the given token
 * as a string
 * 
 * @param token token to find name of
 * @return the name of the given token
 */
const char* get_token_name(token *token);

/**
 * Deallocates memory for token
 * 
 * @param token token to free
 */
void destroy_token(token *token);

/** Lexer stuff */
typedef struct {
	char* input;			// input to lex
	int pos;				// position in the input
	int current_char;		// current character
	int line_number;		// current line number
	int char_number;		// current character at line
	int start_pos;			// keeps track of positions without comments
	bool running;			// if lexer is running 
	vector *token_stream;	// where the tokens are stored
} lexer;

/**
 * Retrieves the line that a token is on
 * @param  lexer              the lexer instance
 * @param  tok                the token to get the context of
 * @param  colour_error_token whether or not to colour the errored token
 * @return                    the context as a string
 */
const char* get_token_context(vector *stream, token *tok, bool colour_error_token);

/**
 * Retrieves the line that a token is on
 * @param  lexer 	the lexer instance
 * @param  line_num the number to get context of
 * @return       	the context as a string
 */
const char* get_line_number_context(vector *stream, int line_num);

/**
 * Create an instance of the Lexer
 * 
 * @param input the input to lex
 * @return instance of Lexer
 */
lexer *create_lexer(char* input);

/**
 * Simple substring, basically extracts the token from
 * the lexers input from [start .. start + length]
 * 
 * @param lexer instance of lexer
 * @param start start of the input
 * @param length of the input
 * @return string cut from buffer
 */
char* extract_token(lexer *lexer, int start, int length);

/**
 * Advance to the next character, consuming the
 * current one.
 * 
 * @param lexer instance of the lexer
 */
void consume_character(lexer *lexer);

/**
 * Skips layout characters, such as spaces,
 * and comments, which are denoted with the 
 * pound (#).
 * 
 * @param lexer the lexer instance
 */
void skip_layout_and_comments(lexer *lexer);

/**
 * Checks if current character is the given character
 * otherwise throws an error
 * 
 * @param lexer the lexer instance
 */
void expect_character(lexer *lexer, char c);

/**
 * Recognize an identifier
 * 
 * @param lexer the lexer instance
 */
void recognize_identifier_token(lexer *lexer);

/**
 * Recognize an Integer
 * 
 * @param lexer the lexer instance
 */
void recognize_number_token(lexer *lexer);

/**
 * Recognize a String
 * 
 * @param lexer the lexer instance
 */
void recognize_string_token(lexer *lexer);

/**
 * Recognize a Character
 * 
 * @param lexer the lexer instance
 */
void recognize_character_token(lexer *lexer);

/**
 * Recognizes the given operator and pushes it
 * @param lexer the lexer for access to the token stream
 */
void recognize_operator_token(lexer *lexer);

/**
 * Recognizes the end of line token
 * @param lexer the lexer for access to the token stream
 */
void recognize_end_of_line_token(lexer *lexer);

/**
 * Recognizes a separator token and pushes it
 * to the tree
 * @param lexer the lexer for access to the token stream
 */
void recognize_separator_token(lexer *lexer);

/**
 * Recognizes an errored token and pushes it to the
 * tree
 * @param lexer the lexer for access to the token stream
 */
void recognize_errorneous_token(lexer *lexer);

/**
 * Pushes a token to the token tree, also captures the 
 * token content so you don't have to.
 * 
 * @param lexer the lexer for access to the token tree
 * @param type  the type of token
 */
void push_token(lexer *lexer, int type);

/**
 * Pushes a token with content to the token tree
 * @param lexer   the lexer for access to the token tree
 * @param type    the type of token to push
 * @param content the content to push
 */
void push_token_c(lexer *lexer, int type, char *content);

/**
 * Peek ahead in the character stream by
 * the given amount
 * 
 * @lexer instance of lexer
 * @ahead amount to peek by
 * @return the char we peeked at
 */
char peek_ahead(lexer *lexer, int ahead);

/**
 * Process the next token in the token stream
 * 
 * @param lexer the lexer instance
 */
void get_next_token(lexer *lexer);

/**
 * Destroys the given lexer instance,
 * freeing any memory
 * 
 * @param lexer the lexer instance to destroy
 */
void destroy_lexer(lexer *lexer);

/**
 * @return if the character given is the end of input
 * @param ch the character to check
 */
static inline bool is_end_of_input(char ch) { 
	return ch == '\0'; 
}

/**
 * @return if the character given is a layout character
 * @param ch the character to check
 */
static inline bool is_layout(char ch) { 
	return !is_end_of_input(ch) && (ch) <= ' '; 
}

/**
 * @return if the character given is a comment closer 
 * @param ch the character to check
 */
static inline bool is_comment_closer(char ch) { 
	return ch == '\n'; 
}

/**
 * @return if the character given is an uppercase letter
 * @param ch the character to check
 */
static inline bool is_upper_letter(char ch) { 
	return 'A' <= ch && ch <= 'Z'; 
}

/**
 * @return if the character given is a lower case letter
 * @param ch the character to check
 */
static inline bool is_lower_letter(char ch) { 
	return 'a' <= ch && ch <= 'z'; 
}

/**
 * @return if the character given is a letter a-z, A-Z
 * @param ch the character to check
 */
static inline bool is_letter(char ch) { 
	return is_upper_letter(ch) || is_lower_letter(ch); 
}

/**
 * @return if the character given is a digit 0-9
 * @param ch the character to check
 */
static inline bool is_digit(char ch) { 
	return '0' <= ch && ch <= '9'; 
}

/**
 * @return if the character given is a letter or digit a-z, A-Z, 0-9
 * @param ch the character to check
 */
static inline bool is_letter_or_digit(char ch) { 
	return is_letter(ch) || is_digit(ch); 
}

/**
 * @return if the character given is an underscore
 * @param ch the character to check
 */
static inline bool is_underscore(char ch) { 
	return ch == '_'; 
}

/**
 * @return if the character given is a quote, denoting a string
 * @param ch the character to check
 */
static inline bool is_string(char ch) { 
	return ch == '"'; 
}

/**
 * @return if the character given is a single quote, denoting a character
 * @param ch the character to check
 */
static inline bool is_character(char ch) { 
	return ch == '\''; 
}

/**
 * @return if the character given is an operator
 * @param ch the character to check
 */
static inline bool is_operator(char ch) { 
	return (strchr("+-*/=><!~?:&%^\"'", ch) != 0); 
}

/**
 * @return if the character given is a separator
 * @param ch the character to check
 */
static inline bool is_separator(char ch) { 
	return (strchr(" ;,.`@(){}[] ", ch) != 0); 
}

/**
 * @return if the character is a special character like the British symbol or alike 
 * @param ch character to check
 */
static inline bool is_special_char(char ch) { 
	return (int) ch >= WEIRD_CHARACTER_ASCII_THRESHOLD; 
}

/**
 * @return if the character is end of line to track line number
 * @param ch character to check
 */
static inline bool is_end_of_line(char ch) { 
	return ch == '\n'; 
}

#endif // LEXER_H
