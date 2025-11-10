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

func Codegen(roots []*ast.Node, t *symbol.Table) string {
	code := ""

	tmp := ""
	for _, node := range roots {
		tmp += genNode(node, t)
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
func genNode(n *ast.Node, t *symbol.Table) string {
	code := ""

	switch n.Tag {
	case ast.NodeScope:
		for _, node := range n.Scope.Stmts {
			code += genNode(node, t)
		}

	case ast.NodeLVar:
		sym := t.Get(n.Id)
		offset := sym.LVar.Offset
		code += fmt.Sprintf("	movq	-%d(%%rbp), %%rax\n", offset)
		code += "	pushq	%rax\n"

	case ast.NodeInt:
		if n.Int.Signed {
			code += fmt.Sprintf("	pushq	$%d\n", n.Int.SValue)
		} else {
			code += fmt.Sprintf("	pushq	$%d\n", n.Int.UValue)
		}

	case ast.NodeBinOp:
		code += genBinOp(n, t)

	case ast.NodeFunEx:
		name := t.Get(n.Id).Name
		externDecls += fmt.Sprintf(".extern %s\n", name)

	case ast.NodeFunCall:
		if len(n.Fun.Args) >= argRegsCount {
			panic("arguments on stack are not supported yet")
		}

		for i, node := range n.Fun.Args {
			code += genNode(node, t)
			code += fmt.Sprintf("	popq	%%%s\n", argRegs[8][i])
		}

		name := t.Get(n.Id).Name
		code += fmt.Sprintf("	call	%s\n", name)
		code += "	pushq	%rax\n"

	case ast.NodeBool:
		if n.Bool.Value {
			code += "	pushq	$1\n"
		} else {
			code += "	pushq	$0\n"
		}

	case ast.NodeIf:
		code += genIf(n, t)

	case ast.NodeWhile:
		code += genWhile(n, t)

	case ast.NodeFor:
		code += genFor(n, t)

	case ast.NodeCast:
		// Right now there are only integer types, so we can
		// simply push the node's value on stack. This code
		// only checks for new and unsupported types.

		from := n.Cast.What.GetTypeDeep(t)

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

		code += genNode(n.Cast.What, t)

	case ast.NodeFunDef:
		code += genFunction(n, t)

	case ast.NodeReturn:
		code += genNode(n.Return.Val, t)
		code += "	popq	%rax\n"
		code += "	movq	%rbp, %rsp\n"
		code += "	popq	%rbp\n"
		code += "	ret\n"

	// Do nothing
	case ast.NodeTypedef:
	case ast.NodeEmpty:
	case ast.NodeLVarDecl:
	case ast.NodeFunDecl:

	default:
		panic("not implemented")
	}

	return code
}

func genBinOp(n *ast.Node, t *symbol.Table) string {
	code := ""

	lval := genNode(n.BinOp.Lval, t)
	rval := genNode(n.BinOp.Rval, t)

	switch n.BinOp.Tag {
	case ast.BinOpAssign:
		sym := t.Get(n.BinOp.Lval.Id)

		if sym.Tag == symbol.LVar {
			offset := sym.LVar.Offset
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

func genIf(n *ast.Node, t *symbol.Table) string {
	code := ""

	code += genNode(n.If.Exp, t)

	code += "	popq	%rax\n"
	code += "	cmpq	$0, %rax\n"

	// End of the entire if/else block.
	end := fmt.Sprintf(".L%d", localCount)
	localCount += 1

	if len(n.If.ElseStmts) == 0 {
		code += fmt.Sprintf("	je	%s\n", end)
		for _, stmt := range n.If.IfStmts {
			code += genNode(stmt, t)
		}
	} else {
		elseStart := fmt.Sprintf(".L%d", localCount)
		localCount += 1

		code += fmt.Sprintf("	je	%s\n", elseStart)
		for _, stmt := range n.If.IfStmts {
			code += genNode(stmt, t)
		}
		code += fmt.Sprintf("	jmp	%s\n", end)
		code += fmt.Sprintf("%s:\n", elseStart)
		for _, stmt := range n.If.ElseStmts {
			code += genNode(stmt, t)
		}
	}

	code += fmt.Sprintf("%s:\n", end)

	return code
}

func genWhile(n *ast.Node, t *symbol.Table) string {
	code := ""

	start := fmt.Sprintf(".L%d", localCount)
	localCount += 1
	end := fmt.Sprintf(".L%d", localCount)
	localCount += 1

	code += fmt.Sprintf("%s:\n", start)
	code += genNode(n.While.Exp, t)
	code += "	popq	%rax\n"
	code += "	cmpq	$0, %rax\n"
	code += fmt.Sprintf("	jz	%s\n", end)
	for _, stmt := range n.While.Stmts {
		code += genNode(stmt, t)
	}
	code += fmt.Sprintf("	jmp	%s\n", start)
	code += fmt.Sprintf("%s:\n", end)

	return code
}

func genFor(n *ast.Node, t *symbol.Table) string {
	code := ""

	start := fmt.Sprintf(".L%d", localCount)
	localCount += 1
	end := fmt.Sprintf(".L%d", localCount)
	localCount += 1

	code += genNode(n.For.Init, t)
	code += fmt.Sprintf("%s:\n", start)
	code += genNode(n.For.Cond, t)
	code += "	popq	%rax\n"
	code += "	cmpq	$0, %rax\n"
	code += fmt.Sprintf("	jz	%s\n", end)
	for _, stmt := range n.For.Stmts {
		code += genNode(stmt, t)
	}
	code += genNode(n.For.Adv, t)
	code += fmt.Sprintf("	jmp	%s\n", start)
	code += fmt.Sprintf("%s:\n", end)

	return code
}

func setVarOffsets(n *ast.Node, t *symbol.Table, reserv uint) uint {
	switch n.Tag {
	case ast.NodeScope:
		for _, stmt := range n.Scope.Stmts {
			reserv = setVarOffsets(stmt, t, reserv)
		}

	case ast.NodeLVarDecl:
		sym := t.Get(n.Id)
		size := types.Get(sym.Type).Size
		sym.LVar.Offset = reserv + size
		reserv += size
		t.Set(n.Id, sym)

	case ast.NodeBinOp:
		reserv = setVarOffsets(n.BinOp.Lval, t, reserv)
		reserv = setVarOffsets(n.BinOp.Rval, t, reserv)

	case ast.NodeFunCall:
		for _, arg := range n.Fun.Args {
			reserv = setVarOffsets(arg, t, reserv)
		}

	case ast.NodeIf:
		reserv = setVarOffsets(n.If.Exp, t, reserv)
		for _, stmt := range n.If.IfStmts {
			reserv = setVarOffsets(stmt, t, reserv)
		}
		for _, stmt := range n.If.ElseStmts {
			reserv = setVarOffsets(stmt, t, reserv)
		}

	case ast.NodeWhile:
		reserv = setVarOffsets(n.While.Exp, t, reserv)
		for _, stmt := range n.While.Stmts {
			reserv = setVarOffsets(stmt, t, reserv)
		}

	case ast.NodeFor:
		reserv = setVarOffsets(n.For.Init, t, reserv)
		reserv = setVarOffsets(n.For.Cond, t, reserv)
		reserv = setVarOffsets(n.For.Adv, t, reserv)
		for _, stmt := range n.For.Stmts {
			reserv = setVarOffsets(stmt, t, reserv)
		}

	case ast.NodeCast:
		reserv = setVarOffsets(n.Cast.What, t, reserv)

	case ast.NodeFunDef:
		for _, stmt := range n.Fun.Stmts {
			reserv = setVarOffsets(stmt, t, reserv)
		}

	case ast.NodeReturn:
		reserv = setVarOffsets(n.Return.Val, t, reserv)

	case ast.NodeInt:
	case ast.NodeLVar:
	case ast.NodeFunEx:
	case ast.NodeBool:
	case ast.NodeTypedef:
	case ast.NodeEmpty:

	default:
		panic("not implemented")
	}

	return reserv
}

func genFunction(n *ast.Node, t *symbol.Table) string {
	code := ""
	reserv := uint(0)

	code += "\n"
	code += t.Get(n.Id).Name + ":\n"
	code += "	pushq	%rbp\n"
	code += "	movq	%rsp, %rbp\n"

	for i, param := range n.Fun.Params {
		sym := t.Get(param)
		if sym.Tag != symbol.LVar {
			panic("param != local var")
		}

		typeNode := types.Get(sym.Type)
		size := typeNode.Size

		sym.LVar.Offset = reserv + size
		reserv += size
		t.Set(param, sym)

		code += fmt.Sprintf("	mov	%%%s, -%d(%%rbp)\n",
			argRegs[size][i], sym.LVar.Offset)
	}
	reserv = setVarOffsets(n, t, reserv)
	reserv += (16 - (reserv % 16)) % 16
	code += fmt.Sprintf("	subq	$%d, %%rsp\n", reserv)

	for _, stmt := range n.Fun.Stmts {
		code += genNode(stmt, t)
	}

	code += "	movq	%rbp, %rsp\n"
	code += "	popq	%rbp\n"
	code += "	ret\n"

	return code
}
