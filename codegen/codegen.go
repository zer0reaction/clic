// This file contains basic x86_64 Linux GAS codegen functions.

package codegen

import (
	"fmt"
	"lisp-go/ast"
	sym "lisp-go/symbol"
	"lisp-go/types"
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

var localCount = 0

func Codegen(roots []*ast.Node) string {
	code := ""

	tmp := ""
	for _, node := range roots {
		tmp += codegenNode(node)
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

// Operands are pushed in the reverse order, for example:
//
// 3 + 4
//
// STACK BASE
// 4 (rval)
// 3 (lval)
func codegenNode(n *ast.Node) string {
	code := ""

	switch n.Tag {
	case ast.NodeBlock:
		for _, node := range n.Block.Stmts {
			code += codegenNode(node)
		}

	case ast.NodeVariableDecl:
		v := sym.GetVariable(n.Id)
		v.Offset = stackOffset + varBytesize
		sym.SetVariable(n.Id, v)

		stackOffset += varBytesize

	case ast.NodeVariable:
		v := sym.GetVariable(n.Id)

		code += "	/* Variable */\n"
		code += fmt.Sprintf("	movq	-%d(%%rbp), %%rax\n", v.Offset)
		code += "	pushq	%rax\n"

	case ast.NodeInteger:
		code += "	/* Integer */\n"
		code += fmt.Sprintf("	pushq	$%d\n", n.Integer.Value)

	case ast.NodeBinOp:
		code += codegenBinOp(n)

	case ast.NodeFunEx:
		f := sym.GetFunction(n.Id)
		externDecls += fmt.Sprintf(".extern %s\n", f.Name)

	case ast.NodeFunCall:
		code += "	/* FunCall */\n"

		if len(n.FunCall.Args) >= len(argRegisters) {
			panic("arguments on stack are not supported yet")
		}

		for i, node := range n.FunCall.Args {
			code += codegenNode(node)
			code += fmt.Sprintf("	popq	%%%s\n", argRegisters[i])
		}

		f := sym.GetFunction(n.Id)
		code += fmt.Sprintf("	call	%s\n", f.Name)
		code += "	pushq	%rax\n"

	case ast.NodeBoolean:
		code += "	/* Boolean */\n"
		if n.Boolean.Value {
			code += "	pushq	$1\n"
		} else {
			code += "	pushq	$0\n"
		}

	case ast.NodeIf:
		code += codegenIf(n)

	case ast.NodeWhile:
		code += codegenWhile(n)

	case ast.NodeCast:
		// Right now there are only integer types, so we can
		// simply push the node's value on stack. This code
		// only checks for new and unsupported types.

		from := n.Cast.What.GetType()

		switch from {
		// Do nothing
		case types.S64:
		case types.U64:
		case types.Bool:

		case types.None:
			panic("trying to cast from type None")

		default:
			panic("not implemented")
		}

		code += codegenNode(n.Cast.What)

	default:
		panic("not implemented")
	}

	return code
}

func codegenBinOp(n *ast.Node) string {
	code := ""

	lval := codegenNode(n.BinOp.Lval)
	rval := codegenNode(n.BinOp.Rval)

	switch n.BinOp.Tag {
	case ast.BinOpAssign:
		v := sym.GetVariable(n.BinOp.Lval.Id)
		offset := v.Offset

		code += rval
		code += "	/* BinOpAssign */\n"
		code += "	popq	%rax\n"
		code += fmt.Sprintf("	movq	%%rax, -%d(%%rbp)\n", offset)

	case ast.BinOpArith:
		switch n.BinOp.ArithTag {
		case ast.BinOpSum:
			code += rval
			code += lval
			code += "	/* BinOpSum */\n"
			code += "	popq	%rax\n" // lval
			code += "	popq	%rdi\n" // rval
			code += "	addq	%rdi, %rax\n"
			code += "	pushq	%rax\n"

		case ast.BinOpSub:
			code += rval
			code += lval
			code += "	/* BinOpSub */\n"
			code += "	popq	%rax\n" // lval
			code += "	popq	%rdi\n" // rval
			code += "	subq	%rdi, %rax\n"
			code += "	pushq	%rax\n"

		case ast.BinOpMult:
			code += rval
			code += lval
			code += "	/* BinOpMult */\n"
			code += "	popq	%rax\n" // lval
			code += "	popq	%rdi\n" // rval
			code += "	imulq	%rdi, %rax\n"
			// TODO: The result is actually stored in
			// [rdx:rax], is this ok to do?
			code += "	pushq	%rax\n"

		case ast.BinOpDiv:
			code += rval
			code += lval
			code += "	/* BinOpDiv */\n"
			code += "	popq	%rax\n" // lval
			code += "	popq	%rdi\n" // rval

			// R[%rax] <- R[%rdx]:R[%rax] / S

			code += "	cqto\n" // sign extend rax to [rdx:rax]
			code += "	idivq	%rdi\n"
			code += "	pushq	%rax\n"

		case ast.BinOpMod:
			code += rval
			code += lval
			code += "	/* BinOpMod */\n"
			code += "	popq	%rax\n" // lval
			code += "	popq	%rdi\n" // rval

			// R[%rdx] <- R[%rdx]:R[%rax] mod S

			code += "	cqto\n" // sign extend rax to [rdx:rax]
			code += "	idivq	%rdi\n"
			code += "	pushq	%rdx\n"

		default:
			panic("invalid arith tag")
		}

	case ast.BinOpComp:
		switch n.BinOp.CompTag {
		case ast.BinOpEq:
			code += rval
			code += lval
			code += "	/* BinOpEq */\n"
			code += "	popq	%rax\n" // lval
			code += "	popq	%rdi\n" // rval
			code += "	xorq	%rsi, %rsi\n"
			code += "	cmpq	%rdi, %rax\n"
			code += "	sete	%sil\n"
			code += "	pushq	%rsi\n"

		case ast.BinOpNeq:
			code += rval
			code += lval
			code += "	/* BinOpNeq */\n"
			code += "	popq	%rax\n" // lval
			code += "	popq	%rdi\n" // rval
			code += "	xorq	%rsi, %rsi\n"
			code += "	cmpq	%rdi, %rax\n"
			code += "	setne	%sil\n"
			code += "	pushq	%rsi\n"

		case ast.BinOpLessEq:
			code += rval
			code += lval
			code += "	/* BinOpLessEq */\n"
			code += "	popq	%rax\n" // lval
			code += "	popq	%rdi\n" // rval
			code += "	xorq	%rsi, %rsi\n"
			code += "	cmpq	%rdi, %rax\n"
			code += "	setle	%sil\n"
			code += "	pushq	%rsi\n"

		case ast.BinOpLess:
			code += rval
			code += lval
			code += "	/* BinOpLess */\n"
			code += "	popq	%rax\n" // lval
			code += "	popq	%rdi\n" // rval
			code += "	xorq	%rsi, %rsi\n"
			code += "	cmpq	%rdi, %rax\n"
			code += "	setl	%sil\n"
			code += "	pushq	%rsi\n"

		case ast.BinOpGreatEq:
			code += rval
			code += lval
			code += "	/* BinOpGreatEq */\n"
			code += "	popq	%rax\n" // lval
			code += "	popq	%rdi\n" // rval
			code += "	xorq	%rsi, %rsi\n"
			code += "	cmpq	%rdi, %rax\n"
			code += "	setge	%sil\n"
			code += "	pushq	%rsi\n"

		case ast.BinOpGreat:
			code += rval
			code += lval
			code += "	/* BinOpGreat */\n"
			code += "	popq	%rax\n" // lval
			code += "	popq	%rdi\n" // rval
			code += "	xorq	%rsi, %rsi\n"
			code += "	cmpq	%rdi, %rax\n"
			code += "	setg	%sil\n"
			code += "	pushq	%rsi\n"

		default:
			panic("invalid comp tag")
		}

	default:
		panic("invalid binop tag")
	}

	return code
}

func codegenIf(n *ast.Node) string {
	code := ""

	code += codegenNode(n.If.Exp)

	code += "	/* If */\n"
	code += "	popq	%rax\n"
	code += "	cmpq	$0, %rax\n"

	// End of the entire if/else block.
	end := fmt.Sprintf(".L%d", localCount)
	localCount += 1

	if n.If.ElseBody == nil {
		code += fmt.Sprintf("	je	%s\n", end)
		code += codegenNode(n.If.IfBody)
	} else {
		elseStart := fmt.Sprintf(".L%d", localCount)
		localCount += 1

		code += fmt.Sprintf("	je	%s\n", elseStart)
		code += codegenNode(n.If.IfBody)
		code += fmt.Sprintf("	jmp	%s\n", end)
		code += fmt.Sprintf("%s:\n", elseStart)
		code += codegenNode(n.If.ElseBody)
	}

	code += fmt.Sprintf("%s:\n", end)

	return code
}

func codegenWhile(n *ast.Node) string {
	code := ""

	start := fmt.Sprintf(".L%d", localCount)
	localCount += 1
	end := fmt.Sprintf(".L%d", localCount)
	localCount += 1

	code += "	/* While */\n"
	code += fmt.Sprintf("%s:\n", start)
	code += codegenNode(n.While.Exp)
	code += "	popq	%rax\n"
	code += "	cmpq	$0, %rax\n"
	code += fmt.Sprintf("	jz	%s\n", end)
	code += codegenNode(n.While.Body)
	code += fmt.Sprintf("	jmp	%s\n", start)
	code += fmt.Sprintf("%s:\n", end)

	return code
}
