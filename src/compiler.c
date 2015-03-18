#include "compiler.h"

static const char* NODE_TYPE[] = {
	"EXPRESSION_AST_NODE", "VARIABLE_DEF_AST_NODE",
	"VARIABLE_DEC_AST_NODE", "FUNCTION_ARG_AST_NODE",
	"FUNCTION_AST_NODE", "FUNCTION_PROT_AST_NODE",
	"BLOCK_AST_NODE", "FUNCTION_CALLEE_AST_NODE",
	"FUNCTION_RET_AST_NODE", "FOR_LOOP_AST_NODE",
	"VARIABLE_REASSIGN_AST_NODE", "INFINITE_LOOP_AST_NODE",
	"BREAK_AST_NODE", "CONTINUE_AST_NODE", "ENUM_AST_NODE", "STRUCT_AST_NODE",
	"IF_STATEMENT_AST_NODE", "MATCH_STATEMENT_AST_NODE", "WHILE_LOOP_AST_NODE",
	"ANON_AST_NODE", "USE_STATEMENT_AST_NODE"
};

Compiler *createCompiler(Vector *sourceFiles) {
	Compiler *self = safeMalloc(sizeof(*self));
	self->abstractSyntaxTree = NULL;
	self->currentNode = 0;
	self->sourceFiles = sourceFiles;
	return self;
}

void emitCode(Compiler *self, char *fmt, ...) {
	va_list args;
	va_start(args, fmt);

	switch (self->writeState) {
		case WRITE_SOURCE_STATE:
			vfprintf(self->currentSourceFile->outputFile, fmt, args);
			va_end(args);
			break;
		case WRITE_HEADER_STATE:
			vfprintf(self->currentSourceFile->headerFile->outputFile, fmt, args);
			va_end(args);
			break;
	}
}

void emitBlock(Compiler *self, BlockAstNode *block) {
	int i;
	for (i = 0; i < block->statements->size; i++) {
		StatementAstNode *currentAstNode = getVectorItem(block->statements, i);
		switch (currentAstNode->type) {
		case VARIABLE_DEC_AST_NODE:
			emitVariableDeclaration(self, currentAstNode->data);
			break;
		case FUNCTION_CALLEE_AST_NODE:
			emitFunctionCall(self, currentAstNode->data);
			break;
		case IF_STATEMENT_AST_NODE:
			emitIfStatement(self, currentAstNode->data);
			break;
		case FUNCTION_RET_AST_NODE:
			emitReturnStatement(self, currentAstNode->data);
			break;
		default:
			printf("wat node is that bby?\n");
			break;
		}
	}
}

void emitVariableDeclaration(Compiler *self, VariableDeclarationAstNode *var) {
	if (var->variableDefinitionAstNode->isConstant) {
		emitCode(self, "const ");
	}
	if (var->variableDefinitionAstNode->isPointer) {
		emitCode(self, "*");
	}
	emitCode(self, "%s %s = ", var->variableDefinitionAstNode->type->content, var->variableDefinitionAstNode->name);
	self->writeState = WRITE_SOURCE_STATE;
	emitExpression(self, var->expression);
	emitCode(self, ";\n");
}

void emitStructure(Compiler *self, StructureAstNode *structure) {

}

void emitExpression(Compiler *self, ExpressionAstNode *expr) {
	int i;
	for (i = 0; i < expr->tokens->size; i++) {
		Token *token = getVectorItem(expr->tokens, i);
		emitCode(self, "%s", token->content);
	}
}

void emitFunctionCall(Compiler *self, FunctionCallAstNode *call) {	
	self->writeState = WRITE_SOURCE_STATE;
	emitCode(self, "%s(", call->name);
	int i;
	for (i = 0; i < call->args->size; i++) {
		FunctionArgumentAstNode *arg = getVectorItem(call->args, i);

		self->writeState = WRITE_SOURCE_STATE;
		emitExpression(self, arg->value);
		if (call->args->size > 1 && i != call->args->size - 1) {
			emitCode(self, ", ");
		}
	}
	emitCode(self, ");\n");
}

void emitIfStatement(Compiler *self, IfStatementAstNode *stmt) {
	self->writeState = WRITE_SOURCE_STATE;
	emitCode(self, "if (");
	emitExpression(self, stmt->condition);
	emitCode(self, ") {\n");
	emitBlock(self, stmt->body);
	emitCode(self, "}\n");
	if (stmt->elseStatement) {
		emitCode(self, "else {\n");
		emitBlock(self, stmt->elseStatement);
		emitCode(self, "}\n");
	}
}

void emitReturnStatement(Compiler *self, FunctionReturnAstNode *ret) {
	self->writeState = WRITE_SOURCE_STATE;
	emitCode(self, "return");
	if (ret->returnValue) {
		emitCode(self, " ");
		self->writeState = WRITE_SOURCE_STATE;
		emitExpression(self, ret->returnValue);
	}
	emitCode(self, ";\n");
}

void emitUseStatement(Compiler *self, UseStatementAstNode *use) {
	self->writeState = WRITE_SOURCE_STATE;
	emitCode(self, "#include %s\n", use->file->content);
}

void emitFunction(Compiler *self, FunctionAstNode *func) {

	// prototype
	self->writeState = WRITE_HEADER_STATE;
	emitCode(self, "%s ", func->prototype->returnType->content);
	emitCode(self, "%s(", func->prototype->name->content);

	int i;
	for (i = 0; i < func->prototype->args->size; i++) {
		FunctionArgumentAstNode *arg = getVectorItem(func->prototype->args, i);

		if (arg->isConstant) {
			emitCode(self, "const ");
		}
		if (arg->isPointer) {
			emitCode(self, "*");
		}

		emitCode(self, "%s %s", arg->type->content, arg->name->content);

		if (func->prototype->args->size > 1 && i != func->prototype->args->size - 1) {
			emitCode(self, ", ");
		}
	}

	emitCode(self, ");\n");

	// WRITE TO THE SOURCE FILE

	self->writeState = WRITE_SOURCE_STATE;
	emitCode(self, "%s ", func->prototype->returnType->content);
	emitCode(self, "%s(", func->prototype->name->content);

	for (i = 0; i < func->prototype->args->size; i++) {
		FunctionArgumentAstNode *arg = getVectorItem(func->prototype->args, i);

		if (arg->isConstant) {
			emitCode(self, "const ");
		}
		if (arg->isPointer) {
			emitCode(self, "*");
		}

		emitCode(self, "%s %s", arg->type->content, arg->name->content);

		if (func->prototype->args->size > 1 && i != func->prototype->args->size - 1) {
			emitCode(self, ", ");
		}
	}

	emitCode(self, ") {\n");
	emitBlock(self, func->body);
	emitCode(self, "}\n");
}

void consumeAstNode(Compiler *self) {
	self->currentNode += 1;
}

void consumeAstNodeBy(Compiler *self, int amount) {
	self->currentNode += amount;
}

void startCompiler(Compiler *self) {
	size_t files_len = 0;

	int i;
	for (i = 0; i < self->sourceFiles->size; i++) {
		SourceFile *sourceFile = getVectorItem(self->sourceFiles, i);
		files_len += strlen(sourceFile->name) + 1;
		self->currentNode = 0;
		self->currentSourceFile = sourceFile;
		self->abstractSyntaxTree = self->currentSourceFile->ast;

		writeFiles(self->currentSourceFile);

		self->writeState = WRITE_SOURCE_STATE;
		emitCode(self, "#include \"%s.h\"\n", self->currentSourceFile->name);

		// write to header
		self->writeState = WRITE_HEADER_STATE;
		char *upper_name = toUppercase(self->currentSourceFile->name);

		emitCode(self, "#ifndef __%s_H\n", upper_name);
		emitCode(self, "#define __%s_H\n\n", upper_name);

		// compile code
		compileAST(self);

		// write to header
		self->writeState = WRITE_HEADER_STATE;
		emitCode(self, "\n");
		emitCode(self, "#endif // __%s_H\n", upper_name);

		// close files
		closeFiles(self->currentSourceFile);
	}

	// 1 for the null terminator
	files_len += 1;
	files_len += 4; // 4 for gcc and a space, bit hacky

	char *files = malloc(sizeof(char) * files_len);
	files[0] = '\0';
	strcat(files, "gcc ");
	for (i = 0; i < self->sourceFiles->size; i++) {
		SourceFile *sourceFile = getVectorItem(self->sourceFiles, i);
		strcat(files, sourceFile->name);
		strcat(files, ".c");
		strcat(files, " ");
	}
	files[files_len] = '\0';
	system(files);

	for (i = 0; i < self->sourceFiles->size; i++) {
		SourceFile *sourceFile = getVectorItem(self->sourceFiles, i);
		destroySourceFile(sourceFile);
		destroyHeaderFile(sourceFile->headerFile);
	}
}

void compileAST(Compiler *self) {
	int i;
	for (i = 0; i < self->abstractSyntaxTree->size; i++) {
		AstNode *currentAstNode = getVectorItem(self->abstractSyntaxTree, i);

		switch (currentAstNode->type) {
			case FUNCTION_AST_NODE:
				emitFunction(self, currentAstNode->data);
				break;
			case VARIABLE_DEC_AST_NODE:
				emitVariableDeclaration(self, currentAstNode->data);
				break;
			case STRUCT_AST_NODE:
				emitStructure(self, currentAstNode->data);
				break;
			case USE_STATEMENT_AST_NODE:
				emitUseStatement(self, currentAstNode->data);
				break;
			default:
				printf("node not yet supported: %s :(\n", NODE_TYPE[currentAstNode->type]);
				break;
		}
	}
}

void destroyCompiler(Compiler *self) {
	free(self);
}
