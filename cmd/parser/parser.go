package parser

import (
	"fmt"
	"strconv"

	"github.com/ohnishat/peepoScript/cmd/ast"
	"github.com/ohnishat/peepoScript/cmd/lexer"
	"github.com/ohnishat/peepoScript/cmd/token"
)

// Parser for peepoScript. Expressions are operator-first (Polish notation),
// so there is no precedence table and no parens: the shape of the token
// stream alone determines the tree. Calls consume arguments greedily until a
// token that cannot start an expression.
type Parser struct {
	l         *lexer.Lexer
	curToken  token.Token
	peekToken token.Token
	pos       int // number of tokens consumed; identity for progress checks
	errors    []string
}

func Parse(input string) (*ast.Program, []string) {
	p := &Parser{l: lexer.New(input)}
	p.nextToken()
	p.nextToken()
	return p.parseProgram(), p.errors
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
	p.pos++
}

// stuck reports whether no token was consumed since mark. Token structs are
// compared by value, so repeated keywords like PepoG PepoG would look like
// "no progress" if we compared tokens directly.
func (p *Parser) stuck(mark int) bool {
	return p.pos == mark
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.errors = append(p.errors, fmt.Sprintf("Sadge... expected %s, got %s instead", t, p.peekToken.Type))
	return false
}

func (p *Parser) parseProgram() *ast.Program {
	program := &ast.Program{}
	for !p.curTokenIs(token.EOF) {
		mark := p.pos
		stmt := p.parseStatement()
		if stmt != nil {
			program.Expressions = append(program.Expressions, stmt)
		}
		if p.stuck(mark) { // guarantee progress on errors
			p.nextToken()
		}
	}
	return program
}

func (p *Parser) parseStatement() ast.Expression {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLet()
	case token.ASSIGN:
		return p.parseAssign()
	case token.IF:
		return p.parseIf()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLet() ast.Expression {
	letTok := p.curToken

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	name := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if p.peekTokenIs(token.FUNCTION) {
		p.nextToken() // SadgeBusiness
		fn := p.parseFunctionLiteral()
		if fn == nil {
			return nil
		}
		return &ast.LetExpression{Token: letTok, Variable: name, AssignValue: fn}
	}

	p.nextToken()
	val := p.parseExpression()
	if val == nil {
		return nil
	}
	p.acceptFullstop()
	return &ast.LetExpression{Token: letTok, Variable: name, AssignValue: val}
}

func (p *Parser) parseAssign() ast.Expression {
	tok := p.curToken

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	name := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	p.nextToken()
	val := p.parseExpression()
	if val == nil {
		return nil
	}
	p.acceptFullstop()

	let := &ast.LetExpression{Token: tok, Variable: name, AssignValue: val}
	return let
}

func (p *Parser) parseIf() ast.Expression {
	ifTok := p.curToken

	p.nextToken()
	cond := p.parseExpression()
	if cond == nil {
		return nil
	}
	// parseExpression leaves cur on the first token that cannot continue
	// the expression, which here must be Wokege.
	if !p.curTokenIs(token.LBRACE) {
		p.errors = append(p.errors, fmt.Sprintf("Sadge... expected Wokege after Hmmge condition, got %s", p.curToken.Type))
		return nil
	}
	consequence := p.parseBlockProgram()

	var alternative *ast.Program
	if p.curTokenIs(token.ELSE) {
		p.nextToken()
		if !p.curTokenIs(token.LBRACE) {
			p.errors = append(p.errors, fmt.Sprintf("Sadge... expected Wokege after peepoShrug, got %s", p.curToken.Type))
			return nil
		}
		alternative = p.parseBlockProgram()
	}

	return &ast.IfExpression{
		Token:       ifTok,
		Condition:   cond,
		Consequence: consequence,
		Alternative: alternative,
	}
}

// parseBlockProgram parses statements between Wokege ... Bedge and consumes
// the closing Bedge.
func (p *Parser) parseBlockProgram() *ast.Program {
	block := &ast.Program{}
	p.nextToken() // move past Wokege

	for !p.curTokenIs(token.RBRACE) {
		if p.curTokenIs(token.EOF) {
			p.errors = append(p.errors, "Sadge... missing Bedge before end of input")
			return block
		}
		mark := p.pos
		stmt := p.parseStatement()
		if stmt != nil {
			block.Expressions = append(block.Expressions, stmt)
		}
		if p.stuck(mark) { // guarantee progress on errors
			p.errors = append(p.errors, fmt.Sprintf("Sadge... unexpected %q", p.curToken.Literal))
			p.nextToken()
		}
	}
	p.nextToken() // consume Bedge
	return block
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	fnTok := p.curToken // SadgeBusiness

	params := []*ast.Identifier{}
	for p.peekTokenIs(token.IDENT) || p.peekTokenIs(token.COMMA) {
		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
			continue
		}
		p.nextToken()
		params = append(params, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	body := p.parseBlockProgram()

	return &ast.FunctionLiteral{Token: fnTok, Parameters: params, Body: body}
}

func (p *Parser) parseExpressionStatement() ast.Expression {
	expr := p.parseExpression()
	if expr == nil {
		p.nextToken() // guarantee progress after a failed expression
		return nil
	}
	p.acceptFullstop()
	return expr
}

// acceptFullstop consumes the statement-terminating fullstop when present.
// parseExpression leaves cur on the first token that cannot continue the
// expression, so that is where we look. A missing one is only tolerated at
// EOF or right before Bedge, so blocks can end with a bare `NODDERS`.
func (p *Parser) acceptFullstop() {
	if p.curTokenIs(token.FULLSTOP) {
		p.nextToken()
		return
	}
	if !p.curTokenIs(token.EOF) && !p.curTokenIs(token.RBRACE) {
		p.errors = append(p.errors, fmt.Sprintf("Sadge... expected ., got %q instead", p.curToken.Literal))
	}
}

func isExprStart(t token.TokenType) bool {
	switch t {
	case token.INT, token.STRING, token.TRUE, token.FALSE,
		token.IDENT, token.NOT,
		token.PLUS, token.MINUS, token.MULTIPLY, token.DIVIDE,
		token.EQUAL, token.LESSTHAN, token.GREATERTHAN:
		return true
	}
	return false
}

var operators = map[token.TokenType]string{
	token.PLUS:        "+",
	token.MINUS:       "-",
	token.MULTIPLY:    "*",
	token.DIVIDE:      "/",
	token.EQUAL:       "==",
	token.LESSTHAN:    "<",
	token.GREATERTHAN: ">",
	token.NOT:         "!",
}

// parseExpression collects a flat operator-first stream of hunks until a
// token that cannot continue an expression. Grouping is left to the
// evaluator, which knows the arity of whatever each identifier is bound to.
func (p *Parser) parseExpression() ast.Expression {
	prefix := &ast.PrefixExpression{Token: p.curToken}
	count := 0

	for isExprStart(p.curToken.Type) {
		switch p.curToken.Type {
		case token.INT:
			value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
			if err != nil {
				p.errors = append(p.errors, fmt.Sprintf("Sadge... could not parse %q as integer", p.curToken.Literal))
				return nil
			}
			prefix.Hunks = append(prefix.Hunks, &ast.IntegerLiteral{Token: p.curToken, Value: value})

		case token.STRING:
			prefix.Hunks = append(prefix.Hunks, &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal})

		case token.TRUE, token.FALSE:
			prefix.Hunks = append(prefix.Hunks, &ast.BooleanLiteral{
				Token: p.curToken,
				Value: p.curToken.Type == token.TRUE,
			})

		case token.IDENT:
			prefix.Hunks = append(prefix.Hunks, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})

		case token.NOT:
			prefix.Hunks = append(prefix.Hunks, &ast.OperatorNode{Token: p.curToken, Operator: "!"})

		case token.PLUS, token.MINUS, token.MULTIPLY, token.DIVIDE,
			token.EQUAL, token.LESSTHAN, token.GREATERTHAN:
			prefix.Hunks = append(prefix.Hunks, &ast.OperatorNode{
				Token:    p.curToken,
				Operator: operators[p.curToken.Type],
			})
		}
		p.nextToken()
		count++
	}

	if count == 0 {
		p.errors = append(p.errors, fmt.Sprintf("Sadge... unexpected %q in expression", p.curToken.Literal))
		return nil
	}
	return prefix
}
