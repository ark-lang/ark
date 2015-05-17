#include "C/codegen.h"
#include "C/private.h"

// Declarations

/**
 * This will emit a literal to the source code
 * @param self the code gen instance
 * @param lit  the literal node
 */
static void emitLiteral(CCodeGenerator *self, Literal *lit);

/**
 * Emits a binary expression into our source code
 * @param self the code gen instance
 * @param expr the binary expr node to emit
 */
static void emitBinaryExpr(CCodeGenerator *self, BinaryExpr *expr);

/**
 * Emits a unary expression to the source code
 * @param self the code gen instance
 * @param expr the expression node to emit
 */
static void emitUnaryExpr(CCodeGenerator *self, UnaryExpr *expr);

/**
 * This emits an expression, which could be a function
 * call, unary expr, binary expr, etc
 * @param self the code gen instance
 * @param expr the expression node to emit
 */
static void emitExpression(CCodeGenerator *self, Expression *expr);

/**
 * This emits code to the current file declared
 * by the "WriteState".
 * @param self the code gen instance
 * @param fmt  the string to print
 * @param ...  additional parameters
 */
static void emitCode(CCodeGenerator *self, const char *fmt, ...);

/**
 * Emits a type literal
 * @param self the code gen instance
 * @param lit  the literal node to emit
 */
static void emitTypeLit(CCodeGenerator *self, TypeLit *lit);

/**
 * Emits a function call, note that it does not 
 * emit the semi colon at the end, as this function
 * can be inside of an expression!
 * @param self the code gen instance
 * @param call the call to emit
 */
static void emitFunctionCall(CCodeGenerator *self, Call *call);

/**
 * Emits an array index
 * @param self the code gen instance
 * @param arrayIndex the array index
 */
static void emitArrayIndex(CCodeGenerator *self, ArrayIndex *arrayIndex);

/**
 * Emits a type of any kind
 * @param self the code gen instance
 * @param type the type to emit
 */
static void emitType(CCodeGenerator *self, Type *type);

/**
 * This will emit the function parameters to the
 * source files
 * @param self   the code gen instance
 * @param params the parameters to emit
 */
static void emitParameters(CCodeGenerator *self, Parameters *params);

/**
 * This will emit a field in a field list.
 *
 * @param self the code gen instance
 * @param decl the field decl
 */
static void emitFieldDecl(CCodeGenerator *self, FieldDecl *decl);

/**
 * This will emit a field list, a field list
 * is the declarations inside of a structure, i.e
 * 	int x;
 *  int y;
 * etc
 * 
 * @param self the code gen instance
 * @param list the list to emit
 */
static void emitFieldList(CCodeGenerator *self, FieldDeclList *list);

/**
 * This will emit a structure declaration, eventually
 * we should make this support declaration of values in
 * a structure.
 * @param self the code gen instance
 * @param decl the declaration to emit
 */
static void emitStructDecl(CCodeGenerator *self, StructDecl *decl);

/**
 * This will emit a function signature
 * @parem self the code gen instance
 * @param signature the function signature
 */
static void emitFunctionSignature(CCodeGenerator *self, FunctionSignature *signature);

/**
 * This will emit a function declaration
 * @param self the code gen instance
 * @param decl the function decl node to emit
 */
static void emitFunctionDecl(CCodeGenerator *self, FunctionDecl *decl);

/**
 * This will emit a for loop that has a single expression
 * or an "index", for instance:
 *
 * for true {
 * 
 * }
 * 
 * @param self the code gen instance
 * @param stmt the for loop to emit
 */
static void emitWhileForLoop(CCodeGenerator *self, ForStat *stmt);

/**
 * This will emit a block
 * @param self the code gen instance
 * @param block the block the emit
 */
static void emitBlock(CCodeGenerator *self, Block *block);

/**
 * This will emit a reassignment, i.e
 * x = 10;
 * 
 * @param self   the code gen instance
 * @param assign the assignment node to emit
 */
static void emitAssignment(CCodeGenerator *self, Assignment *assign);

/**
 * This will emit an infinite for loop, i.e a for loop
 * with no conditions
 * 
 * @param self the code gen instance
 * @param stmt the for loop to emit
 */
static void emitInfiniteForLoop(CCodeGenerator *self, ForStat *stmt);

/**
 * This will emit a for loop with two conditions
 * @param self the code gen instance
 * @param stmt for loop to emit code for
 */
static void emitForStat(CCodeGenerator *self, ForStat *stmt);

/**
 * This will emit a variable decl
 * @param self the code gen instance
 * @param decl the variable decl to emit
 */
static void emitVariableDecl(CCodeGenerator *self, VariableDecl *decl);

/**
 * This will emit a declaration top level node
 * @param self the code gen instance
 * @param decl the decl node to emit
 */
static void emitDeclaration(CCodeGenerator *self, Declaration *decl);

/**
 * This will emit a return statement
 * @param self the code gen instance
 * @param ret  the return statement node to emit
 */
static void emitReturnStat(CCodeGenerator *self, ReturnStat *ret);

/**
 * This will emit a leave statement, which is
 * a top level node for Return, Break and Continue
 * @param self  the code gen instance
 * @param leave the leave statement to emit
 */
static void emitLeaveStat(CCodeGenerator *self, LeaveStat *leave);

/**
 * This will emit the top level node for a statement
 * @param self the code gen instance
 * @param stmt the statement to emit
 */
static void emitStatement(CCodeGenerator *self, Statement *stmt);

/**
 * This will emit a top level unstructed node
 * @param self the code gen instance
 * @param stmt the unstructured node to emit
 */
static void emitUnstructuredStat(CCodeGenerator *self, UnstructuredStatement *stmt);

/**
 * This will emit a structured statement, a structured
 * statement is something with a block or some form of structure,
 * for instance a for loop, or an if statement.
 * @param self the code gen instance
 * @param stmt the structured statement to emit
 */
static void emitStructuredStat(CCodeGenerator *self, StructuredStatement *stmt);

/**
 * Consumes a node in the AST that we're parsing
 * @param self [description]
 */
static void consumeAstNode(CCodeGenerator *self);

/**
 * Jumps ahead in the AST we're parsing
 * @param self   the code gen instance
 * @param amount the amount to consume by
 */
static void consumeAstNodeBy(CCodeGenerator *self, int amount);

/**
 * Run through all the nodes in the AST and
 * generate the code for them!
 * @param self the code gen instance
 */
static void traverseAST(CCodeGenerator *self);

static void emitAlloc(CCodeGenerator *self, Alloc *alloc);

// Definitions

/**
 * A boilerplate for our generated C files. Eventually
 * we should only write this into one file included into
 * all the headers instead of generate on top of all the header
 * files?
 *
 * It includes a few vital header files that will eventually
 * be written in the language itself. Also a few aliases for certain
 * types.
 */
const char *BOILERPLATE =
"#ifndef __BOILERPLATE_H\n"
"#define __BOILERPLATE_H\n"
"#include <stdbool.h>\n"
"#include <stddef.h>\n"
"#include <stdarg.h>\n"
"#include <stdint.h>\n"
"#include <stdlib.h>\n"
CC_NEWLINE
"typedef char *str;" CC_NEWLINE
"typedef size_t usize;" CC_NEWLINE
"typedef uint8_t u8;" CC_NEWLINE
"typedef uint16_t u16;" CC_NEWLINE
"typedef uint32_t u32;" CC_NEWLINE
"typedef uint64_t u64;" CC_NEWLINE
"typedef int8_t i8;" CC_NEWLINE
"typedef int16_t i16;" CC_NEWLINE
"typedef int32_t i32;" CC_NEWLINE
"typedef int64_t i64;" CC_NEWLINE
"typedef float f32;" CC_NEWLINE
"typedef double f64;" CC_NEWLINE
"typedef unsigned int uint;\n"
CC_NEWLINE
"#endif\n"
;

CCodeGenerator *createCCodeGenerator(Vector *sourceFiles) {
	CCodeGenerator *self = safeMalloc(sizeof(*self));
	self->abstractSyntaxTree = NULL;
	self->currentNode = 0;
	self->sourceFiles = sourceFiles;
	self->linkerFlags = sdsempty();
	return self;
}

static void consumeAstNode(CCodeGenerator *self) {
	self->currentNode += 1;
}

static void consumeAstNodeBy(CCodeGenerator *self, int amount) {
	self->currentNode += amount;
}

static void emitLiteral(CCodeGenerator *self, Literal *lit) {
	switch (lit->type) {
		case CHAR_LITERAL_NODE:
			emitCode(self, "%d", lit->charLit->value);
			break;
		case INT_LITERAL_NODE:
			emitCode(self, "%d", lit->intLit->value);
			break;
		case FLOAT_LITERAL_NODE:
			emitCode(self, "%lf", lit->floatLit->value);
			break;
		case STRING_LITERAL_NODE:
			emitCode(self, "%s", lit->stringLit->value);
	}
}

static void emitBinaryExpr(CCodeGenerator *self, BinaryExpr *expr) {
	emitCode(self, "(");
	// for X.Y() calls we need to call the correct function
	if (expr->structType != NULL) {
		Call *call = expr->rhand->call;

		emitCode(self, GENERATED_IMPL_FUNCTION_FORMAT, expr->structType->typeName->name, getVectorItem(call->callee, 0));
		emitCode(self, "(");

		// emit the first "self" argument, by address
		emitCode(self, "&%s", expr->lhand->type->typeName->name);
		if (call->arguments->size > 0) {
			emitCode(self, ", ");
		}

		for (int i = 0; i < call->arguments->size; i++) {
			Expression *expr = getVectorItem(call->arguments, i);
			emitExpression(self, expr);

			if (call->arguments->size > 1 && i != call->arguments->size - 1) {
				emitCode(self, ", ");
			}
		}
		emitCode(self, ")");
	}
	else {
		emitExpression(self, expr->lhand);
		emitCode(self, "%s", expr->binaryOp);
		emitExpression(self, expr->rhand);
	}
	
	emitCode(self, ")");
}

static void emitUnaryExpr(CCodeGenerator *self, UnaryExpr *expr) {
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

static void emitArrayInitializer(CCodeGenerator *self, ArrayInitializer *arr) {
	emitCode(self, "{");
	for (int i = 0; i < arr->values->size; i++) {
		emitExpression(self, getVectorItem(arr->values, i));
		if (arr->values->size > 1 && i != arr->values->size - 1) {
			emitCode(self, ", ");
		}
	}
	emitCode(self, "}");
}

static void emitAlloc(CCodeGenerator *self, Alloc *alloc) {
	if (alloc->type != NULL) {
		emitCode(self, "calloc(1, sizeof(");
		emitType(self, alloc->type);
		emitCode(self, "));" CC_NEWLINE);
	}
	else {
		emitCode(self, "calloc(1, %d);" CC_NEWLINE, alloc->size);
	}
	
	// Emit struct initializers TODO
	/*for (int i = 0; i < decl->structDecl->fields->members->size; ++i) {
		FieldDecl *fieldDecl = getVectorItem(decl->structDecl->fields->members, i);
		if (fieldDecl->defaultValue) {
			emitCode(self, "%s", decl->name);
			emitCode(self, "%s", ".");
			emitCode(self, "%s", fieldDecl->name);
			emitCode(self, " = ");
			emitExpression(self, fieldDecl->defaultValue);
			emitCode(self, ";" CC_NEWLINE);
		}
	}*/
}

static void emitExpression(CCodeGenerator *self, Expression *expr) {
	switch (expr->exprType) {
		case TYPE_NODE: emitType(self, expr->type); break;
		case LITERAL_NODE: emitLiteral(self, expr->lit); break;
		case BINARY_EXPR_NODE: emitBinaryExpr(self, expr->binary); break;
		case UNARY_EXPR_NODE: emitUnaryExpr(self, expr->unary); break;
		case FUNCTION_CALL_NODE: emitFunctionCall(self, expr->call); break;
		case ARRAY_INITIALIZER_NODE: emitArrayInitializer(self, expr->arrayInitializer); break;
		case ARRAY_INDEX_NODE: {
			emitType(self, expr->type);
			emitArrayIndex(self, expr->arrayIndex);
		} break;
		case ALLOC_NODE: emitAlloc(self, expr->alloc); break;
		default:
			errorMessage("Unknown node in expression %d", expr->exprType);
			break;
	}
}

static void emitTypeLit(CCodeGenerator *self, TypeLit *lit) {
	switch (lit->type) {
		case ARRAY_TYPE_NODE: {
			emitType(self, lit->arrayType->type);
			break;
		}
		case POINTER_TYPE_NODE: {
			emitType(self, lit->pointerType->type);
			emitCode(self, "*");
			break;
		}
	}
}

static void emitArrayIndex(CCodeGenerator *self, ArrayIndex *arrayIndex) {
	emitCode(self, "[");
	emitExpression(self, arrayIndex->index);
	emitCode(self, "]");
}

static void emitType(CCodeGenerator *self, Type *type) {
	switch (type->type) {
		case TYPE_NAME_NODE: 
			emitCode(self, "%s", type->typeName->name);
			break;
		case TYPE_LIT_NODE:
			emitTypeLit(self, type->typeLit);
			break;
	}
}

static void emitParameters(CCodeGenerator *self, Parameters *params) {
	for (int i = 0; i < params->paramList->size; i++) {
		ParameterSection *param = getVectorItem(params->paramList, i);
		bool isArray = param->type->type == TYPE_LIT_NODE 
					&& param->type->typeLit->type == ARRAY_TYPE_NODE;

		if (!param->mutable) {
			emitCode(self, "const ");
		}
		emitType(self, param->type);
		emitCode(self, " %s", param->name);
		if (isArray) {
			emitCode(self, "[]");
		}

		if (params->paramList->size > 1 && i != params->paramList->size - 1) {
			emitCode(self, ", "); // cleaner formatting
		}
	}
}

static void emitFieldDecl(CCodeGenerator *self, FieldDecl *decl) {
	// hack
	bool isArray = decl->type->type == TYPE_LIT_NODE
				&& decl->type->typeLit->type == ARRAY_TYPE_NODE;

	emitCode(self, "\t");
	if (!decl->mutable) {
		emitCode(self, "const ");
	}
	emitType(self, decl->type);
	emitCode(self, " %s", decl->name);
	if (isArray) {
		emitCode(self, "[]");
	}
	emitCode(self, ";" CC_NEWLINE);
}

static void emitFieldList(CCodeGenerator *self, FieldDeclList *list) {
	const int size = list->members->size;

	for (int i = 0; i < size; i++) {
		emitFieldDecl(self, getVectorItem(list->members, i));
	}
}

static void emitStructDecl(CCodeGenerator *self, StructDecl *decl) {
	self->writeState = WRITE_HEADER_STATE;
	emitCode(self, "typedef struct {" CC_NEWLINE);
	emitFieldList(self, decl->fields);
	emitCode(self, "} %s;" CC_NEWLINE CC_NEWLINE, decl->name);
}

static void emitFunctionCall(CCodeGenerator *self, Call *call) {
	if (call->typeCast) {
		emitCode(self, "(");
		emitCode(self, "(");
		emitCode(self, "%s", call->typeCast);
		emitCode(self, ") ");
		emitExpression(self, getVectorItem(call->arguments, 0));
		emitCode(self, ")");
	}
	else {
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
}

static void emitFunctionSignature(CCodeGenerator *self, FunctionSignature *signature) {
	emitType(self, signature->type);
	emitCode(self, " ");

	if (signature->owner) {
		emitCode(self, GENERATED_IMPL_FUNCTION_FORMAT, signature->owner, signature->name);
	}
	else {
		emitCode(self, "%s", signature->name);
	}

	emitCode(self, "(");
	if (signature->owner) {
		if (signature->ownerArg) {
			// assumes its a pointer, probably not a good idea.
			emitCode(self, "%s *%s", signature->owner, signature->ownerArg);
		}
		else {
			emitCode(self, "%s *self", signature->owner,signature->ownerArg);
		}

		// this will emit a comma if there are more than zero parameters.
		if (signature->parameters->paramList->size > 0) {
			emitCode(self, ",");
		}
	}
	emitParameters(self, signature->parameters);
	if (signature->parameters->variadic) {
		emitCode(self, ", ...");
	}
	emitCode(self, ")");
}

static void emitFunctionDecl(CCodeGenerator *self, FunctionDecl *decl) {
	// prototype in header
	self->writeState = WRITE_HEADER_STATE;

	emitFunctionSignature(self, decl->signature);
	emitCode(self, ";" CC_NEWLINE);

	// write to the source!
	self->writeState = WRITE_SOURCE_STATE;
	if (decl->body && !decl->prototype) {
		emitFunctionSignature(self, decl->signature);
		emitCode(self, " {" CC_NEWLINE);
		emitBlock(self, decl->body);
		emitCode(self, "}" CC_NEWLINE);
	}
}

static void emitVariableDecl(CCodeGenerator *self, VariableDecl *decl) {
	// hack
	bool isArray = decl->type
					&& decl->type->type == TYPE_LIT_NODE 
					&& decl->type->typeLit->type == ARRAY_TYPE_NODE;
	
	// kinda hacky, but it works
	bool isHeaderVariable = self->writeState == WRITE_HEADER_STATE;

	if (isHeaderVariable && decl->assigned) {
		emitCode(self, "extern ");
	}
	if (!decl->mutable) {
		emitCode(self, "const ");
	}
	
	emitType(self, decl->type);
	emitCode(self, " %s", decl->name);
	if (isArray) {
		emitCode(self, "[");
		if (decl->type->typeLit->arrayType->expr) emitExpression(self, decl->type->typeLit->arrayType->expr);
		emitCode(self, "]");
	}

	if (decl->assigned) {
		if (isHeaderVariable) {
			emitCode(self, ";" CC_NEWLINE);
			self->writeState = WRITE_SOURCE_STATE;
			if (!decl->mutable) {
				emitCode(self, "const ");
			}
			emitType(self, decl->type);
			emitCode(self, " %s", decl->name);
			if (isArray) {
				emitCode(self, "[");
				if (decl->type->typeLit->arrayType->expr) emitExpression(self, decl->type->typeLit->arrayType->expr);
				emitCode(self, "]");
			}
		}
		emitCode(self, " = ");
		emitExpression(self, decl->expr);
	}

	emitCode(self, ";" CC_NEWLINE);
	if (isHeaderVariable) {
		self->writeState = WRITE_HEADER_STATE;
	}

	if (decl->structDecl && !decl->pointer) {
		// Emit struct initializers
		for (int i = 0; i < decl->structDecl->fields->members->size; ++i) {
			FieldDecl *fieldDecl = getVectorItem(decl->structDecl->fields->members, i);
			if (fieldDecl->defaultValue) {
				emitCode(self, "%s", decl->name);
				emitCode(self, "%s", ".");
				emitCode(self, "%s", fieldDecl->name);
				emitCode(self, " = ");
				emitExpression(self, fieldDecl->defaultValue);
				emitCode(self, ";" CC_NEWLINE);
			}
		}
	}
}

static void emitWhileForLoop(CCodeGenerator *self, ForStat *stmt) {
	self->writeState = WRITE_SOURCE_STATE;
	emitCode(self, "while (");
	emitExpression(self, stmt->index);
	emitCode(self, ") {" CC_NEWLINE);
	emitBlock(self, stmt->body);
	emitCode(self, "}" CC_NEWLINE);
}

static void emitBlock(CCodeGenerator *self, Block *block) {
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

static void emitImplDecl(CCodeGenerator *self, ImplDecl *impl) {
	for (int i = 0; i < impl->funcs->size; i++) {
		FunctionDecl *func = getVectorItem(impl->funcs, i);
		emitFunctionDecl(self, func);
	}
}

static void emitAssignment(CCodeGenerator *self, Assignment *assign) {
	emitCode(self, "%s = ", assign->iden);
	emitExpression(self, assign->expr);
	emitCode(self, ";" CC_NEWLINE);
}

static void emitInfiniteForLoop(CCodeGenerator *self, ForStat *stmt) {
	self->writeState = WRITE_SOURCE_STATE;
	emitCode(self, "while (true) {" CC_NEWLINE);
	emitBlock(self, stmt->body);
	emitCode(self, "}" CC_NEWLINE);
}

static void emitMatchClause(CCodeGenerator *self, MatchClause *clause) {
	emitCode(self, "case ");
	emitExpression(self, clause->expr);
	emitCode(self, ": {" CC_NEWLINE);
	emitBlock(self, clause->body);

	// TODO replace this in case someone returns
	emitCode(self, "break", CC_NEWLINE);
	emitCode(self, "};" CC_NEWLINE);
}

static void emitMatchStat(CCodeGenerator *self, MatchStat *match) {
	emitCode(self, "switch (");
	emitExpression(self, match->expr);
	emitCode(self, ") {" CC_NEWLINE);
	for (int i = 0; i < match->clauses->size; i++) {
		emitMatchClause(self, getVectorItem(match->clauses, i));
	}
	emitCode(self, "}" CC_NEWLINE);
}

char *getLoopIndex(CCodeGenerator *self, Expression *expr) {
	switch (expr->exprType) {
		case BINARY_EXPR_NODE: {
			if (expr->binary->lhand) {
				char *name = getLoopIndex(self, expr->binary->lhand);
				return name;
			}
			else {
				// TODO error here
			}
			break;
		}
		case TYPE_NODE: {
			switch (expr->type->type) {
				case TYPE_NAME_NODE: return expr->type->typeName->name;
			}
			break;
		}
	}
	errorMessage("Something went wrong in getLoopIndex");
	return NULL;
}

static void emitIndexForLoop(CCodeGenerator *self, ForStat *stmt) {
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
	if (stmt->index->binary->lhand) {
		char *iden = getLoopIndex(self, stmt->index->binary->lhand);

		emitCode(self, "for (;", iden);

		// emit index
		emitExpression(self, stmt->index);
		emitCode(self, "; ");
		emitExpression(self, stmt->step);

		emitCode(self, ") {" CC_NEWLINE);
		emitBlock(self, stmt->body);
		emitCode(self, "}" CC_NEWLINE);
	}
}

static void emitForStat(CCodeGenerator *self, ForStat *stmt) {
	switch (stmt->forType) {
		case WHILE_FOR_LOOP: emitWhileForLoop(self, stmt); break;
		case INFINITE_FOR_LOOP: emitInfiniteForLoop(self, stmt); break;
		case INDEX_FOR_LOOP: emitIndexForLoop(self, stmt); break;
	}
}

static void emitEnumDecl(CCodeGenerator *self, EnumDecl *enumDecl) {
	emitCode(self, "typedef enum {" CC_NEWLINE);
	for (int i = 0; i < enumDecl->items->size; i++) {
		EnumItem *item = getVectorItem(enumDecl->items, i);
		emitCode(self, "%s", item->name);
		// if (item->val) { FIXME!!
		// 	emitCode(self, "=");
		// 	emitExpression(self, item->val);
		// }
		if (enumDecl->items->size > 1 && i != enumDecl->items->size - 1) {
			emitCode(self, ",");
		}
	}
	emitCode(self, "} %s;" CC_NEWLINE CC_NEWLINE, enumDecl->name);
}

static void emitDeclaration(CCodeGenerator *self, Declaration *decl) {
	switch (decl->type) {
		case FUNCTION_DECL_NODE: emitFunctionDecl(self, decl->funcDecl); break;
		case STRUCT_DECL_NODE: emitStructDecl(self, decl->structDecl); break;
		case VARIABLE_DECL_NODE: emitVariableDecl(self, decl->varDecl); break;
		case IMPL_DECL_NODE: emitImplDecl(self, decl->implDecl); break;
		case ENUM_DECL_NODE: emitEnumDecl(self, decl->enumDecl); break;
		default:
			errorMessage("Unknown node in declaration %s", NODE_NAME[decl->type]);
			break;
	}
}

static void emitReturnStat(CCodeGenerator *self, ReturnStat *ret) {
	emitCode(self, "return ");
	if (ret->expr) {
		emitExpression(self, ret->expr);
	}
	emitCode(self, ";" CC_NEWLINE);
}

static void emitLeaveStat(CCodeGenerator *self, LeaveStat *leave) {
	switch (leave->type) {
		case RETURN_STAT_NODE: emitReturnStat(self, leave->retStmt); break;
		case CONTINUE_STAT_NODE: emitCode(self, "continue;" CC_NEWLINE); break;
		case BREAK_STAT_NODE: emitCode(self, "break;" CC_NEWLINE); break;
	}
}

static void emitFreeStat(CCodeGenerator *self, FreeStat *freeStat) {
	emitCode(self, "free(");
	emitType(self, freeStat->type);
	emitCode(self, ");" CC_NEWLINE);
}

static void emitUnstructuredStat(CCodeGenerator *self, UnstructuredStatement *stmt) {
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
		case FREE_STAT_NODE: emitFreeStat(self, stmt->freeStat); break;
	}
}

static void emitElseIfStat(CCodeGenerator *self, ElseIfStat *elseifs) {
	emitCode(self, " else if (");
	emitExpression(self, elseifs->expr);
	emitCode(self, ") {" CC_NEWLINE);
	emitBlock(self, elseifs->body);
	emitCode(self, "}");
}

static void emitElseStat(CCodeGenerator *self, ElseStat *elses) {
	emitCode(self, " else {" CC_NEWLINE);
	emitBlock(self, elses->body);
	emitCode(self, "}" CC_NEWLINE);
}

static void emitIfStat(CCodeGenerator *self, IfStat *ifs) {
	emitCode(self, "if (");
	emitExpression(self, ifs->expr);
	emitCode(self, ") {" CC_NEWLINE);
	emitBlock(self, ifs->body);
	emitCode(self, "}");

	if (ifs->elseIfStmts == NULL && ifs->elseStmt == NULL) {
		emitCode(self, CC_NEWLINE);
	}

	if (ifs->elseIfStmts != NULL) {
		for (int i = 0; i < ifs->elseIfStmts->size; i++) {
			emitElseIfStat(self, getVectorItem(ifs->elseIfStmts, i));
		}
	}

	if (ifs->elseStmt != NULL) {
		emitElseStat(self, ifs->elseStmt);
	}
}

static void emitStatement(CCodeGenerator *self, Statement *stmt) {
	switch (stmt->type) {
		case UNSTRUCTURED_STATEMENT_NODE: 
			emitUnstructuredStat(self, stmt->unstructured);
			break;
		case STRUCTURED_STATEMENT_NODE: 
			emitStructuredStat(self, stmt->structured);
			break;
		default:
			errorMessage("Unknown node type in block %s", NODE_NAME[stmt->type]);
			break;
	}
}

static void emitStructuredStat(CCodeGenerator *self, StructuredStatement *stmt) {
	switch (stmt->type) {
		case FOR_STAT_NODE: emitForStat(self, stmt->forStmt); break;
		case IF_STAT_NODE: emitIfStat(self, stmt->ifStmt); break;
		case MATCH_STAT_NODE: emitMatchStat(self, stmt->matchStmt); break;
		default:
			errorMessage("Unknown node type found in structured statement %s", NODE_NAME[stmt->type]);
			break;
	}
}

static void traverseAST(CCodeGenerator *self) {
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

void startCCodeGeneration(CCodeGenerator *self) {
	HeaderFile *boilerplate = createHeaderFile("_boilerplate");
	writeHeaderFile(boilerplate);
	fprintf(boilerplate->outputFile, "%s", BOILERPLATE);
	closeHeaderFile(boilerplate);
	
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

		generateMacrosC(self);

		emitCode(self, "#include \"%s\"\n", boilerplate->generatedHeaderName);

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
	buildCommand = sdscat(buildCommand, self->linkerFlags);

	// just for debug purposes
	verboseModeMessage("Running cl args: `%s`", buildCommand);

	// do the command we just created
	int result = system(buildCommand);
	if (result != 0)
		exit(2);
	
	sdsfree(self->linkerFlags);
	sdsfree(buildCommand); // deallocate dat shit baby
	
	destroyHeaderFile(boilerplate);
}

void destroyCCodeGenerator(CCodeGenerator *self) {
	for (int i = 0; i < self->sourceFiles->size; i++) {
		SourceFile *sourceFile = getVectorItem(self->sourceFiles, i);
		destroySourceFile(sourceFile);
		verboseModeMessage("Destroyed source files iteration %d", i);
	}

	free(self);
	verboseModeMessage("Destroyed compiler");
}
