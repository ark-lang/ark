#include "parser.h"

/** List of data types */
static const char* DATA_TYPES[] = {
	"int", "str", "double", "float", "bool",
	"void", "char"
};

/** Supported Operators */
static char* SUPPORTED_OPERANDS[] = {
	"++", "--", 
	"+=", "-=", "*=", "/=", "%=",
	"+", "-", "&", "-", "*", "/", "%", "^", "**",
	">", "<", ">=", "<=", "==", "!=", "&&", "||",
};

/** UTILITY FOR AST NODES */

void parser_error(parser *parser, char *msg, token *tok, bool fatal_error) {
	error_message("%d:%d %s", tok->line_number, tok->char_number, msg);
	char *error = get_token_context(parser->token_stream, tok, true);
	printf("\t%s\n", error);
	parser->exit_on_error = true;
	free(error);
	if (fatal_error) {
		exit(1);
	}
}

void *allocate_ast_node(size_t sz, const char* readable_type) {
	// dont use safe malloc here because we can provide additional
	// error info
	void *ret = malloc(sz);
	if (!ret) {
		fprintf(stderr, "malloc: failed to allocate memory for %s", readable_type);
		exit(1);
	}
	return ret;
}

infinite_loop_ast_node *create_infinite_loop_ast_node() {
	infinite_loop_ast_node *iln = allocate_ast_node(sizeof(infinite_loop_ast_node), "infinite loop");
	iln->body = NULL;
	return iln;
}

break_ast_node *create_break_ast_node() {
	break_ast_node *bn = allocate_ast_node(sizeof(break_ast_node), "break");
	return bn;
}

continue_ast_node *create_continue_ast_node() {
	continue_ast_node *cn = allocate_ast_node(sizeof(continue_ast_node), "continue");
	return cn;
}

variable_reassignment_ast_node *create_variable_reassign_ast_node() {
	variable_reassignment_ast_node *vrn = allocate_ast_node(sizeof(variable_reassignment_ast_node), "variable reassignment");
	vrn->name = NULL;
	vrn->expr = NULL;
	return vrn;
}

statement_ast_node *create_statement_ast_node(void *data, ast_node_type type) {
	statement_ast_node *sn = allocate_ast_node(sizeof(statement_ast_node), "statement");
	sn->data = data;
	sn->type = type;
	return sn;
}

function_return_ast_node *create_function_return_ast_node() {
	function_return_ast_node *frn = allocate_ast_node(sizeof(function_return_ast_node), "function return");
	frn->return_val = NULL;
	return frn;
}

expression_ast_node *create_expression_ast_node() {
	expression_ast_node *expr = allocate_ast_node(sizeof(expression_ast_node), "expression");
	expr->value = NULL;
	expr->lhand = NULL;
	expr->rhand = NULL;
	expr->pointer_option = UNSPECIFIED;
	return expr;
}

variable_define_ast_node *create_variable_define_ast_node() {
	variable_define_ast_node *vdn = allocate_ast_node(sizeof(variable_define_ast_node), "variable definition");
	vdn->name = NULL;
	vdn->is_constant = false;
	vdn->is_global = false;
	return vdn;
}

variable_declare_ast_node *create_variable_declare_ast_node() {
	variable_declare_ast_node *vdn = allocate_ast_node(sizeof(variable_declare_ast_node), "variable declaration");
	vdn->vdn = NULL;
	vdn->expression = NULL;
	return vdn;
}

function_argument_ast_node *create_function_argument_ast_node() {
	function_argument_ast_node *fan = allocate_ast_node(sizeof(function_argument_ast_node), "function argument");
	fan->name = NULL;
	fan->value = NULL;
	return fan;
}

function_callee_ast_node *create_function_callee_ast_node() {
	function_callee_ast_node *fcn = allocate_ast_node(sizeof(function_callee_ast_node), "function callee");
	fcn->callee = NULL;
	fcn->args = NULL;
	return fcn;
}

block_ast_node *create_block_ast_node() {
	block_ast_node *bn = allocate_ast_node(sizeof(block_ast_node), "block");
	bn->statements = NULL;
	return bn;
}

function_prototype_ast_node *create_function_prototype_ast_node() {
	function_prototype_ast_node *fpn = allocate_ast_node(sizeof(function_prototype_ast_node), "function prototype");
	fpn->args = NULL;
	fpn->name = NULL;
	return fpn;
}

enumeration_ast_node *create_enumeration_ast_node() {
	enumeration_ast_node *en = allocate_ast_node(sizeof(enumeration_ast_node), "enum");
	en->name = NULL;
	en->enum_items = create_vector();
	return en;
}

enum_item *create_enum_item(char *name, int value) {
	enum_item *ei = allocate_ast_node(sizeof(enum_item), "enum item");
	ei->name = name;
	ei->value = value;
	return ei;
}

function_ast_node *create_function_ast_node() {
	function_ast_node *fn = allocate_ast_node(sizeof(function_ast_node), "function");
	fn->fpn = NULL;
	fn->body = NULL;
	return fn;
}

for_loop_ast_node *create_for_loop_ast_node() {
	for_loop_ast_node *fln = allocate_ast_node(sizeof(for_loop_ast_node), "for loop");
	return fln;
}

structure_ast_node *create_structure_ast_node() {
	structure_ast_node *sn = allocate_ast_node(sizeof(structure_ast_node), "struct");
	sn->statements = create_vector();
	return sn;
}

if_statement_ast_node *create_if_statement_ast_node() {
	if_statement_ast_node *isn = allocate_ast_node(sizeof(if_statement_ast_node), "if statement");
	return isn;	
}

while_ast_node *create_while_ast_node() {
	while_ast_node *wn = allocate_ast_node(sizeof(while_ast_node), "while loop");
	return wn;
}

match_case_ast_node *create_match_case_ast_node() {
	match_case_ast_node *mcn = allocate_ast_node(sizeof(match_case_ast_node), "match case");
	return mcn;
}

match_ast_node *create_match_ast_node() {
	match_ast_node *mn = allocate_ast_node(sizeof(match_ast_node), "match");
	mn->cases = create_vector();
	return mn;
}

void destroy_variable_reassign_ast_node(variable_reassignment_ast_node *vrn) {
	if (vrn) {
		if (vrn->expr) {
			destroy_expression_ast_node(vrn->expr);
		}
		free(vrn);
	}
}

void destroy_for_loop_ast_node(for_loop_ast_node *fln) {
	if (fln) {
		destroy_vector(fln->params);
		destroy_block_ast_node(fln->body);
		free(fln);
	}
}

void destroy_break_ast_node(break_ast_node *bn) {
	if (bn) {
		free(bn);
	}
}

void destroy_continue_ast_node(continue_ast_node *cn) {
	if (cn) {
		free(cn);
	}
}

void destroy_statement_ast_node(statement_ast_node *sn) {
	if (sn) {
		if (sn->data) {
			switch (sn->type) {
				case VARIABLE_DEF_AST_NODE: destroy_variable_define_ast_node(sn->data); break;
				case VARIABLE_DEC_AST_NODE: destroy_variable_declare_ast_node(sn->data); break;
				case FUNCTION_CALLEE_AST_NODE: destroy_function_callee_ast_node(sn->data); break;
				case FUNCTION_RET_AST_NODE: destroy_function_ast_node(sn->data); break;
				case VARIABLE_REASSIGN_AST_NODE: destroy_variable_reassign_ast_node(sn->data); break;
				case FOR_LOOP_AST_NODE: destroy_for_loop_ast_node(sn->data); break;
				case INFINITE_LOOP_AST_NODE: destroy_infinite_loop_ast_node(sn->data); break;
				case BREAK_AST_NODE: destroy_break_ast_node(sn->data); break;
				case CONTINUE_AST_NODE: destroy_continue_ast_node(sn->data); break;
				case ENUM_AST_NODE: destroy_enumeration_ast_node(sn->data); break;
				case IF_STATEMENT_AST_NODE: destroy_if_statement_ast_node(sn->data); break;
				case MATCH_STATEMENT_AST_NODE: destroy_match_ast_node(sn->data); break;
				case WHILE_LOOP_AST_NODE: destroy_while_ast_node(sn->data); break;
				default: printf("trying to destroy unrecognized statement node %d\n", sn->type); break;
			}
		}
		free(sn);
	}
}

void destroy_function_return_ast_node(function_return_ast_node *frn) {
	if (frn) {
		destroy_expression_ast_node(frn->return_val);
		free(frn);
	}
}

void destroy_expression_ast_node(expression_ast_node *expr) {
	if (expr) {
		if (expr->function_call && expr->type == EXPR_FUNCTION_CALL) {
			destroy_function_callee_ast_node(expr->function_call);
		}
		if (expr->lhand) {
			destroy_expression_ast_node(expr->lhand);
		}
		if (expr->rhand) {
			destroy_expression_ast_node(expr->rhand);
		}
		free(expr);
	}
}

void destroy_variable_define_ast_node(variable_define_ast_node *vdn) {
	if (vdn) {
		free(vdn);
	}
}

void destroy_variable_declare_ast_node(variable_declare_ast_node *vdn) {
	if (vdn) {
		if (vdn->vdn) {
			destroy_variable_define_ast_node(vdn->vdn);
		}
		if (vdn->expression) {
			destroy_expression_ast_node(vdn->expression);
		}
		free(vdn);
	}
}

void destroy_function_argument_ast_node(function_argument_ast_node *fan) {
	if (fan) {
		if (fan->value) {
			destroy_expression_ast_node(fan->value);
		}
		free(fan);
	}
}

void destroy_block_ast_node(block_ast_node *bn) {
	if (bn) {
		if (bn->statements) {
			destroy_vector(bn->statements);
		}
		free(bn);
	}
}

void destroy_infinite_loop_ast_node(infinite_loop_ast_node *iln) {
	if (iln) {
		if (iln->body) {
			destroy_block_ast_node(iln->body);
		}
		free(iln);
	}
}

void destroy_function_prototype_ast_node(function_prototype_ast_node *fpn) {
	if (fpn) {
		if (fpn->args) {
			int i;
			for (i = 0; i < fpn->args->size; i++) {
				statement_ast_node *sn = get_vector_item(fpn->args, i);
				if (sn) {
					destroy_statement_ast_node(sn);
				}
			}
			destroy_vector(fpn->args);
		}
		free(fpn);
	}
}

void destroy_function_ast_node(function_ast_node *fn) {
	if (fn) {
		if (fn->fpn) {
			destroy_function_prototype_ast_node(fn->fpn);
		}
		if (fn->body) {
			destroy_block_ast_node(fn->body);
		}
		free(fn);
	}
}

void destroy_function_callee_ast_node(function_callee_ast_node *fcn) {
	if (fcn) {
		if (fcn->args) {
			destroy_vector(fcn->args);
		}
		free(fcn);
	}
}

void destroy_structure_ast_node(structure_ast_node *sn) {
	if (sn) {
		if (sn->statements) {
			destroy_vector(sn->statements);
		}
		free(sn);
	}
}

void destroy_enumeration_ast_node(enumeration_ast_node *en) {
	if (en) {
		if (en->enum_items) {
			int i;
			for (i = 0; i < en->enum_items->size; i++) {
				destroy_enum_item(get_vector_item(en->enum_items, i));
			}
			destroy_vector(en->enum_items);
		}
		free(en);
	}
}

void destroy_if_statement_ast_node(if_statement_ast_node *isn) {
	if (isn) {
		if (isn->condition) {
			destroy_expression_ast_node(isn->condition);
		}
		if (isn->body) {
			destroy_block_ast_node(isn->body);
		}
		free(isn);
	}
}

void destroy_while_ast_node(while_ast_node *wn) {
	if (wn) {
		if (wn->condition) {
			destroy_expression_ast_node(wn->condition);
		}
		if (wn->body) {
			destroy_block_ast_node(wn->body);
		}
		free(wn);
	}
}

void destroy_match_case_ast_node(match_case_ast_node *mcn) {
	if (mcn) {
		if (mcn->condition) {
			destroy_expression_ast_node(mcn->condition);
		}
		if (mcn->body) {
			destroy_block_ast_node(mcn->body);
		}
		free(mcn);
	}
}

void destroy_match_ast_node(match_ast_node *mn) {
	if (mn) {
		if (mn->condition) {
			destroy_expression_ast_node(mn->condition);
		}
		if (mn->cases) {
			int i;
			for (i = 0; i < mn->cases->size; i++) {
				destroy_match_case_ast_node(get_vector_item(mn->cases, i));
			}
			destroy_vector(mn->cases);
		}
		free(mn);
	}
}

void destroy_enum_item(enum_item *ei) {
	if (ei) {
		free(ei);
	}
}

/** END AST_NODE FUNCTIONS */

parser *create_parser(vector *token_stream) {
	parser *parser = safe_malloc(sizeof(*parser));
	parser->token_stream = token_stream;
	parser->parse_tree = create_vector();
	parser->token_index = 0;
	parser->parsing = true;
	parser->exit_on_error = true;
	parser->sym_table = create_hashmap(16);
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
		parser_error(parser, "unrecognized token found", tok, true);
		return NULL;
	}
}

token *expect_token_content(parser *parser, char *content) {
	token *tok = peek_at_token_stream(parser, 1);
	if (!strcmp(tok->content, content)) {
		return consume_token(parser);
	}
	else {
		parser_error(parser, "unexpected token found", tok, true);
		return NULL;
	}
}

token *expect_token_type_and_content(parser *parser, token_type type, char *content) {
	token *tok = peek_at_token_stream(parser, 1);
	if (tok->type == type && !strcmp(tok->content, content)) {
		return consume_token(parser);
	}
	else {
		parser_error(parser, "unexpected token found", tok, true);
		return NULL;
	}
}

token *match_token_type(parser *parser, token_type type) {
	token *tok = peek_at_token_stream(parser, 0);
	if (tok->type == type) {
		return consume_token(parser);
	}
	else {
		parser_error(parser, "unexpected token found", tok, true);
		return NULL;
	}
}

token *match_token_content(parser *parser, char *content) {
	token *tok = peek_at_token_stream(parser, 0);
	if (!strcmp(tok->content, content)) {
		return consume_token(parser);
	}
	else {
		parser_error(parser, "unexpected token found", tok, true);
		return NULL;
	}
}

token *match_token_type_and_content(parser *parser, token_type type, char *content) {
	token *tok = peek_at_token_stream(parser, 0);
	if (tok->type == type && !strcmp(tok->content, content)) {
		return consume_token(parser);
	}
	else {
		parser_error(parser, "unexpected token found", tok, true);
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

int parse_operand(parser *parser) {
	token *tok = peek_at_token_stream(parser, 0);

	int i;
	int operand_list_size = sizeof(SUPPORTED_OPERANDS) / sizeof(SUPPORTED_OPERANDS[0]);
	for (i = 0; i < operand_list_size; i++) {
		if (!strcmp(SUPPORTED_OPERANDS[i], tok->content)) {
			consume_token(parser);
			return i;
		}
	}

	parser_error(parser, "unsupported operator specified", tok, true);
	return OPER_ERRORNEOUS;
}

while_ast_node *parse_while_loop(parser *parser) {
	while_ast_node *while_loop = create_while_ast_node();

	match_token_type_and_content(parser, IDENTIFIER, WHILE_LOOP_KEYWORD);

	while_loop->condition = parse_expression_ast_node(parser);
	while_loop->body = parse_block_ast_node(parser);

	return while_loop;
}

if_statement_ast_node *parse_if_statement_ast_node(parser *parser) {
	if_statement_ast_node *en = create_if_statement_ast_node();

	match_token_type_and_content(parser, IDENTIFIER, IF_KEYWORD);

	en->condition = parse_expression_ast_node(parser);
	en->body = parse_block_ast_node(parser);
	en->statment_type = IF_STATEMENT;

	if (check_token_type_and_content(parser, IDENTIFIER, ELSE_KEYWORD, 0)) {
		consume_token(parser);
		en->statment_type = ELSE_STATEMENT;
		en->else_statement = parse_block_ast_node(parser);
	}

	return en;
}

match_case_ast_node *parse_match_case_ast_node(parser *parser) {
	match_case_ast_node *case_node = create_match_case_ast_node();

	case_node->condition = parse_expression_ast_node(parser);
	case_node->body = parse_block_ast_node(parser);

	return case_node;
}

match_ast_node *parse_match_ast_node(parser *parser) {
	match_ast_node *mn = create_match_ast_node();

	match_token_type_and_content(parser, IDENTIFIER, MATCH_KEYWORD);

	expression_ast_node *expr = parse_expression_ast_node(parser);
	mn->condition = expr;
	mn->cases = create_vector();

	if (check_token_type_and_content(parser, SEPARATOR, BLOCK_OPENER, 0)) {
		consume_token(parser);

		do {
			if (check_token_type_and_content(parser, SEPARATOR, BLOCK_CLOSER, 0)) {
				consume_token(parser);
				break;
			}

			push_back_item(mn->cases, parse_match_case_ast_node(parser));
		}
		while (true);
	}
	else {
		parser_error(parser, "match expected block denoted with `{}`", consume_token(parser), true);
	}

	return mn;
}

enumeration_ast_node *parse_enumeration_ast_node(parser *parser) {
	enumeration_ast_node *en = create_enumeration_ast_node();

	match_token_type_and_content(parser, IDENTIFIER, ENUM_KEYWORD); // ENUM

	// ENUMERATIONS NAME
	if (check_token_type(parser, IDENTIFIER, 0)) {
		token *enum_dec = consume_token(parser);
		en->name = enum_dec;

		// OPEN OF ENUM BLOCK
		if (check_token_type_and_content(parser, SEPARATOR, BLOCK_OPENER, 0)) {
			consume_token(parser);

			// LOOP
			do {

				// eat the last brace
				if (check_token_type_and_content(parser, SEPARATOR, BLOCK_CLOSER, 0)) {
					consume_token(parser);
					break;
				}

				if (check_token_type(parser, IDENTIFIER, 0)) {
					token *enum_item_name = consume_token(parser);

					// setting the enum = to a value
					if (check_token_type_and_content(parser, OPERATOR, ASSIGNMENT_OPERATOR, 0)) {
						consume_token(parser);

						// to a number
						if (check_token_type(parser, NUMBER, 0)) {
							token *enum_item_value = consume_token(parser);

							// convert to int
							int enum_item_value_as_int = atoi(enum_item_value->content);

							// if we already have items in our enum
							// make sure there are no duplicate values or names
							if (en->enum_items->size >= 1) {
								enum_item *prev_item = get_vector_item(en->enum_items, en->enum_items->size - 1);
								int prev_item_value = prev_item->value;
								char *prev_item_name = prev_item->name;

								// validate names are not duplicate
								if (!strcmp(prev_item_name, enum_item_name->content)) {
									parser_error(parser, "duplicate item in enumeration", enum_item_name, false);
								}

								// validate values are not duplicate
								if (prev_item_value == enum_item_value_as_int) {
									parser_error(parser, "duplicate item value in enumeration", enum_item_name, false);
								}
							}

							// push it back
							enum_item *item = create_enum_item(enum_item_name->content, enum_item_value_as_int);
							push_back_item(en->enum_items, item);
						}
						else {
							parser_error(parser, "invalid integer literal assigned to enumeration item", consume_token(parser), false);
						}

					}
					// ENUM_ITEM with no assignment
					else {
						int enum_item_value_as_int = 0;

						if (en->enum_items->size >= 1) {
							enum_item *prev_item = get_vector_item(en->enum_items, en->enum_items->size - 1);
							enum_item_value_as_int = prev_item->value + 1;
							char *prev_item_name = prev_item->name;

							// validate name
							if (!strcmp(prev_item_name, enum_item_name->content)) {
								parser_error(parser, "duplicate item in enumeration", enum_item_name, false);
							}
						}

						enum_item *item = create_enum_item(enum_item_name->content, enum_item_value_as_int);
						push_back_item(en->enum_items, item);
					}
				}

				if (check_token_type_and_content(parser, SEPARATOR, COMMA_SEPARATOR, 0)) {
					consume_token(parser);
					if (check_token_type_and_content(parser, SEPARATOR, BLOCK_CLOSER, 0)) {
						parser_error(parser, "trailing comma in enumeration", consume_token(parser), false);
						break;
					}
				}
				
				// we've finished parsing jump out
				if (check_token_type_and_content(parser, SEPARATOR, BLOCK_CLOSER, 0)) {
					consume_token(parser);
					break;
				}

			}
			while (true);

			// empty enum, throw an error.
			if (en->enum_items->size == 0) {
				parser_error(parser, "empty enumeration", consume_token(parser), false);
			}
		}
	}

	return en;
}

structure_ast_node *parse_structure_ast_node(parser *parser) {
	match_token_type_and_content(parser, IDENTIFIER, STRUCT_KEYWORD);
	token *struct_name = match_token_type(parser, IDENTIFIER);

	structure_ast_node *sn = create_structure_ast_node();
	sn->struct_name = struct_name->content;

	// parses a block of statements
	if (check_token_type_and_content(parser, SEPARATOR, BLOCK_OPENER, 0)) {
		consume_token(parser);

		do {
			if (check_token_type_and_content(parser, SEPARATOR, BLOCK_CLOSER, 0)) {
				consume_token(parser);
				break;
			}

			push_back_item(sn->statements, parse_variable_ast_node(parser, false));
		}
		while (true);
	}

	return sn;
}

statement_ast_node *parse_for_loop_ast_node(parser *parser) {
	// for token
	match_token_type_and_content(parser, IDENTIFIER, FOR_LOOP_KEYWORD);			// FOR KEYWORd

	// todo inferred data types
	token *index_name = match_token_type(parser, IDENTIFIER);					// INDEX_NAME

	match_token_type_and_content(parser, OPERATOR, ":");						// PARAMS

	// create node with the stuff we just got
	for_loop_ast_node *fln = create_for_loop_ast_node();
	fln->type = TYPE_NULL;			// we don't know yet
	fln->index_name = index_name;
	fln->params = create_vector();

	// consume the args
	if (check_token_type_and_content(parser, SEPARATOR, "(", 0)) {
		token *arg_opener = consume_token(parser);

		int param_count = 0;

		do {
			if (param_count > 3) {
				parser_error(parser, "too many parameters passed to for loop", consume_token(parser), false);
			}
			if (check_token_type_and_content(parser, SEPARATOR, ")", 0)) {
				if (param_count < 2) {
					parser_error(parser, "too few parameters passed to for loop", arg_opener, false);
				}
				consume_token(parser);
				break;
			}

			if (check_token_type(parser, IDENTIFIER, 0)) {
				token *tok = consume_token(parser);
				fln->type = tok->type;
				push_back_item(fln->params, tok);
				if (check_token_type_and_content(parser, SEPARATOR, ",", 0)) {
					if (check_token_type_and_content(parser, SEPARATOR, ")", 1)) {
						parser_error(parser, "trailing comma in for loop declaration", consume_token(parser), false);
					}
					consume_token(parser);
				}
			}
			else if (check_token_type(parser, NUMBER, 0)) {
				token *tok = consume_token(parser);
				fln->type = tok->type;
				push_back_item(fln->params, tok);
				if (check_token_type_and_content(parser, SEPARATOR, ",", 0)) {
					if (check_token_type_and_content(parser, SEPARATOR, ")", 1)) {
						parser_error(parser, "trailing comma in for loop declaration", consume_token(parser), false);
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
						parser_error(parser, "trailing comma in for loop declaration", consume_token(parser), false);
					}
					consume_token(parser);
				}
			}
			else {
				parser_error(parser, "expected a number literal or a variable in for loop parameters", consume_token(parser), false);
				break;
			}

			param_count++;
		}
		while (true);

		fln->body = parse_block_ast_node(parser);

		return create_statement_ast_node(fln, FOR_LOOP_AST_NODE);
	}

	parser_error(parser, "failed to parse for loop", consume_token(parser), true);
	return NULL;
}

expression_ast_node *parse_number_expression(parser *parser) {
	expression_ast_node *expr = create_expression_ast_node(); // the final expression
	expr->type = EXPR_NUMBER;
	expr->value = consume_token(parser);
	return expr;
}

expression_ast_node *parse_paren_expression(parser *parser) {
	expression_ast_node *expr = create_expression_ast_node(); // the final expression
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

		parser_error(parser, "missing closing parenthesis for expression", consume_token(parser), false);
	}

	parser_error(parser, "could not parse parenthesis expressions", consume_token(parser), true);
	return NULL;
}

expression_ast_node *parse_string_expression(parser *parser) {
	expression_ast_node *expr = create_expression_ast_node(); // the final expression
	expr->type = EXPR_STRING;
	expr->value = consume_token(parser);
	return expr;
}

expression_ast_node *parse_character_expression(parser *parser) {
	expression_ast_node *expr = create_expression_ast_node(); // the final expression
	expr->type = EXPR_CHARACTER;
	expr->value = consume_token(parser);
	return expr;
}

expression_ast_node *parse_identifier_expression(parser *parser) {
	expression_ast_node *expr = create_expression_ast_node(); // the final expression
	
	// function call
	if (check_token_type_and_content(parser, SEPARATOR, "(", 1)) {
		expr->type = EXPR_FUNCTION_CALL;
		expr->function_call = parse_function_callee_ast_node(parser);
		return expr;
	}

	// it's going to be a variable
	expr->type = EXPR_VARIABLE;
	expr->value = consume_token(parser);

	return expr;
}

expression_ast_node *parse_expression_ast_node(parser *parser) {

	// we can probably replace the above functions in the standard library with
	// a macro which does the translation for us

	// false
	if (check_token_type_and_content(parser, IDENTIFIER, FALSE_KEYWORD, 0)) {
		expression_ast_node *expr= create_expression_ast_node();
		token *tok = consume_token(parser);
		strcpy(tok->content, "0"); // change false to 0
		expr->type = EXPR_NUMBER;
		expr->value = tok;
		return expr;
	}

	// true
	if (check_token_type_and_content(parser, IDENTIFIER, TRUE_KEYWORD, 0)) {
		expression_ast_node *expr = create_expression_ast_node();
		token *tok = consume_token(parser);
		strcpy(tok->content, "1"); // false to 1
		expr->type = EXPR_NUMBER;
		expr->value = tok;
		return expr;
	}

	if (check_token_type_and_content(parser, OPERATOR, ADDRESS_OF_OPERATOR, 0)) {
		if (check_token_type(parser, IDENTIFIER, 1)) {
			// todo lookup if the value given
			// a) exists
			// b) is a variable (not pointer)
			consume_token(parser);
			expression_ast_node *expr = parse_identifier_expression(parser);
			expr->pointer_option = ADDRESS_OF;
			return expr;
		}
		else {
			parser_error(parser, "cannot get address of constant value, or pointer", consume_token(parser), true);
		}
	}

	if (check_token_type_and_content(parser, OPERATOR, POINTER_OPERATOR, 0)) {
		if (check_token_type(parser, IDENTIFIER, 1)) {
			// todo lookup if the value given
			// a) exists
			// b) is a pointer
			consume_token(parser);
			expression_ast_node *expr = parse_identifier_expression(parser);
			expr->pointer_option = DEREFERENCE;
			return expr;
		}
		else {
			parser_error(parser, "invalid type argument of unary `^`", consume_token(parser), true);
		}
	}

	// number literal
	if (check_token_type(parser, NUMBER, 0)) {
		return parse_number_expression(parser);
	}

	// string literal
	if (check_token_type(parser, STRING, 0)) {
		return parse_string_expression(parser);
	}
	
	// character
	if (check_token_type(parser, CHARACTER, 0)) {
		return parse_character_expression(parser);
	}

	// identifier
	if (check_token_type(parser, IDENTIFIER, 0)) {
		return parse_identifier_expression(parser);
	}

	// expression with parenthesis
	if (check_token_type_and_content(parser, SEPARATOR, "(", 0)) {
		return parse_paren_expression(parser);
	}

	// error here
	return NULL;
}

void *parse_variable_ast_node(parser *parser, bool global) {
	bool is_constant = false;

	if (check_token_type_and_content(parser, IDENTIFIER, CONSTANT_KEYWORD, 0)) {
		consume_token(parser);
		is_constant = true;
	}

	// consume the int data type
	token *variable_data_type = match_token_type(parser, IDENTIFIER);
	variable_define_ast_node *def = create_variable_define_ast_node();
	data_type data_type_raw = match_token_type_to_data_type(parser, variable_data_type);

	// is a pointer
	bool is_pointer = false;
	if (check_token_type_and_content(parser, OPERATOR, POINTER_OPERATOR, 0)) {
		is_pointer = true;
		consume_token(parser);
	}

	// name of the variable
	token *variable_name_token = match_token_type(parser, IDENTIFIER);

	if (check_token_type_and_content(parser, OPERATOR, ASSIGNMENT_OPERATOR, 0)) {
		// consume the equals sign
		consume_token(parser);

		// create variable define ast_node
		def->is_constant = is_constant;
		def->type = data_type_raw;
		def->name = variable_name_token->content;
		def->is_global = global;
		def->is_pointer = is_pointer;

		// create the variable declare ast_node
		variable_declare_ast_node *dec = create_variable_declare_ast_node();
		dec->vdn = def;
		dec->expression = parse_expression_ast_node(parser);

		// this is weird, we can probably clean this up
		if (global) {
			prepare_ast_node(parser, dec, VARIABLE_DEC_AST_NODE);
			return dec;
		}

		// not global, pop it as a statement node
		return create_statement_ast_node(dec, VARIABLE_DEC_AST_NODE);
	}
	else {
		// create variable define ast_node
		def->is_constant = is_constant;
		def->type = data_type_raw;
		def->name = variable_name_token->content;
		def->is_global = global;
		def->is_pointer = is_pointer;

		if (global) {
			prepare_ast_node(parser, def, VARIABLE_DEF_AST_NODE);
			return def;
		}

		// not global, pop it as a statement node
		return create_statement_ast_node(def, VARIABLE_DEF_AST_NODE);
	}
}

block_ast_node *parse_block_ast_node(parser *parser) {
	block_ast_node *block = create_block_ast_node();
	block->statements = create_vector();
	block->single_statement = false;

	if (check_token_type_and_content(parser, OPERATOR, SINGLE_STATEMENT, 0)) {
		consume_token(parser);
		push_back_item(block->statements, parse_statement_ast_node(parser));
		block->single_statement = true;
	}
	else if (check_token_type_and_content(parser, SEPARATOR, BLOCK_OPENER, 0)) {
		consume_token(parser);

		do {
			// check if block is empty before we try parse some statements
			if (check_token_type_and_content(parser, SEPARATOR, BLOCK_CLOSER, 0)) {
				consume_token(parser);
				break;
			}

			push_back_item(block->statements, parse_statement_ast_node(parser));
		}
		while (true);
	}
	else {
		parser_error(parser, "expected a multi-block or single-block statement", consume_token(parser), true);
	}

	return block;
}

statement_ast_node *parse_infinite_loop_ast_node(parser *parser) {
	match_token_type(parser, IDENTIFIER);

	block_ast_node *body = parse_block_ast_node(parser);

	infinite_loop_ast_node *iln = create_infinite_loop_ast_node();
	iln->body = body;

	return create_statement_ast_node(iln, INFINITE_LOOP_AST_NODE);
}

function_ast_node *parse_function_ast_node(parser *parser) {
	match_token_type(parser, IDENTIFIER);	// consume the fn keyword

	token *functionName = match_token_type(parser, IDENTIFIER); // name of function
	vector *args = create_vector();

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

			bool is_constant = false;
			if (check_token_type_and_content(parser, IDENTIFIER, CONSTANT_KEYWORD, 0)) {
				is_constant = true;
				consume_token(parser);
			}

			// data type
			token *argdata_type = match_token_type(parser, IDENTIFIER);
			data_type arg_raw_data_type = match_token_type_to_data_type(parser, argdata_type);

			// look for ^
			bool is_pointer = false;
			if (check_token_type_and_content(parser, OPERATOR, POINTER_OPERATOR, 0)) {
				is_pointer = true;
				consume_token(parser);
			}

			// name of argument
			token *arg_name = match_token_type(parser, IDENTIFIER);

			function_argument_ast_node *arg = create_function_argument_ast_node();
			arg->type = arg_raw_data_type;
			arg->name = arg_name;
			arg->is_pointer = is_pointer;
			arg->is_constant = is_constant;
			arg->value = NULL;

			if (check_token_type_and_content(parser, OPERATOR, ASSIGNMENT_OPERATOR, 0)) {
				consume_token(parser);

				// default expression
				expression_ast_node *expr = parse_expression_ast_node(parser);
				arg->value = expr;
				push_back_item(args, arg);

				if (check_token_type_and_content(parser, SEPARATOR, COMMA_SEPARATOR, 0)) {
					consume_token(parser);
				}
				else if (check_token_type_and_content(parser, SEPARATOR, ")", 0)) {
					consume_token(parser); // eat closing parenthesis
					break;
				}
			}
			else if (check_token_type_and_content(parser, SEPARATOR, COMMA_SEPARATOR, 0)) {
				if (check_token_type_and_content(parser, SEPARATOR, ")", 1)) {
					parser_error(parser, "trailing comma at the end of argument list", consume_token(parser), false);
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

		if (check_token_type_and_content(parser, OPERATOR, ":", 0)) {
			consume_token(parser);

			bool is_constant = false;
			if (check_token_type_and_content(parser, IDENTIFIER, CONSTANT_KEYWORD, 0)) {
				is_constant = true;
				consume_token(parser);
			}

			bool returns_pointer = false;
			if (check_token_type_and_content(parser, OPERATOR, POINTER_OPERATOR, 0)) {
				returns_pointer = true;
				consume_token(parser);
			}

			// returns data type
			// todo: let this return a struct... etc
			if (check_token_type(parser, IDENTIFIER, 0)) {
				token *return_type = consume_token(parser);
				data_type raw_data_type = match_token_type_to_data_type(parser, return_type);
				fpn->ret = raw_data_type;
				fn->returns_pointer = returns_pointer;
				fn->is_constant = is_constant;
			}
			else {
				parser_error(parser, "function declaration return type expected", consume_token(parser), false);
			}
		}
		else if (check_token_type(parser, IDENTIFIER, 0)) {
			// if they do for example
			//              V forgot the :!!!!
			// fn whatever() int {
			// }
			parser_error(parser, "found an identifier after function argument list, perhaps you missed a colon?", consume_token(parser), false);
		}
		else {
			data_type void_type = TYPE_VOID;
			fpn->ret = void_type;
		}

		// set function prototype
		fn->fpn = fpn;
		fn->body = parse_block_ast_node(parser);
		
		prepare_ast_node(parser, fn, FUNCTION_AST_NODE);

		return fn;
	}
	else {
		parser_error(parser, "expecting a parameter list", consume_token(parser), false);
	}

	// just in case we fail to parse, free this shit
	free(fpn);
	fpn = NULL;

	parser_error(parser, "failed to parse function", consume_token(parser), true);
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
					parser_error(parser, "trailing comma at the end of argument list", consume_token(parser), false);
				}
				consume_token(parser);
				push_back_item(args, arg);
			}
			else if (check_token_type_and_content(parser, SEPARATOR, ")", 0)) {
				consume_token(parser); // eat closing parenthesis
				push_back_item(args, arg);
				break;
			}
		}
		while (true);

		// woo we got the function
		function_callee_ast_node *fcn = create_function_callee_ast_node();
		fcn->callee = callee->content;
		fcn->args = args;
		prepare_ast_node(parser, fcn, FUNCTION_CALLEE_AST_NODE);

		return fcn;
	}

	parser_error(parser, "failed to parse function call", consume_token(parser), true);
	return NULL;
}

function_return_ast_node *parse_return_statement_ast_node(parser *parser) {
	// consume the return keyword
	match_token_type_and_content(parser, IDENTIFIER, RETURN_KEYWORD);

	// return value
	function_return_ast_node *frn = create_function_return_ast_node();
	frn->return_val = parse_expression_ast_node(parser);

	return frn;

	parser_error(parser, "failed to parse return statement", consume_token(parser), true);
	return NULL;
}

statement_ast_node *parse_statement_ast_node(parser *parser) {
	// RETURN STATEMENTS
	if (check_token_type_and_content(parser, IDENTIFIER, RETURN_KEYWORD, 0)) {
		return create_statement_ast_node(parse_return_statement_ast_node(parser), FUNCTION_RET_AST_NODE);
	}
	// STRUCTURES
	else if (check_token_type_and_content(parser, IDENTIFIER, STRUCT_KEYWORD, 0)) {
		return create_statement_ast_node(parse_structure_ast_node(parser), STRUCT_AST_NODE);
	}
	// IF STATEMENTS
	else if (check_token_type_and_content(parser, IDENTIFIER, IF_KEYWORD, 0)) {
		return create_statement_ast_node(parse_if_statement_ast_node(parser), IF_STATEMENT_AST_NODE);
	}
	// MATCH STATEMENTS
	else if (check_token_type_and_content(parser, IDENTIFIER, MATCH_KEYWORD, 0)) {
		return create_statement_ast_node(parse_match_ast_node(parser), MATCH_STATEMENT_AST_NODE);
	}
	// WHILE LOOPS
	else if (check_token_type_and_content(parser, IDENTIFIER, WHILE_LOOP_KEYWORD, 0)) {
		return create_statement_ast_node(parse_while_loop(parser), WHILE_LOOP_AST_NODE);
	}
	// FOR LOOPS
	else if (check_token_type_and_content(parser, IDENTIFIER, FOR_LOOP_KEYWORD, 0)) {
		return parse_for_loop_ast_node(parser);
	}
	// INFINITE LOOPS
	else if (check_token_type_and_content(parser, IDENTIFIER, INFINITE_LOOP_KEYWORD, 0)) {
		return parse_infinite_loop_ast_node(parser);
	}
	// ENUMERATION
	else if (check_token_type_and_content(parser, IDENTIFIER, ENUM_KEYWORD, 0)) {
		return create_statement_ast_node(parse_enumeration_ast_node(parser), ENUM_AST_NODE);
	}
	// BREAK STATEMENTS
	else if (check_token_type_and_content(parser, IDENTIFIER, BREAK_KEYWORD, 0)) {
		consume_token(parser);
		return create_statement_ast_node(create_break_ast_node(), BREAK_AST_NODE);
	}
	// CONTINUE STATEMENTS
	else if (check_token_type_and_content(parser, IDENTIFIER, CONTINUE_KEYWORD, 0)) {
		consume_token(parser);
		return create_statement_ast_node(create_continue_ast_node(), CONTINUE_AST_NODE);
	}
	// IDENTIFERS
	else if (check_token_type(parser, IDENTIFIER, 0)) {
		token *look_ahead = peek_at_token_stream(parser, 0);
		
		// VARIABLE REASSIGNMENT
		if (check_token_type_and_content(parser, OPERATOR, ASSIGNMENT_OPERATOR, 1)) {
			return create_statement_ast_node(parse_reassignment_statement_ast_node(parser), VARIABLE_REASSIGN_AST_NODE);
		}
		// FUNCITON CALL
		else if (check_token_type_and_content(parser, SEPARATOR, "(", 1)) {
			return create_statement_ast_node(parse_function_callee_ast_node(parser), FUNCTION_CALLEE_AST_NODE);
		}
		// LOCAL VARIABLE
		else if (check_token_type_is_valid_data_type(parser, look_ahead)) {
			return parse_variable_ast_node(parser, false);
		}
		// ERROR!!
		else {
			parser_error(parser, "unrecognized identifier", consume_token(parser), false);
		}
	}

	parser_error(parser, "unrecognized token specified", consume_token(parser), true);
	return NULL;
}

variable_reassignment_ast_node *parse_reassignment_statement_ast_node(parser *parser) {
	if (check_token_type(parser, IDENTIFIER, 0)) {
		token *variableName = consume_token(parser);

		if (check_token_type_and_content(parser, OPERATOR, "=", 0)) {
			consume_token(parser);

			expression_ast_node *expr = parse_expression_ast_node(parser);

			variable_reassignment_ast_node *vrn = create_variable_reassign_ast_node();
			vrn->name = variableName;
			vrn->expr = expr;
			return vrn;
		}
	}

	parser_error(parser, "failed to parse variable reassignment", consume_token(parser), true);
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
				else if (check_token_type_is_valid_data_type(parser, tok)
					|| check_token_type_and_content(parser, IDENTIFIER, CONSTANT_KEYWORD, 0)) {
					parse_variable_ast_node(parser, true);
				}
				else if (check_token_type_and_content(parser, OPERATOR, "=", 1)) {
					parse_reassignment_statement_ast_node(parser);
				}
				else if (check_token_type_and_content(parser, SEPARATOR, "(", 1)) {
					prepare_ast_node(parser, parse_function_callee_ast_node(parser), FUNCTION_CALLEE_AST_NODE);
				}
				else {
					parser_error(parser, "unrecognized identifier specified", consume_token(parser), false);
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

	parser_error(parser, "unrecognized data-type specified", consume_token(parser), true);
	return 0;
}

void prepare_ast_node(parser *parser, void *data, ast_node_type type) {
	ast_node *ast_node = safe_malloc(sizeof(*ast_node));
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
	if (parser) {
		int i;
		for (i = 0; i < parser->parse_tree->size; i++) {
			ast_node *ast_node = get_vector_item(parser->parse_tree, i);
			remove_ast_node(ast_node);
		}
		destroy_vector(parser->parse_tree);
		destroy_hashmap(parser->sym_table);

		free(parser);
	}
}
