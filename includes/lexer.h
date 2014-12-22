#ifndef LEXER_H
#define LEXER_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "util.h"
#include "vector.h"

#define WEIRD_CHARACTER_ASCII_THRESHOLD 128
#define uint unsigned int

/** Types of token */
typedef enum {
	END_OF_FILE, IDENTIFIER, NUMBER,
	OPERATOR, SEPARATOR, ERRORNEOUS,
	STRING, CHARACTER, UNKNOWN, SPECIAL_CHAR
} token_type;

/** Properties of a token or Lexeme */
typedef struct {
	int type;
	char* content;
} token;

/**
 * Create an empty token
 * 
 * @return allocate memory for token
 */
token *create_token();

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
	vector *token_stream;
	char* input;			// input to lex
	int pos;				// position in the input
	int char_index;			// current character
	token* current_token;	// current token
	int line_number;			// current line number
	bool running;			// if lexer is running 
} Lexer;

/**
 * Create an instance of the Lexer
 * 
 * @param input the input to lex
 * @return instance of Lexer
 */
Lexer *create_lexer(char* input);

/**
 * Simple substring implementation,
 * used to flush buffer. This is malloc'd memory, so free it!
 * 
 * @param lexer instance of lexer
 * @param start start of the input
 * @param length of the input
 * @return string cut from buffer
 */
char* flush_buffer(Lexer *lexer, int start, int length);

/**
 * Advance to the next character, consuming the
 * current one.
 * 
 * @param lexer instance of the lexer
 */
void consume_character(Lexer *lexer);

/**
 * Skips layout characters, such as spaces,
 * and comments, which are denoted with the 
 * pound (#).
 * 
 * @param lexer the lexer instance
 */
void skip_layout_and_comments(Lexer *lexer);

/**
 * Checks if current character is the given character
 * otherwise throws an error
 * 
 * @param lexer the lexer instance
 */
void expect_character(Lexer *lexer, char c);

/**
 * Recognize an identifier
 * 
 * @param lexer the lexer instance
 */
void recognize_identifier_token(Lexer *lexer);

/**
 * Recognize an Integer
 * 
 * @param lexer the lexer instance
 */
void recognize_number_token(Lexer *lexer);

/**
 * Recognize a String
 * 
 * @param lexer the lexer instance
 */
void recognize_string_token(Lexer *lexer);

/**
 * Recognize a Character
 * 
 * @param lexer the lexer instance
 */
void recognize_character_token(Lexer *lexer);

/**
 * Peek ahead in the character stream by
 * the given amount
 * 
 * @lexer instance of lexer
 * @ahead amount to peek by
 * @return the char we peeked at
 */
char peek_ahead(Lexer *lexer, int ahead);

/**
 * Process the next token in the token stream
 * 
 * @param lexer the lexer instance
 */
void get_next_token(Lexer *lexer);

/**
 * Destroys the given lexer instance,
 * freeing any memory
 * 
 * @param lexer the lexer instance to destroy
 */
void destroy_lexer(Lexer *lexer);

/**
 * @return if the character given is the end of input
 * @param ch the character to check
 */
static inline bool is_end_of_input(char ch) 		{ return ch == '\0'; }

/**
 * @return if the character given is a comment opener (#)
 * @param ch the character to check
 */
static inline bool is_comment_opener(char ch) 	{ return ch == '#'; }

/**
 * @return if the character given is a layout character
 * @param ch the character to check
 */
static inline bool is_layout(char ch) 			{ return !is_end_of_input(ch) && (ch) <= ' '; }

/**
 * @return if the character given is a comment closer 
 * @param ch the character to check
 */
static inline bool is_comment_closer(char ch) 	{ return ch == '\n'; }

/**
 * @return if the character given is an uppercase letter
 * @param ch the character to check
 */
static inline bool is_upper_letter(char ch) 		{ return 'A' <= ch && ch <= 'Z'; }

/**
 * @return if the character given is a lower case letter
 * @param ch the character to check
 */
static inline bool is_lower_letter(char ch) 		{ return 'a' <= ch && ch <= 'z'; }

/**
 * @return if the character given is a letter a-z, A-Z
 * @param ch the character to check
 */
static inline bool is_letter(char ch) 			{ return is_upper_letter(ch) || is_lower_letter(ch); }

/**
 * @return if the character given is a digit 0-9
 * @param ch the character to check
 */
static inline bool is_digit(char ch) 			{ return '0' <= ch && ch <= '9'; }

/**
 * @return if the character given is a letter or digit a-z, A-Z, 0-9
 * @param ch the character to check
 */
static inline bool is_letter_or_digit(char ch) 	{ return is_letter(ch) || is_digit(ch); }

/**
 * @return if the character given is an underscore
 * @param ch the character to check
 */
static inline bool is_underscore(char ch) 		{ return ch == '_'; }

/**
 * @return if the character given is a quote, denoting a string
 * @param ch the character to check
 */
static inline bool is_string(char ch) 			{ return ch == '"'; }

/**
 * @return if the character given is a single quote, denoting a character
 * @param ch the character to check
 */
static inline bool is_character(char ch) 		{ return ch == '\''; }

/**
 * @return if the character given is an operator
 * @param ch the character to check
 */
static inline bool is_operator(char ch) 			{ return (strchr("+-*/=><!~?:&%^\"'", ch) != 0); }

/**
 * @return if the character given is a separator
 * @param ch the character to check
 */
static inline bool is_separator(char ch) 		{ return (strchr(" ;,.`@(){}[] ", ch) != 0); }

/**
 * @return if the character is a special character like the British symbol or alike 
 * @param ch character to check
 */
static inline bool is_special_char(char ch)		{ return (int) ch >= WEIRD_CHARACTER_ASCII_THRESHOLD; }

/**
 * @return if the character is end of line to track line number
 * @param ch character to check
 */
static inline bool is_end_of_line(char ch)			{ return ch == '\n'; }

#endif // LEXER_H
