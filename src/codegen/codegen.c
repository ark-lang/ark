#include "codegen.h"

/**
 * A boilerplate for our generated C files. Eventually
 * we should only write this into one file included into
 * all the headers instead of generate on top of all the header
 * files?
 *
 * It includes a few vital header files that will eventually
 * be written in Alloy itself. Also a few aliases for certain
 * types.
 */
const char *BOILERPLATE =
"#include <stdbool.h>\n"
"#include <stddef.h>\n"
"#include <stdarg.h>\n"
"\n"
"typedef char *str;" CC_NEWLINE
"typedef size_t usize;" CC_NEWLINE
"typedef unsigned char u8;" CC_NEWLINE
"typedef unsigned short u16;" CC_NEWLINE
"typedef unsigned int u32;" CC_NEWLINE
"typedef unsigned long long u64;" CC_NEWLINE
"typedef char i8;" CC_NEWLINE
"typedef short i16;" CC_NEWLINE
"typedef int i32;" CC_NEWLINE
"typedef long long i64;" CC_NEWLINE
"typedef float f32;" CC_NEWLINE
"typedef double f64;" CC_NEWLINE
;

/**
 * Node names corresponding to their enumerated counterpart
 */
const char *NODE_NAME[] = {
	"IDENTIFIER_LIST_NODE", "IDENTIFIER_NODE", "LITERAL_NODE", "BINARY_EXPR_NODE",
	"UNARY_EXPR_NODE", "ARRAY_SUB_EXPR_NODE",
	"PRIMARY_EXPR_NODE", "EXPR_NODE", "TYPE_NAME_NODE", "TYPE_LIT_NODE", "PAREN_EXPR_NODE",
	"ARRAY_TYPE_NODE", "POINTER_TYPE_NODE", "FIELD_DECL_NODE",
	"FIELD_DECL_LIST_NODE", "STRUCT_DECL_NODE", "STATEMENT_LIST_NODE",
	"BLOCK_NODE", "PARAMETER_SECTION_NODE", "PARAMETERS_NODE", "IMPL_NODE",
	"FUNCTION_SIGNATURE_NODE", "FUNCTION_DECL_NODE", "VARIABLE_DECL_NODE", "FUNCTION_CALL_NODE",
	"DECLARATION_NODE", "INC_DEC_STAT_NODE", "RETURN_STAT_NODE", "BREAK_STAT_NODE",
	"CONTINUE_STAT_NODE", "LEAVE_STAT_NODE", "ASSIGNMENT_NODE", "UNSTRUCTURED_STATEMENT_NODE",
	"ELSE_STAT_NODE", "IF_STAT_NODE", "MATCH_CLAUSE_STAT", "MATCH_STAT_NODE", "FOR_STAT_NODE",
	"STRUCTURED_STATEMENT_NODE", "STATEMENT_NODE", "TYPE_NODE", "POINTER_FREE_NODE"
};

CodeGenerator *createCodeGenerator(Vector *sourceFiles) {
	CodeGenerator *self = safeMalloc(sizeof(*self));
	self->abstractSyntaxTree = NULL;
	self->currentNode = 0;
	self->sourceFiles = sourceFiles;
	self->LINKER_FLAGS = sdsempty();
	return self;
}

void emitCode(CodeGenerator *self, const char *fmt, ...) {
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

void consumeAstNode(CodeGenerator *self) {
	self->currentNode += 1;
}

void consumeAstNodeBy(CodeGenerator *self, int amount) {
	self->currentNode += amount;
}

void emitLiteral(CodeGenerator *self, Literal *lit) {
	if (lit->type == HEX) {
		
	}
	emitCode(self, "%s", lit->value);
}

void emitBinaryExpr(CodeGenerator *self, BinaryExpr *expr) {
	emitExpression(self, expr->lhand);
	emitCode(self, "%s", expr->binaryOp);
	emitExpression(self, expr->rhand);
}

void emitUnaryExpr(CodeGenerator *self, UnaryExpr *expr) {
	if (!strcmp(expr->unaryOp, "^")) {
		// change it to a dereference
		// im pretty sure this shouldn't conflict with
		// XOR?
		emitCode(self, "*");
	}
	else {
		emitCode(self, "%s", expr->unaryOp);
	}
	emitExpression(self, expr->lhand);
}

void emitExpression(CodeGenerator *self, Expression *expr) {
	switch (expr->exprType) {
		case TYPE_NODE: emitType(self, expr->type); break;
		case LITERAL_NODE: emitLiteral(self, expr->lit); break;
		case BINARY_EXPR_NODE: emitBinaryExpr(self, expr->binary); break;
		case UNARY_EXPR_NODE: emitUnaryExpr(self, expr->unary); break;
		case FUNCTION_CALL_NODE: emitFunctionCall(self, expr->call); break;
		default:
			printf("Unknown node in expression %s\n", NODE_NAME[expr->exprType]);
			break;
	}
}

void emitTypeLit(CodeGenerator *self, TypeLit *lit) {
	switch (lit->type) {
		case ARRAY_TYPE_NODE: {
			emitType(self, lit->arrayType->type);
			emitCode(self, "[");
			emitCode(self, "]");
			break;
		}
		case POINTER_TYPE_NODE: {
			emitType(self, lit->pointerType->type);
			emitCode(self, "*");
			break;
		}
	}
}

void emitType(CodeGenerator *self, Type *type) {
	switch (type->type) {
		case TYPE_NAME_NODE: 
			emitCode(self, "%s", type->typeName->name);
			break;
		case TYPE_LIT_NODE:
			emitTypeLit(self, type->typeLit);
			break;
	}
}

void emitParameters(CodeGenerator *self, Parameters *params) {
	for (int i = 0; i < params->paramList->size; i++) {
		ParameterSection *param = getVectorItem(params->paramList, i);
		if (!param->mutable) {
			emitCode(self, "const ");
		}
		emitType(self, param->type);
		emitCode(self, " %s", param->name);

		if (params->paramList->size > 1 && i != params->paramList->size - 1) {
			emitCode(self, ", "); // cleaner formatting
		}
	}
}

void emitFieldList(CodeGenerator *self, FieldDeclList *list) {
	const int size = list->members->size;

	for (int i = 0; i < size; i++) {
		FieldDecl *decl = getVectorItem(list->members, i);
		emitCode(self, "\t");
		if (!decl->mutable) {
			emitCode(self, "const ");
		}
		emitType(self, decl->type);
		emitCode(self, " %s;" CC_NEWLINE, decl->name);
	}
}

void emitStructDecl(CodeGenerator *self, StructDecl *decl) {
	self->writeState = WRITE_HEADER_STATE;
	emitCode(self, "typedef struct {" CC_NEWLINE);
	emitFieldList(self, decl->fields);
	emitCode(self, "} %s;" CC_NEWLINE CC_NEWLINE, decl->name);
}

void emitFunctionCall(CodeGenerator *self, Call *call) {
	for (int i = 0; i < call->callee->size; i++) {
		char *value = getVectorItem(call->callee, i);
		emitCode(self, "%s", value);
	}

	emitCode(self, "(");
	for (int i = 0; i < call->arguments->size; i++) {
		Expression *expr = getVectorItem(call->arguments, i);
		emitExpression(self, expr);
		
		if (call->arguments->size > 1 && i != call->arguments->size - 1) {
			emitCode(self, ", ");
		}
	}
	emitCode(self, ")");
}

void emitFunctionDecl(CodeGenerator *self, FunctionDecl *decl) {
	// prototype in header
	self->writeState = WRITE_HEADER_STATE;
	emitType(self, decl->signature->type);
	emitCode(self, " %s(", decl->signature->name);
	if (decl->signature->owner && decl->signature->ownerArg) {
		// assumes its a pointer, probably not a good idea.
		emitCode(self, "%s *%s", decl->signature->owner, decl->signature->ownerArg);

		// this will emit a comma if there are more than zero paramters.
		if (decl->signature->parameters->paramList->size > 0) {
			emitCode(self, ",");
		}
	}
	emitParameters(self, decl->signature->parameters);
	if (decl->signature->parameters->variadic) {
		emitCode(self, ", ...");
	}
	emitCode(self, ");" CC_NEWLINE);

	// write to the source!

	self->writeState = WRITE_SOURCE_STATE;
	if (decl->body && !decl->prototype) {
		// definition
		emitType(self, decl->signature->type);
		emitCode(self, " %s(", decl->signature->name);
		if (decl->signature->owner && decl->signature->ownerArg) {
			// this should work?
			// probably shouldnt be a pointer
			emitCode(self, "%s *%s", decl->signature->owner, decl->signature->ownerArg);

			// this will emit a comma if there are more than zero paramters.
			if (decl->signature->parameters->paramList->size > 0) {
				emitCode(self, ",");
			}
		}
		emitParameters(self, decl->signature->parameters);
		if (decl->signature->parameters->variadic) {
			emitCode(self, ", ...");
		}
		emitCode(self, ") {" CC_NEWLINE);
		emitBlock(self, decl->body);
		emitCode(self, "}" CC_NEWLINE);
	}
}

void emitVariableDecl(CodeGenerator *self, VariableDecl *decl) {
	if (!decl->mutable) {
		emitCode(self, "const ");
	}
	emitType(self, decl->type);
	if (decl->assigned) {
		emitCode(self, " %s = ", decl->name);
		emitExpression(self, decl->expr);
		emitCode(self, ";" CC_NEWLINE);
	}
	else {
		emitCode(self, " %s;" CC_NEWLINE, decl->name);
	}
}

void emitWhileForLoop(CodeGenerator *self, ForStat *stmt) {
	self->writeState = WRITE_SOURCE_STATE;
	emitCode(self, "while (");
	emitExpression(self, stmt->index);
	emitCode(self, ") {" CC_NEWLINE);
	emitBlock(self, stmt->body);
	emitCode(self, "}" CC_NEWLINE);
}

void emitBlock(CodeGenerator *self, Block *block) {
	if (block->singleStatementBlock) {
		Statement *stmt = getVectorItem(block->stmtList->stmts, 0);
		if (stmt) {
			emitStatement(self, stmt);
		}
	}
	else {
		for (int i = 0; i < block->stmtList->stmts->size; i++) {
			Statement *stmt = getVectorItem(block->stmtList->stmts, i);
			if (stmt) {
				emitStatement(self, stmt);
			}
		}
	}
}

void emitImpl(CodeGenerator *self, Impl *impl) {
	for (int i = 0; i < impl->funcs->size; i++) {
		FunctionDecl *func = getVectorItem(impl->funcs, i);
		emitFunctionDecl(self, func);
	}
}

void emitAssignment(CodeGenerator *self, Assignment *assign) {
	emitCode(self, "%s = ", assign->iden);
	emitExpression(self, assign->expr);
	emitCode(self, ";" CC_NEWLINE);
}

void emitInfiniteForLoop(CodeGenerator *self, ForStat *stmt) {
	self->writeState = WRITE_SOURCE_STATE;
	emitCode(self, "while (true) {" CC_NEWLINE);
	emitBlock(self, stmt->body);
	emitCode(self, "}" CC_NEWLINE);
}

void emitMatchClause(CodeGenerator *self, MatchClause *clause) {
	ALLOY_UNUSED_OBJ(self);
	ALLOY_UNUSED_OBJ(clause);	
}

void emitMatchStat(CodeGenerator *self, MatchStat *match) {
	emitCode(self, "switch (");
	emitExpression(self, match->expr);
	emitCode(self, ") {" CC_NEWLINE);
	for (int i = 0; i < match->clauses->size; i++) {
		emitMatchClause(self, getVectorItem(match->clauses, i));
	}
	emitCode(self, "}" CC_NEWLINE);
}

void emitIndexForLoop(CodeGenerator *self, ForStat *stmt) {
	/**
	 * We assume that the index for the loop
	 * is already defined outside of the for statement,
	 * for example:
	 *
	 * int x = 0;
	 * for x < 10, x + 1 {
	 * 
	 * }
	 */
	

	// FIXME, this should be better?
	char *iden = stmt->index->binary->lhand->type->typeName->name;

	emitCode(self, "for (%s;", iden);

	// emit index
	emitExpression(self, stmt->index);
	emitCode(self, "; ");
	// then step, but we do x + 1, so we need to prefix
	// x = EXPR??
	emitCode(self, "%s = ", iden);
	emitExpression(self, stmt->step);

	emitCode(self, ") {" CC_NEWLINE);
	emitBlock(self, stmt->body);
	emitCode(self, "}" CC_NEWLINE);
}

void emitForStat(CodeGenerator *self, ForStat *stmt) {
	switch (stmt->forType) {
		case WHILE_FOR_LOOP: emitWhileForLoop(self, stmt); break;
		case INFINITE_FOR_LOOP: emitInfiniteForLoop(self, stmt); break;
		case INDEX_FOR_LOOP: emitIndexForLoop(self, stmt); break;
	}
}

void emitEnumDecl(CodeGenerator *self, EnumDecl *enumDecl) {
	emitCode(self, "typedef enum {" CC_NEWLINE);
	for (int i = 0; i < enumDecl->items->size; i++) {
		EnumItem *item = getVectorItem(enumDecl->items, i);
		emitCode(self, "%s", item->name);
		if (item->val) {
			emitCode(self, "=");
			emitExpression(self, item->val);
		}
		emitCode(self, ",");
	}
	emitCode(self, "} %s;" CC_NEWLINE CC_NEWLINE, enumDecl->name);
}

void emitDeclaration(CodeGenerator *self, Declaration *decl) {
	switch (decl->type) {
		case FUNCTION_DECL_NODE: emitFunctionDecl(self, decl->funcDecl); break;
		case STRUCT_DECL_NODE: emitStructDecl(self, decl->structDecl); break;
		case VARIABLE_DECL_NODE: emitVariableDecl(self, decl->varDecl); break;
		case ENUM_DECL_NODE: emitEnumDecl(self, decl->enumDecl); break;
		default:
			printf("unknown node in declaration %s\n", NODE_NAME[decl->type]);
			break;
	}
}

void emitReturnStat(CodeGenerator *self, ReturnStat *ret) {
	emitCode(self, "return ");
	if (ret->expr) {
		emitExpression(self, ret->expr);
	}
	emitCode(self, ";" CC_NEWLINE);
}

void emitLeaveStat(CodeGenerator *self, LeaveStat *leave) {
	switch (leave->type) {
		case RETURN_STAT_NODE: emitReturnStat(self, leave->retStmt); break;
		case CONTINUE_STAT_NODE: emitCode(self, "continue;" CC_NEWLINE); break;
		case BREAK_STAT_NODE: emitCode(self, "break;" CC_NEWLINE); break;
	}
}

void emitUnstructuredStat(CodeGenerator *self, UnstructuredStatement *stmt) {
	switch (stmt->type) {
		case DECLARATION_NODE: emitDeclaration(self, stmt->decl); break;
		case FUNCTION_CALL_NODE: 
			emitFunctionCall(self, stmt->call); 
			emitCode(self, ";" CC_NEWLINE); // pop a semi colon at the end
			break;
		case EXPR_STAT_NODE:
			emitExpression(self, stmt->expr);
			emitCode(self, ";" CC_NEWLINE);
			break;
		case LEAVE_STAT_NODE: emitLeaveStat(self, stmt->leave); break;
		case ASSIGNMENT_NODE: emitAssignment(self, stmt->assignment); break;
		case IMPL_NODE: emitImpl(self, stmt->impl); break;
		case POINTER_FREE_NODE: 
			emitCode(self, "free(%s);" CC_NEWLINE, stmt->pointerFree->name);
			break;
	}
}

void emitIfStat(CodeGenerator *self, IfStat *ifs) {
	emitCode(self, "if (");
	emitExpression(self, ifs->expr);
	emitCode(self, ") {" CC_NEWLINE);
	emitBlock(self, ifs->body);
	emitCode(self, "}" CC_NEWLINE);
}

void emitStatement(CodeGenerator *self, Statement *stmt) {
	switch (stmt->type) {
		case UNSTRUCTURED_STATEMENT_NODE: 
			emitUnstructuredStat(self, stmt->unstructured);
			break;
		case STRUCTURED_STATEMENT_NODE: 
			emitStructuredStat(self, stmt->structured);
			break;
		default:
			printf("unknown node type in block %s\n", NODE_NAME[stmt->type]);
			break;
	}
}

void emitStructuredStat(CodeGenerator *self, StructuredStatement *stmt) {
	switch (stmt->type) {
		case FOR_STAT_NODE: emitForStat(self, stmt->forStmt); break;
		case IF_STAT_NODE: emitIfStat(self, stmt->ifStmt); break;
		default:
			printf("unknown node type found in structured statement %s\n", NODE_NAME[stmt->type]);
			break;
	}
}

void traverseAST(CodeGenerator *self) {
	for (int i = 0; i < self->abstractSyntaxTree->size; i++) {
		Statement *stmt = getVectorItem(self->abstractSyntaxTree, i);

		switch (stmt->type) {
			case UNSTRUCTURED_STATEMENT_NODE: 
				emitUnstructuredStat(self, stmt->unstructured);
				break;
			case STRUCTURED_STATEMENT_NODE: 
				emitStructuredStat(self, stmt->structured);
				break;
		}
	}
}

void startCodeGeneration(CodeGenerator *self) {
	for (int i = 0; i < self->sourceFiles->size; i++) {
		SourceFile *sf = getVectorItem(self->sourceFiles, i);
		self->currentNode = 0;
		self->currentSourceFile = sf;
		self->abstractSyntaxTree = self->currentSourceFile->ast;
		
		writeFiles(self->currentSourceFile);

		self->writeState = WRITE_SOURCE_STATE;
		// _gen_name.h is the typical name for the headers and c files that are generated
		emitCode(self, "#include \"_gen_%s.h\"\n", self->currentSourceFile->name);

		// write to header
		self->writeState = WRITE_HEADER_STATE;
		emitCode(self, "#ifndef __%s_H\n", self->currentSourceFile->name);
		emitCode(self, "#define __%s_H\n\n", self->currentSourceFile->name);

		generateMacros(self);

		emitCode(self, BOILERPLATE);

		// compile code
		traverseAST(self);

		// write to header
		self->writeState = WRITE_HEADER_STATE;
		emitCode(self, "\n");
		emitCode(self, "#endif // __%s_H\n", self->currentSourceFile->name);

		// close files
		closeFiles(self->currentSourceFile);
	}

	/**
	 *
	 * THIS IS MESSY PLS FIX FELIX!
	 * 
	 */

	// empty command
	sds buildCommand = sdsempty();

	// what compiler to use
	buildCommand = sdscat(buildCommand, COMPILER);
	buildCommand = sdscat(buildCommand, " ");
	
	// additional compiler flags, i.e -g, -Wall etc
	buildCommand = sdscat(buildCommand, ADDITIONAL_COMPILER_ARGS);

	// output name
	buildCommand = sdscat(buildCommand, " -o ");
	buildCommand = sdscat(buildCommand, OUTPUT_EXECUTABLE_NAME);

	// files to compile	
	buildCommand = sdscat(buildCommand, " ");
	// append the filename to the build string
	for (int i = 0; i < self->sourceFiles->size; i++) {
		SourceFile *sourceFile = getVectorItem(self->sourceFiles, i);
		buildCommand = sdscat(buildCommand, sourceFile->generatedSourceName);

		if (i != self->sourceFiles->size - 1) // stop whitespace at the end!
			buildCommand = sdscat(buildCommand, " ");
	}

	// linker options
	buildCommand = sdscat(buildCommand, " ");
	buildCommand = sdscat(buildCommand, self->LINKER_FLAGS);

	// just for debug purposes
	verboseModeMessage("running cl args: `%s`", buildCommand);

	// do the command we just created
	system(buildCommand);
	sdsfree(self->LINKER_FLAGS);
	sdsfree(buildCommand); // deallocate dat shit baby
}

void destroyCodeGenerator(CodeGenerator *self) {
	for (int i = 0; i < self->sourceFiles->size; i++) {
		SourceFile *sourceFile = getVectorItem(self->sourceFiles, i);
		destroySourceFile(sourceFile);
		verboseModeMessage("Destroyed source files on %d iteration.", i);
	}

	free(self);
	verboseModeMessage("Destroyed compiler");
}
