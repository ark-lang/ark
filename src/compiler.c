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
	self->functions = hashmap_new();
	self->structures = hashmap_new();
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

void emitReassignment(Compiler *self, VariableReassignmentAstNode *reassign) {
	if (reassign->isPointer) {
		emitCode(self, "*");
	}
	emitCode(self, "%s =", reassign->name->content);
	emitExpression(self, reassign->expression);
	emitCode(self, ";\n");
}

void emitBlock(Compiler *self, BlockAstNode *block) {
	int i;
	for (i = 0; i < block->statements->size; i++) {
		StatementAstNode *currentAstNode = getVectorItem(block->statements, i);
		switch (currentAstNode->type) {
		case VARIABLE_DEC_AST_NODE:
			emitVariableDeclaration(self, currentAstNode->data);
			break;
		case VARIABLE_DEF_AST_NODE:
			emitVariableDefine(self, currentAstNode->data);
			break;
		case FUNCTION_CALLEE_AST_NODE:
			emitFunctionCall(self, currentAstNode->data);
			break;
		case IF_STATEMENT_AST_NODE:
			emitIfStatement(self, currentAstNode->data);
			break;
		case VARIABLE_REASSIGN_AST_NODE:
			emitReassignment(self, currentAstNode->data);
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

void emitVariableDefine(Compiler *self, VariableDefinitionAstNode *def) {
	emitCode(self, "%s ", def->type->content);
	if (def->isPointer) {
		emitCode(self, "*");
	}
	emitCode(self, "%s = 0;\n", def->name);
}

void emitVariableDeclaration(Compiler *self, VariableDeclarationAstNode *var) {
	if (var->isConstant) {
		emitCode(self, "const ");
	}
	emitCode(self, "%s ", var->variableDefinitionAstNode->type->content);
	if (var->variableDefinitionAstNode->isPointer) {
		emitCode(self, "*");
	}
	emitCode(self, "%s = ", var->variableDefinitionAstNode->name);
	if (var->variableDefinitionAstNode->isPointer) {
		emitCode(self, "malloc(sizeof(*%s));\n", var->variableDefinitionAstNode->name);
		emitCode(self, "*%s = ", var->variableDefinitionAstNode->name);
		emitExpression(self, var->expression);
	}
	else {
		emitExpression(self, var->expression);
	}
	emitCode(self, ";\n");
}

void emitStructureTypeDefine(Compiler *self, VariableDefinitionAstNode *def) {
	emitCode(self, "%s ", def->type->content);
	if (def->isPointer) {
		emitCode(self, "*");
	}
	emitCode(self, "%s;\n", def->name);
}

void emitStructureTypeDeclare(Compiler *self, VariableDeclarationAstNode *def) {

}

void emitStructure(Compiler *self, StructureAstNode *structure) {
	self->writeState = WRITE_HEADER_STATE;
	// hashmap_put(self->structures, )

	emitCode(self, "typedef struct {\n");
	int i;
	for (i = 0; i < structure->statements->size; i++) {
		StatementAstNode *stmt = getVectorItem(structure->statements, i);
		switch (stmt->type) {
		case VARIABLE_DEF_AST_NODE:
			emitStructureTypeDefine(self, stmt->data);
			break;
		case VARIABLE_DEC_AST_NODE:
			emitStructureTypeDeclare(self, stmt->data);
			break;
		default:
			printf("type: %s\n", NODE_TYPE[stmt->type]);
			break;
		}
	}
	emitCode(self, "} %s;\n", structure->name);
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
	// this is for the redirection
	char *randName = randString(12);

	if (call->vars != NULL) {
		char *s = NULL;
		hashmap_get(self->functions, call->name, (void**) &s);
		emitCode(self, "%s %s = ", s, randName);
	}

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

	if (call->vars != NULL) {
		int i;
		for (i = 0; i < call->vars->size; i++) {
			Token *tok = getVectorItem(call->vars, i);
			emitCode(self, "%s = %s;\n", tok->content, randName);
		}
	}

	free(randName);
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

	hashmap_put(self->functions, func->prototype->name->content, func->prototype->returnType->content);

	int i;
	for (i = 0; i < func->prototype->args->size; i++) {
		FunctionArgumentAstNode *arg = getVectorItem(func->prototype->args, i);

		if (arg->isConstant) {
			emitCode(self, "const ");
		}

		emitCode(self, "%s ", arg->type->content);
		if (arg->isPointer) {
			emitCode(self, "*");
		}
		emitCode(self, "%s", arg->name->content);

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

		emitCode(self, "%s ", arg->type->content);
		if (arg->isPointer) {
			emitCode(self, "*");
		}
		emitCode(self, "%s", arg->name->content);

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
	int i;
	for (i = 0; i < self->sourceFiles->size; i++) {
		SourceFile *sourceFile = getVectorItem(self->sourceFiles, i);
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

		free(upper_name);

		// close files
		closeFiles(self->currentSourceFile);
	}

	char *buildCommand = malloc(sizeof(char));
	buildCommand[0] = '\0'; // whatever

	// append the compiler to use etc
	appendString(buildCommand, COMPILER);
	appendString(buildCommand, " ");
	appendString(buildCommand, "-o");
	appendString(buildCommand, " ");
	appendString(buildCommand, OUTPUT_EXECUTABLE_NAME);
	appendString(buildCommand, " ");

	// append the filename to the build string
	for (i = 0; i < self->sourceFiles->size; i++) {
		SourceFile *sourceFile = getVectorItem(self->sourceFiles, i);
		appendString(buildCommand, sourceFile->name);
		appendString(buildCommand, ".c");

		if (i != self->sourceFiles->size - 1) // stop whitespace at the end!
			appendString(buildCommand, " ");
	}

	// just for debug purposes
	debugMessage("running cl args: `%s`\n", buildCommand);
	system(buildCommand);
	free(buildCommand); // deallocate dat shit baby

	// now we can destroy stuff
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
			case VARIABLE_DEF_AST_NODE:
				emitVariableDefine(self, currentAstNode->data);
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
	hashmap_free(self->functions);
	hashmap_free(self->structures);
	free(self);
}
