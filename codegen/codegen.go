package codegen

import (
	"fmt"
	"github.com/zer0reaction/lisp-go/parser"
	"github.com/zer0reaction/lisp-go/symbol"
)

var boilerplate string = `.section .text
.globl main

main:
	pushq	%rbp
	movq	%rsp, %rbp
	/* --------------- */
`

// Scratch registers
// rax, rdi, rsi, rdx, rcx, r8, r9, r10, r11

// TODO: pop dangling stack value
func Codegen(root *parser.Node) string {
	code := boilerplate
	code += codegenNode(root)
	code += "	/* --------------- */\n"
	code += "	movq	$60, %rax\n"
	code += "	movq	$0, %rdi\n"
	code += "	syscall\n"
	return code
}

// Operands are pushed in the reverse order, for example:
//
// 3 + 4
//
// STACK BASE
// 4 (rval)
// 3 (lval)
func codegenNode(n *parser.Node) string {
	code := ""

	switch n.Type {
	case parser.NodeInteger:
		code += "	/* Integer */\n"
		code += fmt.Sprintf("	pushq	$%d\n",
			symbol.GetIntegerValue(n.TableId))
	case parser.NodeBinOpSum:
		code += codegenNode(n.BinOpRval)
		code += codegenNode(n.BinOpLval)

		code += "	/* BinOpSum */\n"
		code += "	popq	%rax\n" // lval
		code += "	popq	%rdi\n" // rval
		code += "	addq	%rdi, %rax\n"
		code += "	pushq	%rax\n"
	default:
		panic("node type not implemented")
	}

	return code
}
