#include "codegen.h"
#include "private.h"

// Declarations

/**
 * Emit a macro for a file inclusion
 * @param self the code gen instance
 * @param use  the use macro
 */
static void emitUseMacro(CCodeGenerator *self, UseMacro *use);

/**
 * Emit a top level node for macros
 * @param self  the code gen instance
 * @param macro the macro to emit
 */
static void emitMacroNode(CCodeGenerator *self, Macro *macro);

// Definitions

static void emitUseMacro(CCodeGenerator *self, UseMacro *use) {
	for (int i = 0; i < use->files->size; i++) {
		char *fileName = getVectorItem(use->files, i);

		size_t len = strlen(fileName);
		char temp[len - 2];
		memcpy(temp, &fileName[1], len - 2);
		temp[len - 2] = '\0';

		emitCode(self, "#include \"_gen_%s.h\"\n", temp);
	}
}

static void emitLinkerFlagMacro(CCodeGenerator *self, LinkerFlagMacro *linker) {
	ALLOY_UNUSED_OBJ(self);

	size_t len = strlen(linker->flag);
	char temp[len - 2];
	memcpy(temp, &linker->flag[1], len - 2);
	temp[len - 2] = '\0';

	self->linkerFlags = sdscat(self->linkerFlags, temp);
	self->linkerFlags = sdscat(self->linkerFlags, " ");
}

static void emitMacroNode(CCodeGenerator *self, Macro *macro) {
	switch (macro->type) {
		case USE_MACRO_NODE: emitUseMacro(self, macro->use); break;
		case LINKER_FLAG_MACRO_NODE: emitLinkerFlagMacro(self, macro->linker); break;
	}
}

void generateMacrosC(CCodeGenerator *self) {
	for (int i = 0; i < self->currentSourceFile->ast->size; i++) {
		Statement *stmt = getVectorItem(self->currentSourceFile->ast, i);

		if (stmt->type == MACRO_NODE) {
			emitMacroNode(self, stmt->macro);
		}
	}
}
