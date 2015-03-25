#include "ast.h"

IdentifierList *createIdentifierList() {
	IdentifierList *list = malloc(sizeof(*list));
	list->values = createVector();
	return list;
}

void destroyIdentifierList(IdentifierList *list) {
	destroyVector(list->values);
	free(list);
}

Expression *createExpression() {
	Expression *expr = malloc(sizeof(*expr));

	return expr;
}

void destroyExpression(Expression *expr) {
	free(expr);
}

TypeName *createTypeName() {
	TypeName *typeName = malloc(sizeof(*typeName));

	return typeName;
}

void destroyTypeName(TypeName *typeName) {
	if (!typeName) return;
	free(typeName);
}

ArrayType *createArrayType() {
	ArrayType *arrayType = malloc(sizeof(*arrayType));

	return arrayType;
}

void destroyArrayType(ArrayType *arrayType) {
	if (!arrayType) return;
	destroyExpression(arrayType->length);
	destroyType(arrayType->type);
	free(arrayType);
}

PointerType *createPointerType() {
	PointerType *pointerType = malloc(sizeof(*pointerType));

	return pointerType;
}

void destroyPointerType(PointerType *pointerType) {
	if (!pointerType) return;
	destroyType(pointerType->type);
	free(pointerType);
}

FieldDecl *createFieldDecl() {
	FieldDecl *fieldDecl = malloc(sizeof(*fieldDecl));

	return fieldDecl;
}

void destroyFieldDecl(FieldDecl *fieldDecl) {
	destroyIdentifierList(fieldDecl->idenList);
	destroyType(fieldDecl->type);
	free(fieldDecl);
}

FieldDeclList *createFieldDeclList() {
	FieldDeclList *fieldDeclList = malloc(sizeof(*fieldDeclList));

	return fieldDeclList;
}

void destroyFieldDeclList(FieldDeclList *fieldDeclList) {
	free(fieldDeclList);
}

StructType *createStructType() {
	StructType *structType = malloc(sizeof(*structType));

	return structType;
}

void destroyStructType(StructType *structType) {
	if (!structType) return;
	destroyFieldDeclList(structType->fields);
	free(structType);
}

FunctionType *createFunctionType() {
	FunctionType *funcType = malloc(sizeof(*funcType));

	return funcType;
}

void destroyFunctionType(FunctionType *funcType) {
	free(funcType);
}

Type *createType() {
	Type *type = malloc(sizeof(*type));
	return type;
}

void destroyType(Type *type) {
	destroyArrayType(type->arrayType);
	destroyPointerType(type->pointerType);
	destroyStructType(type->structType);
	destroyTypeName(type->typeName);
}
