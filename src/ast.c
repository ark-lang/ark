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

StructDecl *createStructDecl() {
	StructDecl *structDecl = malloc(sizeof(*structDecl));

	return structDecl;
}

void destroyStructDecl(StructDecl *structDecl) {
	if (!structDecl) return;
	destroyFieldDeclList(structDecl->fields);
	free(structDecl);
}

FunctionSignature *createFunctionSignature() {
	FunctionSignature *funcSignature = malloc(sizeof(*funcSignature));

	return funcSignature;
}

void destroyFunctionSignature(FunctionSignature *funcSignature) {
	free(funcSignature);
}

Type *createType() {
	Type *type = malloc(sizeof(*type));
	return type;
}

void destroyType(Type *type) {
	destroyArrayType(type->arrayType);
	destroyPointerType(type->pointerType);
	destroyStructDecl(type->structType);
	destroyTypeName(type->typeName);
}
