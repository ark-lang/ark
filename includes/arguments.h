#ifndef __ARGUMENT_H
#define __ARGUMENT_H

#include "util.h"
#include "sourcefile.h"
#include "hashmap.h"

typedef struct {
    char *argName;
    char *argDescription;
    void (*action)(void);
    size_t arguments;
} Argument;

static map_t *arguments;
static int arg_count;
static char **arg_value;
static Argument *currentArgument;
static int currentArgumentIndex;

Argument *createArgument(char *argName, char *argDescription, size_t arguments, void (*action)(void));

void destroyArgument(Argument *arg);

Vector *setup_arguments(int argc, char** argv);

void help();

#endif // __ARGUMENT_H