package codegen

import (
	"fmt"
	"github.com/zer0reaction/lisp-go/parser"
	"github.com/zer0reaction/lisp-go/symbol"
)

// Scratch registers
// rax, rdi, rsi, rdx, rcx, r8, r9, r10, r11

func Codegen(root *parser.Node) string {
	code := ""

	code += ".section .text\n"
	code += ".globl main\n"
	code += "\n"
	code += "main:\n"
	code += "	pushq	%rbp\n"
	code += "	movq	%rsp, %rbp\n"
	code += "	/* --------------- */\n"
	code += codegenNode(root, true)
	code += "	/* --------------- */\n"
	code += "	movq	$0, %rax\n"
	code += "	popq	%rbp\n"
	code += "	ret\n"
	return code
}

// Operands are pushed in the reverse order, for example:
//
// 3 + 4
//
// STACK BASE
// 4 (rval)
// 3 (lval)
func codegenNode(n *parser.Node, orphan bool) string {
	code := ""

	switch n.Type {
	case parser.NodeInteger:
		code += "	/* Integer */\n"
		code += fmt.Sprintf("	pushq	$%d\n",
			symbol.GetIntegerValue(n.TableId))
	case parser.NodeBinOpSum:
		code += codegenNode(n.BinOpRval, false)
		code += codegenNode(n.BinOpLval, false)

		code += "	/* BinOpSum */\n"
		code += "	popq	%rax\n" // lval
		code += "	popq	%rdi\n" // rval
		code += "	addq	%rdi, %rax\n"
		code += "	pushq	%rax\n"
	default:
		panic("node type not implemented")
	}

	if orphan {
		code += "	/* Pop orphan value */\n"
		code += "	add	$8, %rsp\n"
	}

	return code
}
