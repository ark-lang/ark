#ifndef PARSER_H
#define PARSER_H

#include <stdio.h>
#include <stdlib.h>

typedef struct s_Tree {
    char* data;
    struct tree* right;
    struct tree* left;
} Tree;

void treeInsert(Tree **tr, char *value);

void treeDelete(Tree **tr);

#endif // PARSER_H
