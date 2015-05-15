#ifndef __PRIVATE_H
#define __PRIVATE_H

static void emitCode(CCodeGenerator *self, const char *fmt, ...) {
	va_list args;
	va_start(args, fmt);

	switch (self->writeState) {
		case WRITE_SOURCE_STATE:
			vfprintf(self->currentSourceFile->outputFile, fmt, args);
			break;
		case WRITE_HEADER_STATE:
			vfprintf(self->currentSourceFile->headerFile->outputFile, fmt, args);
			break;
	}

	va_end(args);
}

#endif
