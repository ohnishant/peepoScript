package evaluator

import (
	"fmt"

	"github.com/ohnishat/peepoScript/cmd/ast"
	"github.com/ohnishat/peepoScript/cmd/token"
)

var (
	NULL  = &Null{}
	TRUE  = &Boolean{Value: true}
	FALSE = &Boolean{Value: false}
)

func nativeBool(v bool) *Boolean {
	if v {
		return TRUE
	}
	return FALSE
}

func isError(obj Object) bool {
	return obj != nil && obj.Type() == ERROR_OBJ
}

func Eval(program *ast.Program, env *Environment) Object {
	var result Object = NULL

	for _, stmt := range program.Expressions {
		result = eval(stmt, env)
		if isError(result) {
			return result
		}
	}
	return result
}

func eval(expr ast.Expression, env *Environment) Object {
	switch e := expr.(type) {

	case *ast.LetExpression:
		name, ok := e.Variable.(*ast.Identifier)
		if !ok {
			return &Error{Message: "PepoG needs a name"}
		}
		val := eval(e.AssignValue, env)
		if isError(val) {
			return val
		}
		return env.Set(name.Value, val)

	case *ast.IntegerLiteral:
		return &Integer{Value: e.Value}

	case *ast.StringLiteral:
		return &String{Value: e.Value}

	case *ast.BooleanLiteral:
		return nativeBool(e.Value)

	case *ast.ArrayLiteral:
		elements := make([]Object, 0, len(e.Elements))
		for _, elem := range e.Elements {
			v := eval(elem, env)
			if isError(v) {
				return v
			}
			elements = append(elements, v)
		}
		return &Array{Elements: elements}

	case *ast.Identifier:
		val, ok := env.Get(e.Value)
		if !ok {
			return notFound(e.Value)
		}
		return val

	case *ast.PrefixExpression:
		return evalPrefix(e.Hunks, env)

	case *ast.IfExpression:
		cond := eval(e.Condition, env)
		if isError(cond) {
			return cond
		}
		b, ok := cond.(*Boolean)
		if !ok {
			return &Error{Message: fmt.Sprintf("Hmmge condition must be NODDERS/NOPERS, got %s", cond.Type())}
		}
		if b.Value {
			return Eval(e.Consequence, NewEnclosedEnvironment(env))
		}
		if e.Alternative != nil {
			return Eval(e.Alternative, NewEnclosedEnvironment(env))
		}
		return NULL

	case *ast.FunctionLiteral:
		return &Function{Parameters: e.Parameters, Body: e.Body, Env: env}

	case *ast.ForExpression:
		return evalFor(e, env)
	}

	return &Error{Message: fmt.Sprintf("unknown expression: %T", expr)}
}

// evalFor runs a peepoJuice loop. The header clauses and the loop variable
// live in one enclosed environment for the whole loop; each body iteration
// gets a fresh child of it, mirroring how Hmmge scopes its blocks. The loop
// itself evaluates to peepoSilence.
func evalFor(node *ast.ForExpression, env *Environment) Object {
	loopEnv := NewEnclosedEnvironment(env)

	if node.Init != nil {
		v := eval(node.Init, loopEnv)
		if isError(v) {
			return v
		}
	}

	for {
		cond := eval(node.Condition, loopEnv)
		if isError(cond) {
			return cond
		}
		b, ok := cond.(*Boolean)
		if !ok {
			return &Error{Message: fmt.Sprintf("peepoJuice condition must be NODDERS/NOPERS, got %s", cond.Type())}
		}
		if !b.Value {
			return NULL
		}

		Eval(node.Body, NewEnclosedEnvironment(loopEnv))

		if node.Step != nil {
			v := eval(node.Step, loopEnv)
			if isError(v) {
				return v
			}
		}
	}
}

// evalPrefix resolves a flat operator-first stream. Operators consume their
// operands by recursion; an identifier bound to something callable consumes
// as many operands as its arity demands. That is the whole trick behind not
// needing Wokege and Bedge around every expression. Index hunks
// (Thinking1 ... Thinking2) apply to whatever the preceding operand
// evaluated to.
func evalPrefix(hunks []ast.Expression, env *Environment) Object {
	pos := 0

	var primary func() Object
	var next func() Object

	// next evaluates one operand and then applies any index hunks that
	// trail it, so `xs Thinking1 0 Thinking2` chains like a postfix.
	next = func() Object {
		v := primary()
		if isError(v) {
			return v
		}
		for pos < len(hunks) {
			idxNode, ok := hunks[pos].(*ast.IndexNode)
			if !ok {
				break
			}
			pos++
			idx := eval(idxNode.Index, env)
			if isError(idx) {
				return idx
			}
			v = applyIndex(v, idx)
			if isError(v) {
				return v
			}
		}
		return v
	}

	primary = func() Object {
		if pos >= len(hunks) {
			return &Error{Message: "expression ran out of operands"}
		}
		switch h := hunks[pos].(type) {

		case *ast.IntegerLiteral:
			pos++
			return &Integer{Value: h.Value}

		case *ast.StringLiteral:
			pos++
			return &String{Value: h.Value}

		case *ast.BooleanLiteral:
			pos++
			return nativeBool(h.Value)

		case *ast.ArrayLiteral:
			pos++
			return eval(h, env)

		case *ast.OperatorNode:
			op := h.Operator
			pos++
			if h.Token.Type == token.NEGATE {
				val := next()
				if isError(val) {
					return val
				}
				i, ok := val.(*Integer)
				if !ok {
					return &Error{Message: fmt.Sprintf("Sadge... unary - needs an integer, got %s", val.Type())}
				}
				return &Integer{Value: -i.Value}
			}
			left := next()
			if isError(left) {
				return left
			}
			right := next()
			if isError(right) {
				return right
			}
			return applyInfix(op, left, right)

		case *ast.Identifier:
			name := h.Value
			pos++
			callee, ok := env.Get(name)
			if !ok {
				return notFound(name)
			}
			switch fn := callee.(type) {
			case *Builtin:
				var args []Object
				if fn.Arity < 0 {
					for pos < len(hunks) {
						arg := next()
						if isError(arg) {
							return arg
						}
						args = append(args, arg)
					}
					if len(args) == 0 {
						return &Error{Message: fmt.Sprintf("%s needs at least one operand", name)}
					}
				} else {
					args = make([]Object, fn.Arity)
					for i := range args {
						arg := next()
						if isError(arg) {
							return arg
						}
						args[i] = arg
					}
				}
				return fn.Fn(args...)
			case *Function:
				args := make([]Object, len(fn.Parameters))
				for i := range args {
					arg := next()
					if isError(arg) {
						return arg
					}
					args[i] = arg
				}
				callEnv := NewEnclosedEnvironment(fn.Env)
				for i, param := range fn.Parameters {
					callEnv.Set(param.Value, args[i])
				}
				return Eval(fn.Body, callEnv)
			default:
				return callee
			}

		default:
			return &Error{Message: fmt.Sprintf("unknown expression: %T", hunks[pos])}
		}
	}

	result := next()
	if isError(result) {
		return result
	}

	if pos < len(hunks) {
		return &Error{Message: fmt.Sprintf("%d trailing operand(s) never got used", len(hunks)-pos)}
	}
	return result
}

func applyInfix(op string, left, right Object) Object {
	switch op {
	case "+":
		switch l := left.(type) {
		case *Integer:
			r, ok := right.(*Integer)
			if !ok {
				return typeError(op, l.Type(), right.Type())
			}
			return &Integer{Value: l.Value + r.Value}
		case *String:
			r, ok := right.(*String)
			if !ok {
				return typeError(op, l.Type(), right.Type())
			}
			return &String{Value: l.Value + r.Value}
		default:
			return typeError(op, left.Type(), right.Type())
		}

	case "-", "*", "/":
		l, ok := left.(*Integer)
		if !ok {
			return typeError(op, left.Type(), right.Type())
		}
		r, ok := right.(*Integer)
		if !ok {
			return typeError(op, left.Type(), right.Type())
		}
		switch op {
		case "-":
			return &Integer{Value: l.Value - r.Value}
		case "*":
			return &Integer{Value: l.Value * r.Value}
		default:
			if r.Value == 0 {
				return &Error{Message: "division by zero"}
			}
			return &Integer{Value: l.Value / r.Value}
		}

	case "<", ">":
		l, ok := left.(*Integer)
		if !ok {
			return typeError(op, left.Type(), right.Type())
		}
		r, ok := right.(*Integer)
		if !ok {
			return typeError(op, left.Type(), right.Type())
		}
		if op == "<" {
			return nativeBool(l.Value < r.Value)
		}
		return nativeBool(l.Value > r.Value)

	case "==":
		return equals(left, right)
	}

	return &Error{Message: fmt.Sprintf("unknown operator: %s", op)}
}

// applyIndex resolves one Thinking1 index Thinking2 against a target.
func applyIndex(target, index Object) Object {
	arr, ok := target.(*Array)
	if !ok {
		return &Error{Message: fmt.Sprintf("indexing needs a list, got %s", target.Type())}
	}
	i, ok := index.(*Integer)
	if !ok {
		return &Error{Message: fmt.Sprintf("list index must be an integer, got %s", index.Type())}
	}
	if i.Value < 0 || i.Value >= int64(len(arr.Elements)) {
		return &Error{Message: fmt.Sprintf("index out of range: %d, list length is %d", i.Value, len(arr.Elements))}
	}
	return arr.Elements[i.Value]
}

func equals(left, right Object) Object {
	if left.Type() != right.Type() {
		return FALSE
	}
	switch l := left.(type) {
	case *Integer:
		return nativeBool(l.Value == right.(*Integer).Value)
	case *String:
		return nativeBool(l.Value == right.(*String).Value)
	case *Boolean:
		return nativeBool(l.Value == right.(*Boolean).Value)
	}
	return FALSE
}

func typeError(op string, left, right ObjectType) *Error {
	return &Error{Message: fmt.Sprintf("PepegaCredit on %s needs matching operands, got %s and %s", op, left, right)}
}
