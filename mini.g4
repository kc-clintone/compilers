grammar Zing;

program: declaration* statement* EOF;

declaration
    : varDecl ';'
    | structDecl
    | funcDecl
    ;

varDecl: 'var' ID type ('=' expression)?;
structDecl: 'type' ID 'struct' '{' structField* '}';
structField: ID type ';';
funcDecl: 'func' ID '(' parameters? ')' type? block;
parameters: parameter (',' parameter)*;
parameter: ID type;

type
    : 'int'
    | 'char'
    | 'string'
    | 'bool'
    | '[' ']' type
    | 'map' '[' type ']' type
    | ID
    ;

statement
    : varDecl ';'
    | expression '=' expression ';'
    | ifStmt
    | switchStmt
    | forStmt
    | 'break' ';'
    | 'continue' ';'
    | returnStmt ';'
    | expression ';'
    | block
    ;

block: '{' statement* '}';
ifStmt: 'if' expression block ('else' (ifStmt | block))?;
switchStmt: 'switch' expression '{' caseClause* defaultClause? '}';
caseClause: 'case' expression ':' statement*;
defaultClause: 'default' ':' statement*;
forStmt
    : 'for' expression block
    | 'for' varDecl ';' expression ';' expression '=' expression block
    ;
returnStmt: 'return' expression?;

expression
    : '(' expression ')'                                      # parenExpr
    | type '{' compositeElements? '}'                         # compositeExpr
    | 'make' '(' type (',' expression)? ')'                   # makeExpr
    | ID '(' arguments? ')'                                   # callExpr
    | expression '[' expression? ':' expression? ']'         # sliceExpr
    | expression '[' expression ']'                           # indexExpr
    | expression '.' ID                                       # fieldAccessExpr
    | ('!' | '-') expression                                  # unaryExpr
    | expression ('*' | '/' | '%') expression                 # mulDivExpr
    | expression ('+' | '-') expression                       # addSubExpr
    | expression ('<' | '<=' | '>' | '>=') expression        # relExpr
    | expression ('==' | '!=') expression                     # eqExpr
    | expression '&&' expression                              # andExpr
    | expression '||' expression                              # orExpr
    | primary                                                 # primaryExpr
    ;

compositeElements: compositeElement (',' compositeElement)* ','?;
compositeElement: ID ':' expression | expression ':' expression | expression;
arguments: expression (',' expression)*;
primary: INT | CHAR | STRING | BOOL | ID;

BOOL: 'true' | 'false';
ID: [a-zA-Z_] [a-zA-Z_0-9]*;
INT: [0-9]+;
STRING: '"' (~["\\] | '\\' .)* '"';
CHAR: '\'' (~['\\] | '\\' .) '\'';
WS: [ \t\r\n]+ -> skip;
LINE_COMMENT: '//' ~[\r\n]* -> skip;
BLOCK_COMMENT: '/*' .*? '*/' -> skip;
