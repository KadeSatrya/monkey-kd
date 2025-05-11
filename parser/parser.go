package parser

import (
	"fmt"
	"monkey_kd/ast"
	"monkey_kd/lexer"
	"monkey_kd/token"
)

const (
	_ int = iota
	LOWEST
	EQUALS
	LESSGREATER
	SUM
	PRODUCT
	PREFIX
	CALL
)

type Parser struct {
	lex            *lexer.Lexer
	curToken       token.Token
	peekToken      token.Token
	errors         []string
	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(lex *lexer.Lexer) *Parser {
	parse := &Parser{
		lex:    lex,
		errors: []string{},
	}

	parse.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	parse.registerPrefix(token.IDENTIFIER, parse.parseIdentifier)

	// For setting current and peek token
	parse.nextToken()
	parse.nextToken()

	return parse
}

func (parse *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: parse.curToken, Value: parse.curToken.Literal}
}

func (parse *Parser) nextToken() {
	parse.curToken = parse.peekToken
	parse.peekToken = parse.lex.NextToken()
}

func (parse *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}
	for parse.curToken.Type != token.EOF {
		stmt := parse.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		parse.nextToken()
	}
	return program
}

func (parse *Parser) parseStatement() ast.Statement {
	switch parse.curToken.Type {
	case token.LET:
		return parse.parseLetStatement()
	case token.RETURN:
		return parse.parseReturnStatement()
	default:
		return parse.parseExpressionStatement()
	}
}

func (parse *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: parse.curToken}
	if !parse.expectPeek(token.IDENTIFIER) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: parse.curToken, Value: parse.curToken.Literal}
	if !parse.expectPeek(token.ASSIGN) {
		return nil
	}
	for !parse.curTokenIs(token.SEMICOLON) {
		parse.nextToken()
	}
	return stmt
}

func (parse *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: parse.curToken}
	parse.nextToken()
	for !parse.curTokenIs(token.SEMICOLON) {
		parse.nextToken()
	}
	return stmt
}

func (parse *Parser) curTokenIs(tok token.TokenType) bool {
	return parse.curToken.Type == tok
}
func (parse *Parser) peekTokenIs(tok token.TokenType) bool {
	return parse.peekToken.Type == tok
}
func (parse *Parser) expectPeek(tok token.TokenType) bool {
	if parse.peekTokenIs(tok) {
		parse.nextToken()
		return true
	} else {
		parse.peekError(tok)
		return false
	}
}

func (parse *Parser) Errors() []string {
	return parse.errors
}

func (parse *Parser) peekError(tok token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		tok, parse.peekToken.Type)
	parse.errors = append(parse.errors, msg)
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

func (parse *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	parse.prefixParseFns[tokenType] = fn
}
func (parse *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	parse.infixParseFns[tokenType] = fn
}

func (parse *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: parse.curToken}
	stmt.Expression = parse.parseExpression(LOWEST)
	if parse.peekTokenIs(token.SEMICOLON) {
		parse.nextToken()
	}
	return stmt
}

func (parse *Parser) parseExpression(precedence int) ast.Expression {
	prefix := parse.prefixParseFns[parse.curToken.Type]
	if prefix == nil {
		return nil
	}
	leftExpression := prefix()
	return leftExpression
}
