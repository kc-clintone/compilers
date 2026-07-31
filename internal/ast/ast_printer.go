package ast

import (
	"fmt"
	"strings"
)

// Print formats and displays a structured, indented tree representation of the AST.
func Print(prog *Program) {
	if prog == nil {
		fmt.Println("<nil Program>")
		return
	}
	fmt.Println("Program")
	for _, decl := range prog.Decls {
		printNode(decl, 1)
	}
	for _, stmt := range prog.Stmts {
		printNode(stmt, 1)
	}
}

func printNode(node Node, indent int) {
	if node == nil {
		return
	}
	pad := strings.Repeat("  ", indent)

	switch n := node.(type) {
	case *FuncDecl:
		fmt.Printf("%sFuncDecl: %s(%s)\n", pad, n.Name, strings.Join(n.Params, ", "))
		printNode(n.Body, indent+1)

	case *VarDecl:
		fmt.Printf("%sVarDecl: %s\n", pad, n.Name)
		if n.Init != nil {
			printNode(n.Init, indent+1)
		}

	case *AssignStmt:
		fmt.Printf("%sAssignStmt: %s =\n", pad, n.Name)
		printNode(n.Value, indent+1)

	case *IfStmt:
		fmt.Printf("%sIfStmt:\n", pad)
		fmt.Printf("%s  Cond:\n", pad)
		printNode(n.Cond, indent+2)
		fmt.Printf("%s  Then:\n", pad)
		printNode(n.Then, indent+2)
		if n.Else != nil {
			fmt.Printf("%s  Else:\n", pad)
			printNode(n.Else, indent+2)
		}

	case *ForStmt:
		fmt.Printf("%sForStmt:\n", pad)
		fmt.Printf("%s  Cond:\n", pad)
		printNode(n.Cond, indent+2)
		fmt.Printf("%s  Body:\n", pad)
		printNode(n.Body, indent+2)

	case *BlockStmt:
		fmt.Printf("%sBlockStmt:\n", pad)
		for _, s := range n.Stmts {
			printNode(s, indent+1)
		}

	case *ReturnStmt:
		fmt.Printf("%sReturnStmt:\n", pad)
		if n.Value != nil {
			printNode(n.Value, indent+1)
		}

	case *PrintStmt:
		fmt.Printf("%sPrintStmt:\n", pad)
		for _, arg := range n.Args {
			printNode(arg, indent+1)
		}

	case *ExprStmt:
		printNode(n.Expr, indent)

	case *BinaryExpr:
		fmt.Printf("%sBinaryExpr (%s):\n", pad, n.Op)
		printNode(n.Left, indent+1)
		printNode(n.Right, indent+1)

	case *UnaryExpr:
		fmt.Printf("%sUnaryExpr (%s):\n", pad, n.Op)
		printNode(n.Right, indent+1)

	case *CallExpr:
		fmt.Printf("%sCallExpr: %s()\n", pad, n.Callee)
		for _, arg := range n.Args {
			printNode(arg, indent+1)
		}

	case *IdentExpr:
		fmt.Printf("%sIdent: %s\n", pad, n.Name)

	case *LiteralExpr:
		fmt.Printf("%sLiteral: %v\n", pad, n.Value)

	default:
		fmt.Printf("%sUnknown AST Node %T\n", pad, node)
	}
}
