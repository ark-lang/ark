#include "compiler.h"

char *BOILERPLATE =
"#include <stdlib.h>\n"
"#include <stdio.h>\n"
"#include <stdbool.h>\n"
"\n"
"typedef char *str;" CC_NEWLINE
"typedef unsigned long long u64;" CC_NEWLINE
"typedef unsigned int u32;" CC_NEWLINE
"typedef unsigned short u16;" CC_NEWLINE
"typedef unsigned char u8;" CC_NEWLINE
"typedef long long s64;" CC_NEWLINE
"typedef int s32;" CC_NEWLINE
"typedef short s16;" CC_NEWLINE
"typedef char s8;" CC_NEWLINE
"typedef float f32;" CC_NEWLINE
"typedef double f64;" CC_NEWLINE
;

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

void emitExpression(Compiler *self, Expression *expr) {
	emitCode(self, "5"); // lol it'll do for now
}

void emitType(Compiler *self, Type *type) {
	switch (type->type) {
	case POINTER_TYPE_NODE:
		emitCode(self, "*");
		emitType(self, type->pointerType->type);
		break;
	case ARRAY_TYPE_NODE:
		emitType(self, type->arrayType->type);
		emitCode(self, "[");
		emitExpression(self, type->arrayType->length);
		emitCode(self, "]");
		break;
	case TYPE_NAME_NODE:
		emitCode(self, "%s", type->typeName->name);
		break;
	}
}

void emitParameters(Compiler *self, Parameters *params) {
	for (int i = 0; i < params->paramList->size; i++) {
		ParameterSection *param = getVectorItem(params->paramList, i);
		if (!param->mutable) {
			emitCode(self, "const ");
		}
		emitType(self, param->type);
		emitCode(self, " %s", param->name);

		if (params->paramList->size > 1 && i != params->paramList->size - 1) {
			emitCode(self, ", ");
		}
	}
}

void emitReceiver(Compiler *self, Receiver *rec) {
	if (rec != NULL){
		if (!rec->mutable) {
			emitCode(self, "const ");
		}
		emitCode(self, rec->name);
		emitCode(self, " ");
		emitType(self, rec->type);
	}
}

void emitFunctionSignature(Compiler *self, FunctionSignature *func) {
	if (!func->mutable) {
		emitCode(self, "const ");
	}
	emitType(self, func->type);
	emitCode(self, " %s(", func->name);
	emitReceiver(self, func->receiver);
	// cleaner formatting, also avoids trailing comma
	if (func->receiver && func->parameters->paramList->size > 1) { emitCode(self, ", "); }
	emitParameters(self, func->parameters);
	emitCode(self, ")");
}

void emitStructuredStatement(Compiler *self, StructuredStatement *stmt) {
	switch (stmt->type) {
		case BLOCK_NODE: emitBlock(self, stmt->block); break;
		case FOR_STAT_NODE: emitForStat(self, stmt->forStmt); break;
		case IF_STAT_NODE: emitIfStat(self, stmt->ifStmt); break;
		case MATCH_STAT_NODE: emitMatchStat(self, stmt->matchStmt); break;
	}
}

void emitUnstructuredStatement(Compiler *self, UnstructuredStatement *stmt) {
	switch (stmt->type) {
		case DECLARATION_NODE: emitDeclaration(self, stmt->decl); break;
	}
}

void emitBlock(Compiler *self, Block *block) {
	emitCode(self, " {" CC_NEWLINE);
	emitStatementList(self, block->stmtList);
	emitCode(self, "}" CC_NEWLINE);
}

void emitForStat(Compiler *self, ForStat *stmt) {

}

void emitIfStat(Compiler *self, IfStat *stmt) {

}

void emitMatchStat(Compiler *self, MatchStat *stmt) {

}

void emitStatementList(Compiler *self, StatementList *stmtList) {
	for (int i = 0; i < stmtList->stmts->size; i++) {
		Statement *stmt = getVectorItem(stmtList->stmts, i);
		switch (stmt->type) {
			case STRUCTURED_STMT: emitStructuredStatement(self, stmt->structured); break;
			case UNSTRUCTURED_STMT: emitUnstructuredStatement(self, stmt->unstructured); break;
		}
	}
}

void emitFunctionDecl(Compiler *self, FunctionDecl *decl) {
	self->writeState = WRITE_HEADER_STATE;
	emitFunctionSignature(self, decl->signature);
	emitCode(self, ";"); // semi colon at end of prototype

	self->writeState = WRITE_SOURCE_STATE;
	emitFunctionSignature(self, decl->signature);
	emitBlock(self, decl->body);
}

void emitIdentifierList(Compiler *self, IdentifierList *list) {
	for (int i = 0; i < list->values->size; i++) {
		Token *tok = getVectorItem(list->values, i);
		emitCode(self, tok->content);
		if (i != list->values->size - 1 && list->values->size > 1) {
			emitCode(self, ", ");
		}
	}
}

void emitFieldDeclList(Compiler *self, FieldDeclList *list) {
	for (int i = 0; i < list->members->size; i++) {
		FieldDecl *decl = getVectorItem(list->members, i);
		if (!decl->mutable) {
			emitCode(self, "const ");
		}
		emitType(self, decl->type);
		emitIdentifierList(self, decl->idenList);
	}
}

void emitStructDecl(Compiler *self, StructDecl *decl) {
	self->writeState = WRITE_HEADER_STATE;
	emitCode(self, "typedef struct {" CC_NEWLINE);
	emitFieldDeclList(self, decl->fields);
	emitCode(self, "} %s;" CC_NEWLINE, decl->name);
}

void emitDeclaration(Compiler *self, Declaration *decl) {
	switch (decl->declType) {
		case FUNC_DECL: emitFunctionDecl(self, decl->funcDecl); break;
		case STRUCT_DECL: emitStructDecl(self, decl->structDecl); break;
		case VAR_DECL: break;
	}
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
		// _gen_name.h is the typical name for the headers and c files that are generated
		emitCode(self, "#include \"_gen_%s.h\"\n", self->currentSourceFile->name);

		// write to header
		self->writeState = WRITE_HEADER_STATE;
		emitCode(self, "#ifndef __%s_H\n", self->currentSourceFile->name);
		emitCode(self, "#define __%s_H\n\n", self->currentSourceFile->name);

		emitCode(self, BOILERPLATE);

		// compile code
		compileAST(self);

		// write to header
		self->writeState = WRITE_HEADER_STATE;
		emitCode(self, "\n");
		emitCode(self, "#endif // __%s_H\n", self->currentSourceFile->name);

		// close files
		closeFiles(self->currentSourceFile);
	}
	sds buildCommand = sdsempty();

	// append the compiler to use etc
	buildCommand = sdscat(buildCommand, COMPILER);
	buildCommand = sdscat(buildCommand, " ");
	buildCommand = sdscat(buildCommand, ADDITIONAL_COMPILER_ARGS);
	buildCommand = sdscat(buildCommand, " -o ");
	buildCommand = sdscat(buildCommand, OUTPUT_EXECUTABLE_NAME);
	buildCommand = sdscat(buildCommand, " ");

	// append the filename to the build string
	for (i = 0; i < self->sourceFiles->size; i++) {
		SourceFile *sourceFile = getVectorItem(self->sourceFiles, i);
		buildCommand = sdscat(buildCommand, sourceFile->generatedSourceName);

		if (i != self->sourceFiles->size - 1) // stop whitespace at the end!
			buildCommand = sdscat(buildCommand, " ");
	}

	// just for debug purposes
	verboseModeMessage("running cl args: `%s`", buildCommand);
	system(buildCommand);
	sdsfree(buildCommand); // deallocate dat shit baby
}

void compileAST(Compiler *self) {
	int i;
	for (i = 0; i < self->abstractSyntaxTree->size; i++) {
		Statement *currentStmt = getVectorItem(self->abstractSyntaxTree, i);

		switch (currentStmt->type) {
			case UNSTRUCTURED_STMT: emitUnstructuredStatement(self, currentStmt->unstructured); break;
			case STRUCTURED_STMT: emitStructuredStatement(self, currentStmt->structured); break;
			default:
				printf("idk?\n");
				break;
		}
	}
}

void destroyCompiler(Compiler *self) {
	// now we can destroy stuff
	int i;
	for (i = 0; i < self->sourceFiles->size; i++) {
		SourceFile *sourceFile = getVectorItem(self->sourceFiles, i);
		// don't call destroyHeaderFile since it's called in this function!!!!
		destroySourceFile(sourceFile);
		verboseModeMessage("Destroyed source files on %d iteration.", i);
	}

	hashmap_free(self->functions);
	hashmap_free(self->structures);
	free(self);
	verboseModeMessage("Destroyed compiler");
}
