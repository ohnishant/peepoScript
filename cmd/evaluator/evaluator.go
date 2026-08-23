package evaluator

import (
	"fmt"

	"github.com/ohnishat/peepoScript/cmd/ast"
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
	}

	return &Error{Message: fmt.Sprintf("unknown expression: %T", expr)}
}

// evalPrefix resolves a flat operator-first stream. Operators consume their
// operands by recursion; an identifier bound to something callable consumes
// as many operands as its arity demands. That is the whole trick behind not
// needing Wokege and Bedge around every expression.
func evalPrefix(hunks []ast.Expression, env *Environment) Object {
	pos := 0

	var next func() Object
	next = func() Object {
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

		case *ast.OperatorNode:
			op := h.Operator
			pos++
			if op == "!" {
				val := next()
				if isError(val) {
					return val
				}
				b, ok := val.(*Boolean)
				if !ok {
					return &Error{Message: fmt.Sprintf("peepoJuice needs a boolean, got %s", val.Type())}
				}
				return nativeBool(!b.Value)
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

	// peepoJuice is allowed to trail its operand: `NODDERS peepoJuice`.
	for pos < len(hunks) {
		op, ok := hunks[pos].(*ast.OperatorNode)
		if !ok || op.Operator != "!" {
			return &Error{Message: fmt.Sprintf("%d trailing operand(s) never got used", len(hunks)-pos)}
		}
		pos++
		b, ok := result.(*Boolean)
		if !ok {
			return &Error{Message: fmt.Sprintf("peepoJuice needs a boolean, got %s", result.Type())}
		}
		result = nativeBool(!b.Value)
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
