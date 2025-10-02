package codegen

import (
	"fmt"
	"github.com/zer0reaction/lisp-go/parser"
	sym "github.com/zer0reaction/lisp-go/symbol"
)

// Scratch registers:
// rax, rdi, rsi, rdx, rcx, r8, r9, r10, r11

// Argument registers:
// rdi, rsi, rdx, rcx, r8, r9

var argRegisters = [...]string{"rdi", "rsi", "rdx", "rcx", "r8", "r9"}

const varBytesize uint = 8

// This is really bad.
var stackOffset uint = 0

var externDecls = ""

func Codegen(root *parser.Node) string {
	code := ""

	tmp := ""
	for root != nil {
		tmp += codegenNode(root)
		root = root.Next
	}

	code += ".section .text\n"
	code += ".globl main\n"
	code += externDecls
	code += "\n"
	code += "main:\n"
	code += "	pushq	%rbp\n"
	code += "	movq	%rsp, %rbp\n"

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
//
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
		v := sym.GetVariable(n.Id)
		v.Offset = stackOffset + varBytesize
		sym.SetVariable(n.Id, v)

		stackOffset += varBytesize
	case parser.NodeVariable:
		v := sym.GetVariable(n.Id)

		code += "	/* Variable */\n"
		code += fmt.Sprintf("	movq	-%d(%%rbp), %%rax\n", v.Offset)
		code += "	pushq	%rax\n"
	case parser.NodeInteger:
		code += "	/* Integer */\n"
		code += fmt.Sprintf("	pushq	$%d\n", n.Integer.Value)
	case parser.NodeBinOp:
		code += codegenBinOp(n)
	case parser.NodeFunEx:
		f := sym.GetFunction(n.Id)
		externDecls += fmt.Sprintf(".extern %s\n", f.Name)
	case parser.NodeFunCall:
		cur := n.Function.ArgStart
		argCount := 0

		code += "	/* NodeFunCall */\n"

		for cur != nil {
			if argCount >= len(argRegisters) {
				panic("arguments on stack are not supported yet")
			}
			code += codegenNode(cur)
			code += fmt.Sprintf("	popq	%%%s\n", argRegisters[argCount])
			argCount++
			cur = cur.Next
		}

		f := sym.GetFunction(n.Id)
		code += fmt.Sprintf("	call	%s\n", f.Name)
		code += "	pushq	%rax\n"
	default:
		panic("node type not implemented")
	}

	return code
}

func codegenBinOp(n *parser.Node) string {
	code := ""

	lval := codegenNode(n.BinOp.Lval)
	rval := codegenNode(n.BinOp.Rval)

	switch n.BinOp.Tag {
	case parser.BinOpSum:
		code += rval
		code += lval
		code += "	/* BinOpSum */\n"
		code += "	popq	%rax\n" // lval
		code += "	popq	%rdi\n" // rval
		code += "	addq	%rdi, %rax\n"
		code += "	pushq	%rax\n"
	case parser.BinOpSub:
		code += rval
		code += lval
		code += "	/* BinOpSub */\n"
		code += "	popq	%rax\n" // lval
		code += "	popq	%rdi\n" // rval
		code += "	subq	%rdi, %rax\n"
		code += "	pushq	%rax\n"
	case parser.BinOpAssign:
		v := sym.GetVariable(n.BinOp.Lval.Id)
		offset := v.Offset

		code += rval
		code += "	/* BinOpAssign */\n"
		code += "	popq	%rax\n"
		code += fmt.Sprintf("	movq	%%rax, -%d(%%rbp)\n", offset)
	default:
		panic("node type not implemented")
	}

	return code
}
