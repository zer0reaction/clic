// This file contains basic x86_64 Linux GAS codegen functions.

package codegen

import (
	"clic/ast"
	"clic/symbol"
	"clic/types"
	"fmt"
)

// Scratch registers:
// rax, rdi, rsi, rdx, rcx, r8, r9, r10, r11

const argRegsCount = 6

var argRegs = [...][argRegsCount]string{
	1: {"dil", "sil", "dl", "cl", "r8b", "r9b"},
	2: {"di", "si", "dx", "cx", "r8w", "r9w"},
	4: {"edi", "esi", "edx", "ecx", "r8d", "r9d"},
	8: {"rdi", "rsi", "rdx", "rcx", "r8", "r9"},
}

var externDecls = ""

var localCount = 0

func Codegen(roots []*ast.Node) string {
	code := ""

	tmp := ""
	for _, node := range roots {
		tmp += genNode(node)
	}

	code += ".section .text\n"
	code += ".globl main\n"
	code += externDecls
	code += tmp

	return code
}

// Operands are pushed in the reverse order, for example:
//
// 3 + 4
//
// STACK BASE
// 4 (rval)
// 3 (lval)
func genNode(n *ast.Node) string {
	code := ""

	switch n.Tag {
	case ast.NodeBlock:
		for _, node := range n.Block.Stmts {
			code += genNode(node)
		}

	case ast.NodeLocVar:
		sym := symbol.Get(n.Id)
		offset := sym.LocVar.Offset
		code += fmt.Sprintf("	movq	-%d(%%rbp), %%rax\n", offset)
		code += "	pushq	%rax\n"

	case ast.NodeInteger:
		if n.Integer.Signed {
			code += fmt.Sprintf("	pushq	$%d\n", n.Integer.SValue)
		} else {
			code += fmt.Sprintf("	pushq	$%d\n", n.Integer.UValue)
		}

	case ast.NodeBinOp:
		code += genBinOp(n)

	case ast.NodeFunEx:
		name := symbol.Get(n.Id).Name
		externDecls += fmt.Sprintf(".extern %s\n", name)

	case ast.NodeFunCall:
		if len(n.Function.Args) >= argRegsCount {
			panic("arguments on stack are not supported yet")
		}

		for i, node := range n.Function.Args {
			code += genNode(node)
			code += fmt.Sprintf("	popq	%%%s\n", argRegs[8][i])
		}

		name := symbol.Get(n.Id).Name
		code += fmt.Sprintf("	call	%s\n", name)
		code += "	pushq	%rax\n"

	case ast.NodeBoolean:
		if n.Boolean.Value {
			code += "	pushq	$1\n"
		} else {
			code += "	pushq	$0\n"
		}

	case ast.NodeIf:
		code += genIf(n)

	case ast.NodeWhile:
		code += genWhile(n)

	case ast.NodeFor:
		code += genFor(n)

	case ast.NodeCast:
		// Right now there are only integer types, so we can
		// simply push the node's value on stack. This code
		// only checks for new and unsupported types.

		from := n.Cast.What.GetTypeDeep()

		switch from {
		// Do nothing
		case types.GetBuiltin(types.S64):
		case types.GetBuiltin(types.U64):
		case types.GetBuiltin(types.Bool):

		case types.GetBuiltin(types.Void):
			panic("trying to cast from type 'void'")

		default:
			panic("not implemented")
		}

		code += genNode(n.Cast.What)

	case ast.NodeFunDef:
		code += genFunction(n)

	case ast.NodeReturn:
		code += genNode(n.Return.Value)
		code += "	popq	%rax\n"
		code += "	movq	%rbp, %rsp\n"
		code += "	popq	%rbp\n"
		code += "	ret\n"

	// Do nothing
	case ast.NodeTypedef:
	case ast.NodeEmpty:
	case ast.NodeVarDecl:

	default:
		panic("not implemented")
	}

	return code
}

func genBinOp(n *ast.Node) string {
	code := ""

	lval := genNode(n.BinOp.Lval)
	rval := genNode(n.BinOp.Rval)

	switch n.BinOp.Tag {
	case ast.BinOpAssign:
		sym := symbol.Get(n.BinOp.Lval.Id)

		if sym.Tag == symbol.LocVar {
			offset := sym.LocVar.Offset
			code += rval
			code += "	popq	%rax\n"
			code += fmt.Sprintf("	movq	%%rax, -%d(%%rbp)\n", offset)
		} else {
			panic("not implemented")
		}

	case ast.BinOpArith:
		switch n.BinOp.ArithTag {
		case ast.BinOpSum:
			code += rval
			code += lval
			code += "	popq	%rax\n" // lval
			code += "	popq	%rdi\n" // rval
			code += "	addq	%rdi, %rax\n"
			code += "	pushq	%rax\n"

		case ast.BinOpSub:
			code += rval
			code += lval
			code += "	popq	%rax\n" // lval
			code += "	popq	%rdi\n" // rval
			code += "	subq	%rdi, %rax\n"
			code += "	pushq	%rax\n"

		case ast.BinOpMult:
			code += rval
			code += lval
			code += "	popq	%rax\n" // lval
			code += "	popq	%rdi\n" // rval
			code += "	imulq	%rdi, %rax\n"
			// TODO: The result is actually stored in
			// [rdx:rax], is this ok to do?
			code += "	pushq	%rax\n"

		case ast.BinOpDiv:
			code += rval
			code += lval
			code += "	popq	%rax\n" // lval
			code += "	popq	%rdi\n" // rval

			// R[%rax] <- R[%rdx]:R[%rax] / S

			code += "	cqto\n" // sign extend rax to [rdx:rax]
			code += "	idivq	%rdi\n"
			code += "	pushq	%rax\n"

		case ast.BinOpMod:
			code += rval
			code += lval
			code += "	popq	%rax\n" // lval
			code += "	popq	%rdi\n" // rval

			// R[%rdx] <- R[%rdx]:R[%rax] mod S

			code += "	cqto\n" // sign extend rax to [rdx:rax]
			code += "	idivq	%rdi\n"
			code += "	pushq	%rdx\n"

		default:
			panic("not implemented")
		}

	case ast.BinOpComp:
		switch n.BinOp.CompTag {
		case ast.BinOpEq:
			code += rval
			code += lval
			code += "	popq	%rax\n" // lval
			code += "	popq	%rdi\n" // rval
			code += "	xorq	%rsi, %rsi\n"
			code += "	cmpq	%rdi, %rax\n"
			code += "	sete	%sil\n"
			code += "	pushq	%rsi\n"

		case ast.BinOpNeq:
			code += rval
			code += lval
			code += "	popq	%rax\n" // lval
			code += "	popq	%rdi\n" // rval
			code += "	xorq	%rsi, %rsi\n"
			code += "	cmpq	%rdi, %rax\n"
			code += "	setne	%sil\n"
			code += "	pushq	%rsi\n"

		case ast.BinOpLessEq:
			code += rval
			code += lval
			code += "	popq	%rax\n" // lval
			code += "	popq	%rdi\n" // rval
			code += "	xorq	%rsi, %rsi\n"
			code += "	cmpq	%rdi, %rax\n"
			code += "	setle	%sil\n"
			code += "	pushq	%rsi\n"

		case ast.BinOpLess:
			code += rval
			code += lval
			code += "	popq	%rax\n" // lval
			code += "	popq	%rdi\n" // rval
			code += "	xorq	%rsi, %rsi\n"
			code += "	cmpq	%rdi, %rax\n"
			code += "	setl	%sil\n"
			code += "	pushq	%rsi\n"

		case ast.BinOpGreatEq:
			code += rval
			code += lval
			code += "	popq	%rax\n" // lval
			code += "	popq	%rdi\n" // rval
			code += "	xorq	%rsi, %rsi\n"
			code += "	cmpq	%rdi, %rax\n"
			code += "	setge	%sil\n"
			code += "	pushq	%rsi\n"

		case ast.BinOpGreat:
			code += rval
			code += lval
			code += "	popq	%rax\n" // lval
			code += "	popq	%rdi\n" // rval
			code += "	xorq	%rsi, %rsi\n"
			code += "	cmpq	%rdi, %rax\n"
			code += "	setg	%sil\n"
			code += "	pushq	%rsi\n"

		default:
			panic("not implemented")
		}

	default:
		panic("not implemented")
	}

	return code
}

func genIf(n *ast.Node) string {
	code := ""

	code += genNode(n.If.Exp)

	code += "	popq	%rax\n"
	code += "	cmpq	$0, %rax\n"

	// End of the entire if/else block.
	end := fmt.Sprintf(".L%d", localCount)
	localCount += 1

	if len(n.If.ElseStmts) == 0 {
		code += fmt.Sprintf("	je	%s\n", end)
		for _, stmt := range n.If.IfStmts {
			code += genNode(stmt)
		}
	} else {
		elseStart := fmt.Sprintf(".L%d", localCount)
		localCount += 1

		code += fmt.Sprintf("	je	%s\n", elseStart)
		for _, stmt := range n.If.IfStmts {
			code += genNode(stmt)
		}
		code += fmt.Sprintf("	jmp	%s\n", end)
		code += fmt.Sprintf("%s:\n", elseStart)
		for _, stmt := range n.If.ElseStmts {
			code += genNode(stmt)
		}
	}

	code += fmt.Sprintf("%s:\n", end)

	return code
}

func genWhile(n *ast.Node) string {
	code := ""

	start := fmt.Sprintf(".L%d", localCount)
	localCount += 1
	end := fmt.Sprintf(".L%d", localCount)
	localCount += 1

	code += fmt.Sprintf("%s:\n", start)
	code += genNode(n.While.Exp)
	code += "	popq	%rax\n"
	code += "	cmpq	$0, %rax\n"
	code += fmt.Sprintf("	jz	%s\n", end)
	for _, stmt := range n.While.Stmts {
		code += genNode(stmt)
	}
	code += fmt.Sprintf("	jmp	%s\n", start)
	code += fmt.Sprintf("%s:\n", end)

	return code
}

func genFor(n *ast.Node) string {
	code := ""

	start := fmt.Sprintf(".L%d", localCount)
	localCount += 1
	end := fmt.Sprintf(".L%d", localCount)
	localCount += 1

	code += genNode(n.For.Init)
	code += fmt.Sprintf("%s:\n", start)
	code += genNode(n.For.Cond)
	code += "	popq	%rax\n"
	code += "	cmpq	$0, %rax\n"
	code += fmt.Sprintf("	jz	%s\n", end)
	for _, stmt := range n.For.Stmts {
		code += genNode(stmt)
	}
	code += genNode(n.For.Adv)
	code += fmt.Sprintf("	jmp	%s\n", start)
	code += fmt.Sprintf("%s:\n", end)

	return code
}

func setVarOffsets(n *ast.Node, reserv uint) uint {
	switch n.Tag {
	case ast.NodeBlock:
		for _, stmt := range n.Block.Stmts {
			reserv = setVarOffsets(stmt, reserv)
		}

	case ast.NodeVarDecl:
		sym := symbol.Get(n.Id)

		if sym.Tag == symbol.LocVar {
			size := types.Get(sym.LocVar.Type).Size
			sym.LocVar.Offset = reserv + size
			reserv += size
			symbol.Set(n.Id, sym)
		}

	case ast.NodeBinOp:
		reserv = setVarOffsets(n.BinOp.Lval, reserv)
		reserv = setVarOffsets(n.BinOp.Rval, reserv)

	case ast.NodeFunCall:
		for _, arg := range n.Function.Args {
			reserv = setVarOffsets(arg, reserv)
		}

	case ast.NodeIf:
		reserv = setVarOffsets(n.If.Exp, reserv)
		for _, stmt := range n.If.IfStmts {
			reserv = setVarOffsets(stmt, reserv)
		}
		for _, stmt := range n.If.ElseStmts {
			reserv = setVarOffsets(stmt, reserv)
		}

	case ast.NodeWhile:
		reserv = setVarOffsets(n.While.Exp, reserv)
		for _, stmt := range n.While.Stmts {
			reserv = setVarOffsets(stmt, reserv)
		}

	case ast.NodeFor:
		reserv = setVarOffsets(n.For.Init, reserv)
		reserv = setVarOffsets(n.For.Cond, reserv)
		reserv = setVarOffsets(n.For.Adv, reserv)
		for _, stmt := range n.For.Stmts {
			reserv = setVarOffsets(stmt, reserv)
		}

	case ast.NodeCast:
		reserv = setVarOffsets(n.Cast.What, reserv)

	case ast.NodeFunDef:
		for _, stmt := range n.Function.Stmts {
			reserv = setVarOffsets(stmt, reserv)
		}

	case ast.NodeReturn:
		reserv = setVarOffsets(n.Return.Value, reserv)

	case ast.NodeInteger:
	case ast.NodeLocVar:
	case ast.NodeFunEx:
	case ast.NodeBoolean:
	case ast.NodeTypedef:
	case ast.NodeEmpty:

	default:
		panic("not implemented")
	}

	return reserv
}

func genFunction(n *ast.Node) string {
	code := ""
	reserv := uint(0)

	code += "\n"
	code += symbol.Get(n.Id).Name + ":\n"
	code += "	pushq	%rbp\n"
	code += "	movq	%rsp, %rbp\n"

	for i, param := range n.Function.Params {
		sym := symbol.Get(param)
		if sym.Tag != symbol.LocVar {
			panic("param != local var")
		}

		typeNode := types.Get(sym.LocVar.Type)
		size := typeNode.Size

		sym.LocVar.Offset = reserv + size
		reserv += size
		symbol.Set(param, sym)

		code += fmt.Sprintf("	mov	%%%s, -%d(%%rbp)\n",
			argRegs[size][i], sym.LocVar.Offset)
	}
	reserv = setVarOffsets(n, reserv)
	reserv += (16 - (reserv % 16)) % 16
	code += fmt.Sprintf("	subq	$%d, %%rsp\n", reserv)

	for _, stmt := range n.Function.Stmts {
		code += genNode(stmt)
	}

	code += "	movq	%rbp, %rsp\n"
	code += "	popq	%rbp\n"
	code += "	ret\n"

	return code
}
