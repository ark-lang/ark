#include "compiler.h"

Compiler *createCompiler() {
	Compiler *self = safeMalloc(sizeof(*self));
	self->abstractSyntaxTree = NULL;
	self->currentNode = 0;

	self->sourceFileSize = 128;
	self->sourcePosition = 0;
	self->sourceName = "test.c";
	self->sourceContents = malloc(sizeof(char) * (self->sourceFileSize + 1));
	self->sourceContents[0] = '\0';

	return self;
}

void appendToSource(Compiler *self, char *str) {
	self->sourceContents = realloc(self->sourceContents, strlen(self->sourceContents) + strlen(str) + 1);
	strcat(self->sourceContents, str);
}

void emitExpression(Compiler *self, ExpressionAstNode *expr) {
	int i;
	for (i = 0; i < expr->expressionValues->size; i++) {
		Token *tok = getVectorItem(expr->expressionValues, i);
		appendToSource(self, tok->content);
		
		if (i != expr->expressionValues->size - 1) {
			appendToSource(self, SPACE_CHAR);
		}
	}
}

void emitVariableDeclaration(Compiler *self, VariableDeclarationAstNode *var) {
	if (var->variableDefinitionAstNode->isConstant) {
		appendToSource(self, CONST_KEYWORD);
		appendToSource(self, SPACE_CHAR);
	}
	appendToSource(self, var->variableDefinitionAstNode->type->content);
	appendToSource(self, SPACE_CHAR);
	if (var->variableDefinitionAstNode->isPointer) {
		appendToSource(self, ASTERISKS);
	}
	appendToSource(self, var->variableDefinitionAstNode->name);
	appendToSource(self, SPACE_CHAR);
	appendToSource(self, EQUAL_SYM);
	appendToSource(self, SPACE_CHAR);
	emitExpression(self, var->expression);
	appendToSource(self, SEMICOLON);
}

void emitFunctionCall(Compiler *self, FunctionCallAstNode *call) {	
	appendToSource(self, call->name);
	appendToSource(self, OPEN_BRACKET);
	
	int i;
	for (i = 0; i < call->args->size; i++) {
		ExpressionAstNode *expr = getVectorItem(call->args, i);
		emitExpression(self, expr);
	}

	appendToSource(self, CLOSE_BRACKET);
	appendToSource(self, SEMICOLON);
}

void emit_if_statement(Compiler *self, IfStatementAstNode *stmt) {
	appendToSource(self, "if");
	appendToSource(self, SPACE_CHAR);

	appendToSource(self, OPEN_BRACKET);
	emitExpression(self, stmt->condition);
	appendToSource(self, CLOSE_BRACKET);

	emitBlock(self, stmt->body);
	if (stmt->elseStatement) {
		appendToSource(self, "else");
		emitBlock(self, stmt->elseStatement);
	}
}

void emitBlock(Compiler *self, BlockAstNode *block) {
	appendToSource(self, SPACE_CHAR);
	appendToSource(self, OPEN_BRACE);
	appendToSource(self, NEWLINE);

	int i;
	for (i = 0; i < block->statements->size; i++) {
		StatementAstNode *current = getVectorItem(block->statements, i);
		appendToSource(self, TAB);
		switch (current->type) {
			case VARIABLE_DEC_AST_NODE:
				emitVariableDeclaration(self, current->data);
				break;
			case FUNCTION_CALLEE_AST_NODE:
				emitFunctionCall(self, current->data);
				break;
			case FUNCTION_RET_AST_NODE:
				emitReturnStatement(self, current->data);
				break;
			case IF_STATEMENT_AST_NODE:
				emit_if_statement(self, current->data);
				break;
			default:
				printf("idk fuk off\n");
				break;
		}
		if (i != block->statements->size - 1) {
			appendToSource(self, NEWLINE);
		}
	}

	appendToSource(self, NEWLINE);
	appendToSource(self, CLOSE_BRACE);
	appendToSource(self, NEWLINE);
}

void emitArguments(Compiler *self, Vector *args) {
	int i;
	for (i = 0; i < args->size; i++) {
		FunctionArgumentAstNode *current = getVectorItem(args, i);

		if (current->isConstant) {
			appendToSource(self, CONST_KEYWORD);
			appendToSource(self, SPACE_CHAR);
		}

		if (current->type) {
			appendToSource(self, current->type->content);
			appendToSource(self, SPACE_CHAR);
		}

		if (current->isPointer) appendToSource(self, ASTERISKS);

		// this could fail if we're doing a function call
		if (current->name) {
			appendToSource(self, current->name->content);
		}

		if (current->value) {
			emitExpression(self, current->value);
		}

		if (args->size > 1 && i != args->size - 1) {
			appendToSource(self, COMMA_SYM);
			appendToSource(self, SPACE_CHAR);
		}
	}
}

void emitReturnStatement(Compiler *self, FunctionReturnAstNode *ret) {
	appendToSource(self, "return");
	if (ret->returnValue) {
		appendToSource(self, SPACE_CHAR);
		emitExpression(self, ret->returnValue);
	}
	appendToSource(self, SEMICOLON);
}

void emitFunction(Compiler *self, FunctionAstNode *func) {
	char *return_type = func->prototype->returnType->content;

	appendToSource(self, return_type);
	appendToSource(self, SPACE_CHAR);
	appendToSource(self, func->prototype->name->content);
	appendToSource(self, OPEN_BRACKET);

	emitArguments(self, func->prototype->args);

	appendToSource(self, CLOSE_BRACKET);
	appendToSource(self, SPACE_CHAR);

	emitBlock(self, func->body);
}

void consumeAstNode(Compiler *self) {
	self->currentNode += 1;
}

void consumeAstNodeBy(Compiler *self, int amount) {
	self->currentNode += amount;
}

void write_file(Compiler *self) {
	FILE *file = fopen("temp.c", "w");
	if (!file) {
		perror("fopen: failed to open file");
		return;
	}

	fprintf(file, "%s", self->sourceContents);
	fclose(file);

	system(COMPILER " temp.c");

	// remove the file since we don't need it
	remove("temp.c");
}

void startCompiler(Compiler *self, Vector *ast) {
	self->abstractSyntaxTree = ast;

	// include for standard library
	// eventually I'll write an inclusion system
	// and some way of calling C functions
	appendToSource(self, "#include <stdio.h>" NEWLINE);

	int i;
	for (i = 0; i < self->abstractSyntaxTree->size; i++) {
		AstNode *current_ast_node = getVectorItem(self->abstractSyntaxTree, i);

		switch (current_ast_node->type) {
			case FUNCTION_AST_NODE: 
				emitFunction(self, current_ast_node->data); 
				break;
			case VARIABLE_DEC_AST_NODE:
				emitVariableDeclaration(self, current_ast_node->data);
				appendToSource(self, NEWLINE);
				break;
			default:
				printf("wat ast node bby %d is %d?\n", current_ast_node->type, i);
				break;
		}
	}

	printf("%s\n", self->sourceContents);
	write_file(self);
}

void destroyCompiler(Compiler *self) {
	// destroy table
	free(self);
}
