#include "util.h"

#ifdef _WIN32
HANDLE console_handle;
CONSOLE_SCREEN_BUFFER_INFO console_info;
WORD saved_attributes;
#endif

extern void debug_message(const char *fmt, ...) {
	if (!console_handle) {
		GetStdHandle(STD_OUTPUT_HANDLE);
	}

	if (DEBUG_MODE) {
		va_list arg;
		va_start(arg, fmt);
		
		KYEL();

#ifdef _WIN32
		GetConsoleScreenBufferInfo(console_handle, &console_info);
		saved_attributes = console_info.wAttributes;
		SetConsoleTextAttribute(console_handle, FOREGROUND_BLUE);
#endif

		printf(fmt, arg);
		fflush(stdout);

		KNRM();

#ifdef _WIN32
		SetConsoleTextAttribute(console_handle, saved_attributes);
#endif

		va_end(arg);
	}
}

extern void error_message(const char *fmt, ...) {
	if (!console_handle) {
		GetStdHandle(STD_OUTPUT_HANDLE);
	}

	va_list arg;
	va_start(arg, fmt);
		
	KYEL();

#ifdef _WIN32
	GetConsoleScreenBufferInfo(console_handle, &console_info);
	saved_attributes = console_info.wAttributes;
	SetConsoleTextAttribute(console_handle, FOREGROUND_BLUE);
#endif

	printf(fmt, arg);
	fflush(stdout);

	KNRM();

#ifdef _WIN32
	SetConsoleTextAttribute(console_handle, saved_attributes);
#endif

	va_end(arg);
	exit(1);
}

extern void primary_message(const char *fmt, ...) {
	if (!console_handle) {
		GetStdHandle(STD_OUTPUT_HANDLE);
	}

	va_list arg;
	va_start(arg, fmt);
		
	KYEL();

#ifdef _WIN32
	GetConsoleScreenBufferInfo(console_handle, &console_info);
	saved_attributes = console_info.wAttributes;
	SetConsoleTextAttribute(console_handle, FOREGROUND_BLUE);
#endif

	printf(fmt, arg);

	KNRM();

#ifdef _WIN32
	SetConsoleTextAttribute(console_handle, saved_attributes);
#endif

	va_end(arg);
	exit(1);
}