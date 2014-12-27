#include "parser.h"

/** List of token names */
static const char* token_NAMES[] = {
	"END_OF_FILE", "IDENTIFIER", "NUMBER",
	"OPERATOR", "SEPARATOR", "ERRORNEOUS",
	"STRING", "CHARACTER", "UNKNOWN"
};

/** List of data types */
static const char* DATA_TYPES[] = {
	"int", "str", "double", "float", "bool",
	"void", "char", "tup"
};

/** UTILITY FOR AST NODES */

infinite_loop_ast_node *create_infinite_loop_ast_node() {
	infinite_loop_ast_node *iln = malloc(sizeof(*iln));
	if (!iln) {
		perror("malloc: failed to allocate memory for Infinite Loop AST Node");
		exit(1);
	}
	iln->body = NULL;
	return iln;
}

break_ast_node *create_break_ast_node() {
	break_ast_node *bn = malloc(sizeof(*bn));
	if (!bn) {
		perror("malloc: failed to allocate memory for Break AST NODE");
		exit(1);
	}
	return bn;
}

variable_reassignment_ast_node *create_variable_reassign_ast_node() {
	variable_reassignment_ast_node *vrn = malloc(sizeof(*vrn));
	if (!vrn) {
		perror("malloc: failed to allocate memory for Variable Reassign ast_node");
		exit(1);
	}
	vrn->name = NULL;
	vrn->expr = NULL;
	return vrn;
}

statement_ast_node *create_statement_ast_node() {
	statement_ast_node *sn = malloc(sizeof(*sn));
	if (!sn) {
		perror("malloc: failed to allocate memory for Statement ast_node");
		exit(1);
	}
	sn->data = NULL;
	sn->type = 0;
	return sn;
}

function_return_ast_node *create_function_return_ast_node() {
	function_return_ast_node *frn = malloc(sizeof(*frn));
	if (!frn) {
		perror("malloc: failed to allocate memory for Function Return ast_node");
		exit(1);
	}
	frn->returnVals = NULL;
	return frn;
}

expression_ast_node *create_expression_ast_node() {
	expression_ast_node *expr = malloc(sizeof(*expr));
	if (!expr) {
		perror("malloc: failed to allocate memory for expression_ast_node");
		exit(1);
	}
	expr->value = NULL;
	expr->lhand = NULL;
	expr->rhand = NULL;
	return expr;
}

bool_expression_ast_node *create_boolean_expression_ast_node() {
    bool_expression_ast_node *boolExpr = malloc(sizeof(*boolExpr));
    if (!boolExpr) {
        perror("malloc: failed to allocate memory for bool_expression_ast_node");
        exit(1);
    }
    boolExpr->expr = NULL;
    boolExpr->lhand = NULL;
    boolExpr->rhand = NULL;
    return boolExpr;
}

variable_define_ast_node *create_variable_define_ast_node() {
	variable_define_ast_node *vdn = malloc(sizeof(*vdn));
	if (!vdn) {
		perror("malloc: failed to allocate memory for variable_define_ast_node");
		exit(1);
	}
	vdn->name = NULL;
	return vdn;
}

variable_declare_ast_node *create_variable_declare_ast_node() {
	variable_declare_ast_node *vdn = malloc(sizeof(*vdn));
	if (!vdn) {
		perror("malloc: failed to allocate memory for variable_declare_ast_node");
		exit(1);
	}
	vdn->vdn = NULL;
	vdn->expr = NULL;
	return vdn;
}

function_argument_ast_node *create_function_argument_ast_node() {
	function_argument_ast_node *fan = malloc(sizeof(*fan));
	if (!fan) {
		perror("malloc: failed to allocate memory for function_argument_ast_node");
		exit(1);
	}
	fan->name = NULL;
	fan->value = NULL;
	return fan;
}

function_callee_ast_node *create_function_callee_ast_node() {
	function_callee_ast_node *fcn = malloc(sizeof(*fcn));
	if (!fcn) {
		perror("malloc: failed to allocate memory for function_callee_ast_node");
		exit(1);
	}
	fcn->callee = NULL;
	fcn->args = NULL;
	return fcn;
}

block_ast_node *create_block_ast_node() {
	block_ast_node *bn = malloc(sizeof(*bn));
	if (!bn) {
		perror("malloc: failed to allocate memory for block_ast_node");
		exit(1);
	}
	bn->statements = NULL;
	return bn;
}

function_prototype_ast_node *create_function_prototype_ast_node() {
	function_prototype_ast_node *fpn = malloc(sizeof(*fpn));
	if (!fpn) {
		perror("malloc: failed to allocate memory for function_prototype_ast_node");
		exit(1);
	}
	fpn->args = NULL;
	fpn->name = NULL;
	return fpn;
}

enumeration_ast_node *create_enumeration_ast_node() {
	enumeration_ast_node *en = malloc(sizeof(*en));
	if (!en) {
		perror("malloc: failed to create enum ast node");
		exit(1);
	}
	en->name = NULL;
	en->enum_items = create_vector();
	return en;
}

enum_item *create_enum_item(char *name, int value) {
	enum_item *ei = malloc(sizeof(*ei));
	if (!ei) {
		perror("malloc: failed to allocate memory for enum item");
		exit(1);
	}
	ei->name = name;
	ei->value = value;
	return ei;
}

function_ast_node *create_function_ast_node() {
	function_ast_node *fn = malloc(sizeof(*fn));
	if (!fn) {
		perror("malloc: failed to allocate memory for function_ast_node");
		exit(1);
	}
	fn->fpn = NULL;
	fn->body = NULL;
	return fn;
}

for_loop_ast_node *create_for_loop_ast_node() {
	for_loop_ast_node *fln = malloc(sizeof(*fln));
	if (!fln) {
		perror("malloc: failed to allocate memory for for_loop_ast_node");
		exit(1);
	}
	return fln;
}

structure_ast_node *create_structure_ast_node() {
	structure_ast_node *sn = malloc(sizeof(*sn));
	if (!sn) {
		perror("malloc: failed to allocate memory for structure_ast_node");
		exit(1);
	}
	sn->statements = create_vector();
	return sn;
}

void destroy_variable_reassign_ast_node(variable_reassignment_ast_node *vrn) {
	if (!vrn) {
		if (!vrn->expr) {
			destroy_expression_ast_node(vrn->expr);
		}
		free(vrn);
		vrn = NULL;
	}
}

void destroy_for_loop_ast_node(for_loop_ast_node *fln) {
	if (!fln) {
		free(fln);
		fln = NULL;
	}
}

void destroy_break_ast_node(break_ast_node *bn) {
	if (!bn) {
		free(bn);
		bn = NULL;
	}
}

void destroy_statement_ast_node(statement_ast_node *sn) {
	if (!sn) {
		if (!sn->data) {
			switch (sn->type) {
				case VARIABLE_DEF_AST_NODE:
					destroy_variable_define_ast_node(sn->data);
					break;
				case VARIABLE_DEC_AST_NODE:
					destroy_variable_declare_ast_node(sn->data);
					break;
				case FUNCTION_CALLEE_AST_NODE:
					destroy_function_callee_ast_node(sn->data);
					break;
				case FUNCTION_RET_AST_NODE:
					destroy_function_ast_node(sn->data);
					break;
				case VARIABLE_REASSIGN_AST_NODE:
					destroy_variable_reassign_ast_node(sn->data);
					break;
				case FOR_LOOP_AST_NODE:
					destroy_for_loop_ast_node(sn->data);
					break;
				case INFINITE_LOOP_AST_NODE:
					destroy_infinite_loop_ast_node(sn->data);
					break;
				case BREAK_AST_NODE:
					destroy_break_ast_node(sn->data);
					break;
				case ENUM_AST_NODE:
					destroy_enumeration_ast_node(sn->data);
					break;
				default: break;
			}
		}
		free(sn);
		sn = NULL;
	}
}

void destroy_function_return_ast_node(function_return_ast_node *frn) {
	if (!frn) {
		if (!frn->returnVals) {
			int i;
			for (i = 0; i < frn->returnVals->size; i++) {
				expression_ast_node *temp = get_vector_item(frn->returnVals, i);
				if (!temp) {
					destroy_expression_ast_node(temp);
				}
			}
			destroy_vector(frn->returnVals);
		}
		free(frn);
		frn = NULL;
	}
}

void destroy_expression_ast_node(expression_ast_node *expr) {
	if (!expr) {
		if (!expr->lhand) {
			destroy_expression_ast_node(expr->lhand);
		}
		if (!expr->rhand) {
			destroy_expression_ast_node(expr->rhand);
		}
		free(expr);
		expr = NULL;
	}
}

void destroy_variable_define_ast_node(variable_define_ast_node *vdn) {
	if (!vdn) {
		free(vdn);
		vdn = NULL;
	}
}

void destroy_variable_declare_ast_node(variable_declare_ast_node *vdn) {
	if (!vdn) {
		if (!vdn->vdn) {
			destroy_variable_define_ast_node(vdn->vdn);
		}
		if (!vdn->expr) {
			destroy_expression_ast_node(vdn->expr);
		}
		free(vdn);
		vdn = NULL;
	}
}

void destroy_function_argument_ast_node(function_argument_ast_node *fan) {
	if (!fan) {
		if (!fan->value) {
			destroy_expression_ast_node(fan->value);
		}
		free(fan);
		fan = NULL;
	}
}

void destroy_block_ast_node(block_ast_node *bn) {
	if (!bn) {
		if (!bn->statements) {
			destroy_vector(bn->statements);
		}
		free(bn);
		bn = NULL;
	}
}

void destroy_infinite_loop_ast_node(infinite_loop_ast_node *iln) {
	if (!iln) {
		if (!iln->body) {
			destroy_block_ast_node(iln->body);
		}
		free(iln);
		iln = NULL;
	}
}

void destroy_function_prototype_ast_node(function_prototype_ast_node *fpn) {
	if (!fpn) {
		if (!fpn->args) {
			int i;
			for (i = 0; i < fpn->args->size; i++) {
				statement_ast_node *sn = get_vector_item(fpn->args, i);
				if (!sn) {
					destroy_statement_ast_node(sn);
				}
			}
			destroy_vector(fpn->args);
		}
		free(fpn);
		fpn = NULL;
	}
}

void destroy_function_ast_node(function_ast_node *fn) {
	if (!fn) {
		if (!fn->fpn) {
			destroy_function_prototype_ast_node(fn->fpn);
		}
		if (!fn->body) {
			destroy_block_ast_node(fn->body);
		}
		if (fn->ret) {
			destroy_vector(fn->ret);
		}
		free(fn);
		fn = NULL;
	}
}

void destroy_function_callee_ast_node(function_callee_ast_node *fcn) {
	if (!fcn) {
		if (!fcn->args) {
			destroy_vector(fcn->args);
		}
		free(fcn);
		fcn = NULL;
	}
}

void destroybool_expression_ast_node(bool_expression_ast_node *ben) {
    if (!ben) {
        if(!ben->lhand) {
            destroybool_expression_ast_node(ben->lhand);
        }
        if(!ben->rhand) {
            destroybool_expression_ast_node(ben->rhand);
        }
        free(ben);
        ben = NULL;
    }
}

void destroy_structure_ast_node(structure_ast_node *sn) {
	if (!sn) {
		if (!sn->statements) {
			destroy_vector(sn->statements);
		}
		free(sn);
		sn = NULL;
	}
}

void destroy_enumeration_ast_node(enumeration_ast_node *en) {
	if (!en) {
		if (!en->enum_items) {
			int i;
			for (i = 0; i < en->enum_items->size; i++) {
				destroy_enum_item(get_vector_item(en->enum_items, i));
			}
			destroy_vector(en->enum_items);
		}
		free(en);
		en = NULL;
	}
}

void destroy_enum_item(enum_item *ei) {
	if (!ei) {
		free(ei);
		ei = NULL;
	}
}

/** END ast_node FUNCTIONS */

parser *create_parser(vector *token_stream) {
	parser *parser = malloc(sizeof(*parser));
	if (!parser) {
		perror("malloc: failed to allocate memory for parser");
		exit(1);
	}
	parser->token_stream = token_stream;
	parser->parse_tree = create_vector();
	parser->token_index = 0;
	parser->parsing = true;
	return parser;
}

token *consume_token(parser *parser) {
	// return the token we are consuming, then increment token index
	return get_vector_item(parser->token_stream, parser->token_index++);
}

token *peek_at_token_stream(parser *parser, int ahead) {
	return get_vector_item(parser->token_stream, parser->token_index + ahead);
}

token *expect_token_type(parser *parser, token_type type) {
	token *tok = peek_at_token_stream(parser, 1);
	if (tok->type == type) {
		return consume_token(parser);
	}
	else {
		error_message("expected %s but found `%s`\n", token_NAMES[type], tok->content);
		return NULL;
	}
}

token *expect_token_content(parser *parser, char *content) {
	token *tok = peek_at_token_stream(parser, 1);
	if (!strcmp(tok->content, content)) {
		return consume_token(parser);
	}
	else {
		error_message("expected %s but found `%s`\n", tok->content, content);
		return NULL;
	}
}

token *expect_token_type_and_content(parser *parser, token_type type, char *content) {
	token *tok = peek_at_token_stream(parser, 1);
	if (tok->type == type && !strcmp(tok->content, content)) {
		return consume_token(parser);
	}
	else {
		error_message("expected %s but found `%s`\n", token_NAMES[type], tok->content);
		return NULL;
	}
}

token *match_token_type(parser *parser, token_type type) {
	token *tok = peek_at_token_stream(parser, 0);
	if (tok->type == type) {
		return consume_token(parser);
	}
	else {
		error_message("expected %s but found `%s`\n", token_NAMES[type], tok->content);
		return NULL;
	}
}

token *match_token_content(parser *parser, char *content) {
	token *tok = peek_at_token_stream(parser, 0);
	if (!strcmp(tok->content, content)) {
		return consume_token(parser);
	}
	else {
		error_message("expected %s but found `%s`\n", tok->content, content);
		return NULL;
	}
}

token *match_token_type_and_content(parser *parser, token_type type, char *content) {
	token *tok = peek_at_token_stream(parser, 0);
	if (tok->type == type && !strcmp(tok->content, content)) {
		return consume_token(parser);
	}
	else {
		error_message("expected %s but found `%s`\n", token_NAMES[type], tok->content);
		return NULL;
	}
}

bool check_token_type(parser *parser, token_type type, int ahead) {
	token *tok = peek_at_token_stream(parser, ahead);
	return tok->type == type;
}

bool check_token_content(parser *parser, char* content, int ahead) {
	token *tok = peek_at_token_stream(parser, ahead);
	return !strcmp(tok->content, content);
}

bool check_token_type_and_content(parser *parser, token_type type, char* content, int ahead) {
	return check_token_type(parser, type, ahead) && check_token_content(parser, content, ahead);
}

char parse_operand(parser *parser) {
	token *tok = peek_at_token_stream(parser, 0);
	char tokChar = tok->content[0];

	switch (tokChar) {
		case '+': consume_token(parser); return tokChar;
		case '-': consume_token(parser); return tokChar;
		case '*': consume_token(parser); return tokChar;
		case '/': consume_token(parser); return tokChar;
		case '%': consume_token(parser); return tokChar;
		case '>': consume_token(parser); return tokChar;
		case '<': consume_token(parser); return tokChar;
		case '^': consume_token(parser); return tokChar;
	}

	error_message("error: invalid operator ('%c') specified\n", tok->content[0]);
	return '\0';
}

enumeration_ast_node *parse_enumeration_ast_node(parser *parser) {
	enumeration_ast_node *en = create_enumeration_ast_node();
	
	match_token_type_and_content(parser, IDENTIFIER, ENUM_KEYWORD); // ENUM

	if (check_token_type(parser, IDENTIFIER, 0)) {
		en->name = consume_token(parser);	
		
		if (check_token_type_and_content(parser, SEPARATOR, "{", 0)) {
			consume_token(parser);

			do {
				if (check_token_type_and_content(parser, SEPARATOR, "}", 0)) {
					consume_token(parser);
					break;
				}

				
				// ENUM_ITEM = 0
				if (check_token_type(parser, IDENTIFIER, 0)) {
					token *enum_item_name = consume_token(parser);

					if (check_token_type_and_content(parser, OPERATOR, "=", 0)) {
						consume_token(parser);

						if (check_token_type(parser, NUMBER, 0)) {
							token *enum_item_value = consume_token(parser);

							int enum_item_value_as_int = atoi(enum_item_value->content);

							if (en->enum_items->size >= 1) {
								enum_item *prev_item = get_vector_item(en->enum_items, en->enum_items->size - 1);
								int prev_item_value = prev_item->value;
								char *prev_item_name = prev_item->name;

								// validate names are not duplicate
								if (!strcmp(prev_item_name, enum_item_name->content)) {
									error_message("error: duplicate enum items: \"%s\"", enum_item_name->content, prev_item_name);
								}
	
								// validate values are not duplicate
								if (prev_item_value == enum_item_value_as_int) {
									error_message("error: enum values cannot be the same: `%s = %d`", enum_item_name->content, enum_item_value_as_int);
								}
							}

							enum_item *item = create_enum_item(enum_item_name->content, enum_item_value_as_int);
							push_back_item(en->enum_items, item);

							parse_optional_semi_colon(parser);
							printf("pushed back %d as %s\n", enum_item_value_as_int, enum_item_name->content);
						}
						else {
							token *errorneous_token = consume_token(parser);
							error_message("error enum item expecting valid integer constant, found %s\n", errorneous_token->content);
						}
					}
					// ENUM_ITEM
					else {
						int enum_item_value_as_int = 0;
						if (en->enum_items->size >= 1) {
							enum_item *prev_item = get_vector_item(en->enum_items, en->enum_items->size - 1);
							enum_item_value_as_int = prev_item->value + 1;
							char *prev_item_name = prev_item->name;

							// validate name
							if (!strcmp(prev_item_name, enum_item_name->content)) {
								error_message("error: duplicate enum items: \"%s\"", enum_item_name->content, prev_item_name);
							}
						} 

						enum_item *item = create_enum_item(enum_item_name->content, enum_item_value_as_int);
						push_back_item(en->enum_items, item);

						printf("pushed back %d as %s\n", enum_item_value_as_int, enum_item_name->content);

						parse_optional_semi_colon(parser);
					}
				}

				// empty enum?
			}
			while (true);
		}
	}

	return en;
}

/*
 * struct struct_name {
 *     int x = 5;
 *     object y;
 *     struct another {
 *     }
 * }
 */
structure_ast_node *parse_structure_ast_node(parser *parser) {
	match_token_type_and_content(parser, IDENTIFIER, STRUCT_KEYWORD);
	token *struct_name = match_token_type(parser, IDENTIFIER);

	structure_ast_node *sn = create_structure_ast_node();
	sn->struct_name = struct_name->content;

	if (check_token_type_and_content(parser, SEPARATOR, "{", 0)) {
		consume_token(parser);

		do {
			if (check_token_type_and_content(parser, SEPARATOR, "}", 0)) {
				consume_token(parser);
				parse_optional_semi_colon(parser);
			}

			push_back_item(sn->statements, parse_variable_ast_node(parser, false));
		}
		while (true);
	}

	return sn;
}

statement_ast_node *parse_for_loop_ast_node(parser *parser) {
	/**
	 * for int x:(0, 10, 2) {
	 * 
	 * }
	 */

	// for token
	match_token_type_and_content(parser, IDENTIFIER, FOR_LOOP_KEYWORD);		// FOR
	
	token *type_tok = match_token_type(parser, IDENTIFIER);					// DATA_TYPE
	data_type type_raw = match_token_type_to_data_type(parser, type_tok);
	
	token *indexName = match_token_type(parser, IDENTIFIER);					// INDEX_NAME

	match_token_type_and_content(parser, OPERATOR, ":");						// PARAMS

	for_loop_ast_node *fln = create_for_loop_ast_node();
	fln->type = type_raw;
	fln->indexName = indexName;
	fln->params = create_vector();

	if (check_token_type_and_content(parser, SEPARATOR, "(", 0)) {
		consume_token(parser);

		int paramCount = 0;

		do {
			if (paramCount > 3) {
				error_message("error: for loop has one too many arguments %d\n", paramCount);
			}
			if (check_token_type_and_content(parser, SEPARATOR, ")", 0)) {
				if (paramCount < 2) {
					error_message("error: for loop expects a maximum of 3 arguments, you have %d\n", paramCount);
				}
				consume_token(parser);
				break;
			}

			if (check_token_type(parser, IDENTIFIER, 0)) {
				push_back_item(fln->params, consume_token(parser));
				if (check_token_type_and_content(parser, SEPARATOR, ",", 0)) {
					if (check_token_type_and_content(parser, SEPARATOR, ")", 1)) {
						error_message("error: trailing comma in for loop declaration!\n");
					}
					consume_token(parser);
				}
			}
			else if (check_token_type(parser, NUMBER, 0)) {
				push_back_item(fln->params, consume_token(parser));	
				if (check_token_type_and_content(parser, SEPARATOR, ",", 0)) {
					if (check_token_type_and_content(parser, SEPARATOR, ")", 1)) {
						error_message("error: trailing comma in for loop declaration!\n");
					}
					consume_token(parser);
				}
			}
			// it's an expression probably
			else if (check_token_type_and_content(parser, SEPARATOR, "(", 0)) {
				expression_ast_node *expr = parse_expression_ast_node(parser);
				push_back_item(fln->params, expr);
				if (check_token_type_and_content(parser, SEPARATOR, ",", 0)) {
					if (check_token_type_and_content(parser, SEPARATOR, ")", 1)) {
						error_message("error: trailing comma in for loop declaration!\n");
					}
					consume_token(parser);
				}
			}
			else {
				print_current_token(parser);
				error_message("error: expected a number or variable in for loop parameters, found:\n");
				return NULL;
			}

			paramCount++;
		}
		while (true);	
	
		fln->body = parse_block_ast_node(parser);

		statement_ast_node *sn = create_statement_ast_node();
		sn->type = FOR_LOOP_AST_NODE;
		sn->data = fln;
		return sn;
	}

	error_message("failed to parse for loop\n");
	return NULL;
}

expression_ast_node *parse_expression_ast_node(parser *parser) {
	expression_ast_node *expr = create_expression_ast_node(); // the final expression

	// number literal
	if (check_token_type(parser, NUMBER, 0)) {
		expr->type = EXPR_NUMBER;
		expr->value = consume_token(parser);
		return expr;
	}
	// string literal
	if (check_token_type(parser, STRING, 0)) {
		expr->type = EXPR_STRING;
		expr->value = consume_token(parser);
		return expr;
	}
	// character
	if (check_token_type(parser, CHARACTER, 0)) {
		expr->type = EXPR_CHARACTER;
		expr->value = consume_token(parser);
		return expr;
	}
	if (check_token_type(parser, IDENTIFIER, 0)) {
		expr->type = EXPR_VARIABLE;
		expr->value = consume_token(parser);
		return expr;
	}
	if (check_token_type_and_content(parser, SEPARATOR, "(", 0)) {
		consume_token(parser);
		expr->type = EXPR_PARENTHESIS;
		expr->lhand = parse_expression_ast_node(parser);
		expr->operand = parse_operand(parser);
		expr->rhand = parse_expression_ast_node(parser);
		if (check_token_type_and_content(parser, SEPARATOR, ")", 0)) {
			consume_token(parser);
			return expr;
		}
		error_message("error: missing closing parenthesis on expression\n");
		return NULL;
	}
    if(check_token_type_and_content(parser, OPERATOR, "!", 0)) {
        consume_token(parser);
        expr->type = EXPR_LOGICAL_OPERATOR;
    }

	print_current_token(parser);
	error_message("error: failed to parse expression, only character, string and numbers are supported\n");
	return NULL;
}

void print_current_token(parser *parser) {
	token *tok = peek_at_token_stream(parser, 0);
	error_message("current token is type: %s, value: %s\n", token_NAMES[tok->type], tok->content);
}

void *parse_variable_ast_node(parser *parser, bool global) {
	// TYPE NAME = 5;
	// TYPE NAME;

	bool is_constant = false;

	if (check_token_type_and_content(parser, IDENTIFIER, CONSTANT_KEYWORD, 0)) {
		consume_token(parser);
		is_constant = true;
	}

	// consume the int data type
	token *variabledata_type = match_token_type(parser, IDENTIFIER);

	// convert the data type for enum
	data_type data_typeRaw = match_token_type_to_data_type(parser, variabledata_type);

	// name of the variable
	token *variableNametoken = match_token_type(parser, IDENTIFIER);

	if (check_token_type_and_content(parser, OPERATOR, "=", 0)) {
		// consume the equals sign
		consume_token(parser);

		// create variable define ast_node
		variable_define_ast_node *def = create_variable_define_ast_node();
		def->is_constant = is_constant;
		def->type = data_typeRaw;
		def->name = variableNametoken;
		def->is_global = global;

		// parses the expression we're assigning to
		expression_ast_node *expr = parse_expression_ast_node(parser);

		// create the variable declare ast_node
		variable_declare_ast_node *dec = create_variable_declare_ast_node();
		dec->vdn = def;
		dec->expr = expr;

		parse_optional_semi_colon(parser);

		if (global) {
			prepare_ast_node(parser, dec, VARIABLE_DEC_AST_NODE);
			return dec;
		}

		// not global, pop it as a statement node
		statement_ast_node *sn = create_statement_ast_node();
		sn->data = dec;
		sn->type = VARIABLE_DEC_AST_NODE;
		return sn;
	}
	else {
		parse_optional_semi_colon(parser);

		// create variable define ast_node
		variable_define_ast_node *def = create_variable_define_ast_node();
		def->is_constant = is_constant;
		def->type = data_typeRaw;
		def->name = variableNametoken;
		def->is_global = global;
		
		if (global) {
			prepare_ast_node(parser, def, VARIABLE_DEF_AST_NODE);
			return def;
		}

		// not global, pop it as a statement node
		statement_ast_node *sn = create_statement_ast_node();
		sn->data = def;
		sn->type = VARIABLE_DEF_AST_NODE;
		return sn;
	}
}

block_ast_node *parse_block_ast_node(parser *parser) {
	block_ast_node *block = create_block_ast_node();
	block->statements = create_vector();

	match_token_type_and_content(parser, SEPARATOR, "{");
	
	do {
		// check if block is empty before we try parse some statements
		if (check_token_type_and_content(parser, SEPARATOR, "}", 0)) {
			consume_token(parser);
			break;
		}

		push_back_item(block->statements, parse_statement_ast_node(parser));
	}
	while (true);

	return block;
}

statement_ast_node *parse_infinite_loop_ast_node(parser *parser) {
	match_token_type(parser, IDENTIFIER);
	
	block_ast_node *body = parse_block_ast_node(parser);	
	
	infinite_loop_ast_node *iln = create_infinite_loop_ast_node();
	iln->body = body;

	statement_ast_node *sn = create_statement_ast_node();
	sn->data = iln;
	sn->type = INFINITE_LOOP_AST_NODE;
	
	return sn;
}

function_ast_node *parse_function_ast_node(parser *parser) {
	match_token_type(parser, IDENTIFIER);	// consume the fn keyword

	token *functionName = match_token_type(parser, IDENTIFIER); // name of function
	vector *args = create_vector(); // null for now till I add arg parsing

	// Create function signature
	function_prototype_ast_node *fpn = create_function_prototype_ast_node();
	fpn->args = args;
	fpn->name = functionName;

	// parameter list
	if (check_token_type_and_content(parser, SEPARATOR, "(", 0)) {
		consume_token(parser);

		do {
			// NO ARGUMENTS PROVIDED TO FUNCTION
			if (check_token_type_and_content(parser, SEPARATOR, ")", 0)) {
				consume_token(parser);
				break;
			}

			token *argdata_type = match_token_type(parser, IDENTIFIER);
			data_type argRawdata_type = match_token_type_to_data_type(parser, argdata_type);
			token *argName = match_token_type(parser, IDENTIFIER);

			function_argument_ast_node *arg = create_function_argument_ast_node();
			arg->type = argRawdata_type;
			arg->name = argName;
			arg->value = NULL;

			if (check_token_type_and_content(parser, OPERATOR, "=", 0)) {
				consume_token(parser);

				// default expression
				expression_ast_node *expr = parse_expression_ast_node(parser);
				arg->value = expr;
				push_back_item(args, arg);

				if (check_token_type_and_content(parser, SEPARATOR, ",", 0)) {
					consume_token(parser);
				}
				else if (check_token_type_and_content(parser, SEPARATOR, ")", 0)) {
					consume_token(parser); // eat closing parenthesis
					break;
				}
			}
			else if (check_token_type_and_content(parser, SEPARATOR, ",", 0)) {
				if (check_token_type_and_content(parser, SEPARATOR, ")", 1)) {
					error_message("error: trailing comma at the end of argument list\n");
				}
				consume_token(parser); // eat the comma
				push_back_item(args, arg);
			}
			else if (check_token_type_and_content(parser, SEPARATOR, ")", 0)) {
				consume_token(parser); // eat closing parenthesis
				push_back_item(args, arg);
				break;
			}
		}
		while (true);

		function_ast_node *fn = create_function_ast_node();
		fn->ret = create_vector();
		fn->numOfReturnValues = 0;
		fn->isTuple = false;

		if (check_token_type_and_content(parser, OPERATOR, ":", 0)) {
			consume_token(parser);
		}
		else {
			error_message("error: function signature missing colon\n");
		}

		// START OF TUPLE
		if (check_token_type_and_content(parser, OPERATOR, "<", 0)) {
			consume_token(parser);
			fn->isTuple = true;

			do {
				if (check_token_type_and_content(parser, OPERATOR, ">", 0)) {
					if (fn->numOfReturnValues < 1) {
						error_message("error: function expects a return type\n");
					}
					consume_token(parser); // eat
					break;
				}

				if (check_token_type(parser, IDENTIFIER, 0)) {
					token *tok = consume_token(parser);
					if (check_token_type_is_valid_data_type(parser, tok)) {
						data_type rawdata_type = match_token_type_to_data_type(parser, tok);
						push_back_item(fn->ret, &rawdata_type);
						fn->numOfReturnValues++;
					}
					else {
						error_message("error: invalid data type specified: `%s`\n", tok->content);
					}
					if (check_token_type_and_content(parser, SEPARATOR, ",", 0)) {
						if (check_token_type_and_content(parser, OPERATOR, ">", 1)) {
							error_message("error: trailing comma in function declaraction\n");
						}
						consume_token(parser);
					}
				}
			}
			while (true);
		}
		else if (check_token_type(parser, IDENTIFIER, 0)) {
			token *returnType = consume_token(parser);
			data_type rawdata_type = match_token_type_to_data_type(parser, returnType);
			push_back_item(fn->ret, &rawdata_type);
			fn->numOfReturnValues += 1;
		}
		else {
			print_current_token(parser);
			error_message("error: function declaration return type expected, found this:\n");
		}

		// start block
		block_ast_node *body = parse_block_ast_node(parser);
		fn->fpn = fpn;
		fn->body = body;
		prepare_ast_node(parser, fn, FUNCTION_AST_NODE);

		return fn;
	}
	else {
		error_message("error: no parameter list provided\n");
	}

	// just in case we fail to parse, free this shit
	free(fpn);
	fpn = NULL;
	error_message("Failed to parse function");
	return NULL;
}

function_callee_ast_node *parse_function_callee_ast_node(parser *parser) {
	// consume function name
	token *callee = match_token_type(parser, IDENTIFIER);

	if (check_token_type_and_content(parser, SEPARATOR, "(", 0)) {
		consume_token(parser);	// eat open bracket

		vector *args = create_vector();

		do {
			// NO ARGUMENTS PROVIDED TO FUNCTION
			if (check_token_type_and_content(parser, SEPARATOR, ")", 0)) {
				consume_token(parser);
				break;
			}

			expression_ast_node *expr = parse_expression_ast_node(parser);

			function_argument_ast_node *arg = create_function_argument_ast_node();
			arg->value = expr;

			if (check_token_type_and_content(parser, SEPARATOR, ",", 0)) {
				if (check_token_type_and_content(parser, SEPARATOR, ")", 1)) {
					error_message("error: trailing comma at the end of argument list\n");
				}
				consume_token(parser); // eat the comma
				push_back_item(args, arg);
			}
			else if (check_token_type_and_content(parser, SEPARATOR, ")", 0)) {
				consume_token(parser); // eat closing parenthesis
				push_back_item(args, arg);
				break;
			}
		}
		while (true);

		parse_optional_semi_colon(parser);

		// woo we got the function
		function_callee_ast_node *fcn = create_function_callee_ast_node();
		fcn->callee = callee;
		fcn->args = args;
		prepare_ast_node(parser, fcn, FUNCTION_CALLEE_AST_NODE);
		return fcn;
	}

	error_message("error: failed to parse function call\n");
	return NULL;
}

function_return_ast_node *parse_return_statement_ast_node(parser *parser) {
	// consume the return keyword
	match_token_type_and_content(parser, IDENTIFIER, RETURN_KEYWORD);

	function_return_ast_node *frn = create_function_return_ast_node();
	frn->returnVals = create_vector();
	frn->numOfReturnValues = 0;

	if (check_token_type_and_content(parser, OPERATOR, "<", 0)) {
		consume_token(parser);

		do {
			if (check_token_type_and_content(parser, OPERATOR, ">", 0)) {
				consume_token(parser);
				parse_optional_semi_colon(parser);
				return frn;
			}

			expression_ast_node *expr = parse_expression_ast_node(parser);
			push_back_item(frn->returnVals, expr);
			if (check_token_type_and_content(parser, SEPARATOR, ",", 0)) {
				if (check_token_type_and_content(parser, OPERATOR, ">", 1)) {
					error_message("error: trailing comma in return statement\n");
				}
				consume_token(parser);
				frn->numOfReturnValues++;
			}
		}
		while (true);
	}
	else {
		// only one return type
		expression_ast_node *expr = parse_expression_ast_node(parser);
		push_back_item(frn->returnVals, expr);
		frn->numOfReturnValues++;

		// consume semi colon if present
		parse_optional_semi_colon(parser);
		return frn;
	}

	error_message("error: failed to parse return statement\n");
	return NULL;
}

void parse_optional_semi_colon(parser *parser) {
	if (check_token_type_and_content(parser, SEPARATOR, ";", 0)) {
		consume_token(parser);
	}
}

statement_ast_node *parse_statement_ast_node(parser *parser) {
	// ret keyword	
	if (check_token_type_and_content(parser, IDENTIFIER, RETURN_KEYWORD, 0)) {
		statement_ast_node *sn = create_statement_ast_node();
		sn->data = parse_return_statement_ast_node(parser); 
		sn->type = FUNCTION_RET_AST_NODE;
		return sn;
	}
	else if (check_token_type_and_content(parser, IDENTIFIER, STRUCT_KEYWORD, 0)) {
		statement_ast_node *sn = create_statement_ast_node();
		sn->data = parse_structure_ast_node(parser);
		sn->type = STRUCT_AST_NODE;
		return sn;
	}
	else if (check_token_type_and_content(parser, IDENTIFIER, FOR_LOOP_KEYWORD, 0)) {
		return parse_for_loop_ast_node(parser);
	}
	else if (check_token_type_and_content(parser, IDENTIFIER, INFINITE_LOOP_KEYWORD, 0)) {
		return parse_infinite_loop_ast_node(parser);
	}
	else if (check_token_type_and_content(parser, IDENTIFIER, ENUM_KEYWORD, 0)) {
		statement_ast_node *sn = create_statement_ast_node();
		sn->data = parse_enumeration_ast_node(parser);
		sn->type = ENUM_AST_NODE;
		return sn;
	}
	else if (check_token_type_and_content(parser, IDENTIFIER, BREAK_KEYWORD, 0)) {
		// consume the token
		consume_token(parser);

		statement_ast_node *sn = create_statement_ast_node();
		sn->data = create_break_ast_node();
		sn->type = BREAK_AST_NODE;

		parse_optional_semi_colon(parser);

		return sn;
	}
	else if (check_token_type(parser, IDENTIFIER, 0)) {
		token *tok = peek_at_token_stream(parser, 0);
		
		// variable reassignment
		if (check_token_type_and_content(parser, OPERATOR, "=", 1)) {
			statement_ast_node *sn = create_statement_ast_node();
			sn->data = parse_reassignment_statement_ast_node(parser);
			sn->type = VARIABLE_REASSIGN_AST_NODE;
			return sn;
		}
		// function call
		else if (check_token_type_and_content(parser, SEPARATOR, "(", 1)) {
			statement_ast_node *sn = create_statement_ast_node();
			sn->data = parse_function_callee_ast_node(parser);
			sn->type = FUNCTION_CALLEE_AST_NODE;
			return sn;
		}
		// local variable
		else if (check_token_type_is_valid_data_type(parser, tok)) {
			return parse_variable_ast_node(parser, false);
		}
		// fuck knows
		else {
			error_message("error: unrecognized identifier %s\n", tok->content);
		}
	}

	token *tok = peek_at_token_stream(parser, 0);
	error_message("error: unrecognized token %s(%s)\n", tok->content, token_NAMES[tok->type]);
	return NULL;
}

variable_reassignment_ast_node *parse_reassignment_statement_ast_node(parser *parser) {
	if (check_token_type(parser, IDENTIFIER, 0)) {
		token *variableName = consume_token(parser);

		if (check_token_type_and_content(parser, OPERATOR, "=", 0)) {
			consume_token(parser);

			expression_ast_node *expr = parse_expression_ast_node(parser);

			parse_optional_semi_colon(parser);

			variable_reassignment_ast_node *vrn = create_variable_reassign_ast_node();
			vrn->name = variableName;
			vrn->expr = expr;
			return vrn;
		}
	}

	error_message("error: failed to parse variable reassignment\n");
	return NULL;
}

void start_parsing_token_stream(parser *parser) {
	while (parser->parsing) {
		// get current token
		token *tok = get_vector_item(parser->token_stream, parser->token_index);

		// TODO: improve this
		switch (tok->type) {
			case IDENTIFIER:
				if (!strcmp(tok->content, FUNCTION_KEYWORD)) {
					parse_function_ast_node(parser);
				}
				else if (check_token_type_and_content(parser, IDENTIFIER, STRUCT_KEYWORD, 0)) {
					prepare_ast_node(parser, parse_structure_ast_node(parser), STRUCT_AST_NODE);
				}
				else if (check_token_type_and_content(parser, IDENTIFIER, ENUM_KEYWORD, 0)) {
					prepare_ast_node(parser, parse_enumeration_ast_node(parser), ENUM_AST_NODE);
				}
				else if (check_token_type_is_valid_data_type(parser, tok) || check_token_type_and_content(parser, IDENTIFIER, CONSTANT_KEYWORD, 0)) {
					parse_variable_ast_node(parser, true);
				}
				else if (check_token_type_and_content(parser, OPERATOR, "=", 1)) {
					parse_reassignment_statement_ast_node(parser);
				}
				else if (check_token_type_and_content(parser, SEPARATOR, "(", 1)) {
					prepare_ast_node(parser, parse_function_callee_ast_node(parser), FUNCTION_CALLEE_AST_NODE);
				}
				else {
					error_message("error: unrecognized identifier found: `%s`\n", tok->content);
				}
				break;
			case END_OF_FILE:
				parser->parsing = false;
				break;
		}
	}
}

bool check_token_type_is_valid_data_type(parser *parser, token *tok) {
	int size = sizeof(DATA_TYPES) / sizeof(DATA_TYPES[0]);
	int i;
	for (i = 0; i < size; i++) {
		if (!strcmp(tok->content, DATA_TYPES[i])) {
			return true;
		}
	}
	return false;
}

data_type match_token_type_to_data_type(parser *parser, token *tok) {
	int size = sizeof(DATA_TYPES) / sizeof(DATA_TYPES[0]);
	int i;
	for (i = 0; i < size; i++) {
		if (!strcmp(tok->content, DATA_TYPES[i])) {
			return i;
		}
	}
	error_message("error: invalid data type specified: %s!\n", tok->content);
	return 0;
}

void prepare_ast_node(parser *parser, void *data, ast_node_type type) {
	ast_node *ast_node = malloc(sizeof(*ast_node));
	ast_node->data = data;
	ast_node->type = type;
	push_back_item(parser->parse_tree, ast_node);
}

void remove_ast_node(ast_node *ast_node) {
	/**
	 * This could probably be a lot more cleaner
	 */
	if (!ast_node->data) {
		switch (ast_node->type) {
			case EXPRESSION_AST_NODE:
				destroy_expression_ast_node(ast_node->data);
				break;
			case VARIABLE_DEF_AST_NODE: 
				destroy_variable_define_ast_node(ast_node->data);
				break;
			case VARIABLE_DEC_AST_NODE:
				destroy_variable_declare_ast_node(ast_node->data);
				break;
			case FUNCTION_ARG_AST_NODE:
				destroy_function_argument_ast_node(ast_node->data);
				break;
			case FUNCTION_AST_NODE:
				destroy_function_ast_node(ast_node->data);
				break;
			case FUNCTION_PROT_AST_NODE:
				destroy_function_prototype_ast_node(ast_node->data);
				break;
			case BLOCK_AST_NODE:
				destroy_block_ast_node(ast_node->data);
				break;
			case FUNCTION_CALLEE_AST_NODE:
				destroy_function_callee_ast_node(ast_node->data);
				break;
			case FUNCTION_RET_AST_NODE:
				destroy_function_return_ast_node(ast_node->data);
				break;
			case FOR_LOOP_AST_NODE:
				destroy_for_loop_ast_node(ast_node->data);
				break;
			case VARIABLE_REASSIGN_AST_NODE:
				destroy_variable_reassign_ast_node(ast_node->data);
				break;
			case INFINITE_LOOP_AST_NODE:
				destroy_infinite_loop_ast_node(ast_node->data);
				break;
			case BREAK_AST_NODE:
				destroy_break_ast_node(ast_node->data);
				break;
			case ENUM_AST_NODE:
				destroy_enumeration_ast_node(ast_node->data);
				break;
			default:
				error_message("attempting to remove unrecognized ast_node(%d)?\n", ast_node->type);
				break;
		}
	}
	free(ast_node);
}

void destroy_parser(parser *parser) {
	int i;
	for (i = 0; i < parser->token_stream->size; i++) {
		token *tok = get_vector_item(parser->token_stream, i);
		destroy_token(tok);
	}
	destroy_vector(parser->token_stream);

	for (i = 0; i < parser->parse_tree->size; i++) {
		ast_node *ast_node = get_vector_item(parser->parse_tree, i);
		remove_ast_node(ast_node);
	}
	destroy_vector(parser->parse_tree);

	free(parser);
}