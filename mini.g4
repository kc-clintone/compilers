grammar MiniLang;

/* * -----------------------------------------
 * PARSER RULES
 * -----------------------------------------
 */

program
    : declaration* EOF
    ;

declaration
    : varDecl ';'
    | structDecl
    | funcDecl
    | statement
    ;

// Variables and Data Structures
varDecl
    : 'var' ID type ('=' expression)?
    ;

structDecl
    : 'type' ID 'struct' '{' structField* '}'
    ;

structField
    : ID type ';'
    ;

// Functions
funcDecl
    : 'func' ID '(' parameters? ')' type? block
    ;

parameters
    : parameter (',' parameter)*
    ;

parameter
    : ID type
    ;

// Types
type
    : 'int'
    | 'char'
    | 'string'
    | 'bool'
    | '[' ']' type                 // Array type
    | 'map' '[' type ']' type      // Hash map type
    | ID                           // Struct type
    ;

// Statements
statement
    : varDecl ';'
    | assignment ';'
    | ifStmt
    | switchStmt
    | forStmt
    | returnStmt ';'
    | expression ';'
    | block
    ;

assignment
    : ID '=' expression
    | ID '[' expression ']' '=' expression     // Array/Map assignment
    | ID '.' ID '=' expression                 // Struct field assignment
    ;

block
    : '{' statement* '}'
    ;

// Control Flow
ifStmt
    : 'if' expression block ('else' (ifStmt | block))?
    ;

switchStmt
    : 'switch' expression '{' caseClause* defaultClause? '}'
    ;

caseClause
    : 'case' expression ':' statement*
    ;

defaultClause
    : 'default' ':' statement*
    ;

forStmt
    : 'for' expression block                                // While-style loop
    | 'for' varDecl ';' expression ';' assignment block     // 3-statement loop
    ;

returnStmt
    : 'return' expression?
    ;

// Expressions
expression
    : '(' expression ')'                             # parenExpr
    | 'print' '(' expression (',' expression)* ')'   # printExpr
    | 'open' '(' expression ',' expression ')'       # openExpr
    | ID '(' arguments? ')'                          # callExpr
    | expression '[' expression ']'                  # indexExpr
    | expression '.' ID                              # fieldAccessExpr
    | ('!' | '-') expression                         # unaryExpr
    | expression ('*' | '/' | '%') expression        # mulDivExpr
    | expression ('+' | '-') expression              # addSubExpr
    | expression ('<' | '<=' | '>' | '>=') expression# relExpr
    | expression ('==' | '!=') expression            # eqExpr
    | expression '&&' expression                     # andExpr
    | expression '||' expression                     # orExpr
    | primary                                        # primaryExpr
    ;

arguments
    : expression (',' expression)*
    ;

primary
    : INT
    | CHAR
    | STRING
    | BOOL
    | ID
    ;

/* * -----------------------------------------
 * LEXER RULES
 * -----------------------------------------
 */

BOOL   : 'true' | 'false' ;
ID     : [a-zA-Z_] [a-zA-Z_0-9]* ;
INT    : [0-9]+ ;
STRING : '"' (~["\\] | '\\' .)* '"' ;
CHAR   : '\'' (~['\\] | '\\' .) '\'' ;

WS     : [ \t\r\n]+ -> skip ;
LINE_COMMENT : '//' ~[\r\n]* -> skip ;
BLOCK_COMMENT: '/*' .*? '*/' -> skip ;