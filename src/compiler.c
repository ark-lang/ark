#include "compiler.h"

static const char* NODE_TYPE[] = {
	"EXPRESSION_AST_NODE", "VARIABLE_DEF_AST_NODE",
	"VARIABLE_DEC_AST_NODE", "FUNCTION_ARG_AST_NODE",
	"FUNCTION_AST_NODE", "FUNCTION_PROT_AST_NODE",
	"BLOCK_AST_NODE", "FUNCTION_CALLEE_AST_NODE",
	"FUNCTION_RET_AST_NODE", "FOR_LOOP_AST_NODE",
	"VARIABLE_REASSIGN_AST_NODE", "INFINITE_LOOP_AST_NODE",
	"BREAK_AST_NODE", "DO_WHILE_AST_NODE", "CONTINUE_AST_NODE", "ENUM_AST_NODE", "STRUCT_AST_NODE",
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
		case FOR_LOOP_AST_NODE:
			emitForLoop(self, currentAstNode->data);
			break;
		case INFINITE_LOOP_AST_NODE:
			emitInfiniteLoop(self, currentAstNode->data);
			break;
		case BREAK_AST_NODE:
			// why even have it's own function?
			emitCode(self, "break;");
			break;
		case ENUM_AST_NODE:
			self->writeState = WRITE_SOURCE_STATE;
			emitEnumeration(self, currentAstNode->data);
			break;
		case WHILE_LOOP_AST_NODE:
			emitWhileLoop(self, currentAstNode->data);
			break;
		case DO_WHILE_AST_NODE:
			emitDoWhileLoop(self, currentAstNode->data);
			break;
		default:
			printf("wat node is that bby?\n");
			break;
		}
	}
}

void emitVariableDefine(Compiler *self, VariableDefinitionAstNode *def) {
	char *exists = NULL;
	bool structDef = !hashmap_get(self->structures, def->type->content, (void**) &exists);

	emitCode(self, "%s ", def->type->content);
	if (def->isPointer) {
		emitCode(self, "*");
	}

	// it's not a structure definition, set it to 0
	if (!structDef || (structDef && !def->isPointer)) {
		emitCode(self, "%s = 0;\n", def->name);
	}
	else {
		emitCode(self, "%s = malloc(sizeof(*%s));\n", def->name, def->name);
	}
}

void emitVariableDeclaration(Compiler *self, VariableDeclarationAstNode *var) {
	if (!var->isMutable) {
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

void emitStructureTypeDeclare(Compiler *self, VariableDeclarationAstNode *dec) {
	emitCode(self, "%s ", dec->variableDefinitionAstNode->type->content);
	if (dec->variableDefinitionAstNode->isPointer) {
		emitCode(self, "*");
	}
	emitCode(self, "%s;\n", dec->variableDefinitionAstNode->name);
}

void emitStructure(Compiler *self, StructureAstNode *structure) {
	hashmap_put(self->structures, structure->name, structure->name);

	self->writeState = WRITE_HEADER_STATE;

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
	sds randName = randString(12);

	if (call->isFunctionRedirect) {
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

	if (call->isFunctionRedirect) {
		int i;
		for (i = 0; i < call->vars->size; i++) {
			Token *tok = getVectorItem(call->vars, i);
			emitCode(self, "%s = %s;\n", tok->content, randName);
		}
	}

	sdsfree(randName);
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

void emitForLoop(Compiler *self, ForLoopAstNode *forLoop) {
	self->writeState = WRITE_SOURCE_STATE;

	sds indexRandName = NULL;
	if (!strcmp(forLoop->indexName->content, "_")) {
		indexRandName = randString(12);
	}
	else {
		indexRandName = sdsnew(forLoop->indexName->content);
	}

	sds stepValue = NULL;
	if (getVectorItem(forLoop->parameters, FOR_STEP) != NULL) {
		stepValue = sdsnew(((Token *) getVectorItem(forLoop->parameters, FOR_STEP))->content);
	}
	else {
		stepValue = sdsnew("1"); // default step of 1.
	}

	// TODO: check if inclusive/exclusive, and handle cases where the operator is <=, >=, >, ==, != etc
	emitCode(self, "for (%s %s = %s; %s < %s; %s += %s) {\n",
			forLoop->type->content,
			indexRandName,
			((Token*) getVectorItem(forLoop->parameters, FOR_START))->content,
			indexRandName,
			((Token*) getVectorItem(forLoop->parameters, FOR_END))->content,
			indexRandName,
			stepValue);

	emitBlock(self, forLoop->body);
	emitCode(self, "}\n");
}

void emitInfiniteLoop(Compiler *self, InfiniteLoopAstNode *infinite) {
	self->writeState = WRITE_SOURCE_STATE;
	emitCode(self, "for(;;) {\n");
	emitBlock(self, infinite->body);
	emitCode(self, "}\n");
}

void emitWhileLoop(Compiler *self, WhileLoopAstNode *whileLoop) {
	self->writeState = WRITE_SOURCE_STATE;
	emitCode(self, "while (");
	emitExpression(self, whileLoop->condition);
	emitCode(self, ") {\n");
	emitBlock(self, whileLoop->body);
	emitCode(self, "}\n");
}

void emitDoWhileLoop(Compiler *self, DoWhileAstNode *doWhile) {
	self->writeState = WRITE_SOURCE_STATE;
	emitCode(self, "do {\n");
	emitBlock(self, doWhile->body);
	emitCode(self, "} while (");
	emitExpression(self, doWhile->condition);
	emitCode(self, ");\n");
}

void emitEnumeration(Compiler *self, EnumAstNode *enumeration) {
	emitCode(self, "typedef enum {\n");
	int i;
	for (i = 0; i < enumeration->enumItems->size; i++) {
		EnumItem *enumItem = getVectorItem(enumeration->enumItems, i);
		emitCode(self, "%s", enumItem->name);
		if (enumItem->hasValue) {
			emitCode(self, " = %d", enumItem->value);
		}
		emitCode(self, ",\n");
	}
	emitCode(self, "} %s;\n", enumeration->name->content);
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

		if (!arg->isMutable) {
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

		if (!arg->isMutable) {
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
		// __gen_name.h is the typical name for the headers and c files that are generated
		emitCode(self, "#include \"__gen_%s.h\"\n", self->currentSourceFile->name);

		// write to header
		self->writeState = WRITE_HEADER_STATE;
		sds nameInUpperCase = toUppercase(self->currentSourceFile->name);

		emitCode(self, "#ifndef __%s_H\n", nameInUpperCase);
		emitCode(self, "#define __%s_H\n\n", nameInUpperCase);

		// compile code
		compileAST(self);

		// write to header
		self->writeState = WRITE_HEADER_STATE;
		emitCode(self, "\n");
		emitCode(self, "#endif // __%s_H\n", nameInUpperCase);

		sdsfree(nameInUpperCase);

		// close files
		closeFiles(self->currentSourceFile);
	}

	sds buildCommand = sdsempty();

	// append the compiler to use etc
	sdscat(buildCommand, COMPILER);
	sdscat(buildCommand, " -std=c99 -Wall -o ");
	sdscat(buildCommand, OUTPUT_EXECUTABLE_NAME);
	sdscat(buildCommand, " ");

	// append the filename to the build string
	for (i = 0; i < self->sourceFiles->size; i++) {
		SourceFile *sourceFile = getVectorItem(self->sourceFiles, i);
		sdscat(buildCommand, "__gen_");
		sdscat(buildCommand, sourceFile->name);
		sdscat(buildCommand, ".c");

		if (i != self->sourceFiles->size - 1) // stop whitespace at the end!
			sdscat(buildCommand, " ");
	}

	// just for debug purposes
	debugMessage("running cl args: `%s`\n", buildCommand);
	system(buildCommand);
	sdsfree(buildCommand); // deallocate dat shit baby

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
			case ENUM_AST_NODE:
				self->writeState = WRITE_HEADER_STATE;
				emitEnumeration(self, currentAstNode->data);
				break;
			default:
				printf("node not yet supported: %s :(\n", NODE_TYPE[currentAstNode->type]);
				break;
		}
	}
}

void destroyCompiler(Compiler *self) {
	if (self) {
		hashmap_free(self->functions);
		hashmap_free(self->structures);
		free(self);
		debugMessage("Destroyed compiler");
	}
}
