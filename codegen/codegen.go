package codegen

import (
	"fmt"
	"github.com/zer0reaction/lisp-go/parser"
	"github.com/zer0reaction/lisp-go/symbol"
)

// Scratch registers:
// rax, rdi, rsi, rdx, rcx, r8, r9, r10, r11

const varBytesize uint = 8

// This is really bad.
var stackOffset uint = 0

func Codegen(root *parser.Node) string {
	code := ""

	code += ".section .text\n"
	code += ".globl main\n"
	code += "\n"
	code += "main:\n"
	code += "	pushq	%rbp\n"
	code += "	movq	%rsp, %rbp\n"

	tmp := ""
	for root != nil {
		tmp += codegenNode(root)
		root = root.Next
	}
	code += fmt.Sprintf("	leaq	-%d(%%rbp), %%rsp\n", stackOffset)
	code += "	/* --------------- */\n"
	code += tmp

	code += "	/* --------------- */\n"
	code += "	movq	%rbp, %rsp\n"
	code += "	movq	$0, %rax\n"
	code += "	popq	%rbp\n"
	code += "	ret\n"
	return code
}

// Each variable is int64 (8 bytes).
// Currently there are no AST checks at all.
// Operands are pushed in the reverse order, for example:
//
// 3 + 4
//
// STACK BASE
// 4 (rval)
// 3 (lval)
func codegenNode(n *parser.Node) string {
	code := ""

	switch n.Tag {
	case parser.NodeBlock:
		cur := n.Block.Start
		for cur != nil {
			code += codegenNode(cur)
			cur = cur.Next
		}
	case parser.NodeVariableDecl:
		id := n.Variable.TableId
		symbol.VariableSetOffset(id, stackOffset+varBytesize)
		stackOffset += varBytesize
	case parser.NodeVariable:
		offset := symbol.VariableGetOffset(n.Variable.TableId)
		code += "	/* Variable */\n"
		code += fmt.Sprintf("	movq	-%d(%%rbp), %%rax\n", offset)
		code += "	pushq	%rax\n"
	case parser.NodeInteger:
		code += "	/* Integer */\n"
		code += fmt.Sprintf("	pushq	$%d\n", n.Integer.Value)
	case parser.NodeBinOpSum:
		code += codegenNode(n.BinOp.Rval)
		code += codegenNode(n.BinOp.Lval)

		code += "	/* BinOpSum */\n"
		code += "	popq	%rax\n" // lval
		code += "	popq	%rdi\n" // rval
		code += "	addq	%rdi, %rax\n"
		code += "	pushq	%rax\n"
	case parser.NodeBinOpAssign:
		code += codegenNode(n.BinOp.Rval)
		code += "	/* BinOpAssign */\n"
		offset := symbol.VariableGetOffset(n.BinOp.Lval.Variable.TableId)
		code += "	popq	%rax\n"
		code += fmt.Sprintf("	movq	%%rax, -%d(%%rbp)\n", offset)
	default:
		panic("node type not implemented")
	}

	return code
}
