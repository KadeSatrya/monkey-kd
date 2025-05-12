package parser

import (
	"fmt"
	"monkey_kd/ast"
	"monkey_kd/lexer"
	"monkey_kd/token"
	"strconv"
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

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
}

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

	// Prefix
	parse.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	parse.registerPrefix(token.IDENTIFIER, parse.parseIdentifier)
	parse.registerPrefix(token.INT, parse.parseIntegerLiteral)
	parse.registerPrefix(token.BANG, parse.parsePrefixExpression)
	parse.registerPrefix(token.MINUS, parse.parsePrefixExpression)
	parse.registerPrefix(token.TRUE, parse.parseBoolean)
	parse.registerPrefix(token.FALSE, parse.parseBoolean)
	parse.registerPrefix(token.LPAREN, parse.parseGroupedExpression)
	parse.registerPrefix(token.IF, parse.parseIfExpression)
	parse.registerPrefix(token.FUNCTION, parse.parseFunctionLiteral)

	// Infix
	parse.infixParseFns = make(map[token.TokenType]infixParseFn)
	parse.registerInfix(token.PLUS, parse.parseInfixExpression)
	parse.registerInfix(token.MINUS, parse.parseInfixExpression)
	parse.registerInfix(token.SLASH, parse.parseInfixExpression)
	parse.registerInfix(token.ASTERISK, parse.parseInfixExpression)
	parse.registerInfix(token.EQ, parse.parseInfixExpression)
	parse.registerInfix(token.NOT_EQ, parse.parseInfixExpression)
	parse.registerInfix(token.LT, parse.parseInfixExpression)
	parse.registerInfix(token.GT, parse.parseInfixExpression)
	parse.registerInfix(token.LPAREN, parse.parseCallExpression)

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
	parse.nextToken()
	stmt.Value = parse.parseExpression(LOWEST)
	for !parse.curTokenIs(token.SEMICOLON) {
		parse.nextToken()
	}
	return stmt
}

func (parse *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: parse.curToken}
	parse.nextToken()
	stmt.ReturnValue = parse.parseExpression(LOWEST)

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

func (parse *Parser) noPrefixParseFnError(tt token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", tt)
	parse.errors = append(parse.errors, msg)
}

func (parse *Parser) parseExpression(precedence int) ast.Expression {
	prefix := parse.prefixParseFns[parse.curToken.Type]
	if prefix == nil {
		parse.noPrefixParseFnError(parse.curToken.Type)
		return nil
	}

	leftExpression := prefix()

	for !parse.peekTokenIs(token.SEMICOLON) && precedence < parse.peekPrecedence() {
		infix := parse.infixParseFns[parse.peekToken.Type]
		if infix == nil {
			return leftExpression
		}
		parse.nextToken()
		leftExpression = infix(leftExpression)
	}
	return leftExpression
}

func (parse *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: parse.curToken}
	value, err := strconv.ParseInt(parse.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", parse.curToken.Literal)
		parse.errors = append(parse.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}

func (parse *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    parse.curToken,
		Operator: parse.curToken.Literal,
	}
	parse.nextToken()
	expression.Right = parse.parseExpression(PREFIX)
	return expression
}

func (parse *Parser) peekPrecedence() int {
	if parse, ok := precedences[parse.peekToken.Type]; ok {
		return parse
	}
	return LOWEST
}

func (parse *Parser) curPrecedence() int {
	if parse, ok := precedences[parse.curToken.Type]; ok {
		return parse
	}
	return LOWEST
}

func (parse *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    parse.curToken,
		Operator: parse.curToken.Literal,
		Left:     left,
	}
	precedence := parse.curPrecedence()
	parse.nextToken()
	expression.Right = parse.parseExpression(precedence)
	return expression
}

func (parse *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: parse.curToken, Value: parse.curTokenIs(token.TRUE)}
}

func (parse *Parser) parseGroupedExpression() ast.Expression {
	parse.nextToken()
	exp := parse.parseExpression(LOWEST)
	if !parse.expectPeek(token.RPAREN) {
		return nil
	}
	return exp
}

func (parse *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: parse.curToken}
	if !parse.expectPeek(token.LPAREN) {
		return nil
	}
	parse.nextToken()
	expression.Condition = parse.parseExpression(LOWEST)
	if !parse.expectPeek(token.RPAREN) {
		return nil
	}
	if !parse.expectPeek(token.LBRACE) {
		return nil
	}
	expression.Consequence = parse.parseBlockStatement()
	if parse.peekTokenIs(token.ELSE) {
		parse.nextToken()
		if !parse.expectPeek(token.LBRACE) {
			return nil
		}
		expression.Alternative = parse.parseBlockStatement()
	}
	return expression
}

func (parse *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: parse.curToken}
	if !parse.expectPeek(token.LPAREN) {
		return nil
	}
	lit.Parameters = parse.parseFunctionParameters()
	if !parse.expectPeek(token.LBRACE) {
		return nil
	}
	lit.Body = parse.parseBlockStatement()
	return lit
}

func (parse *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: parse.curToken}
	block.Statements = []ast.Statement{}
	parse.nextToken()
	for !parse.curTokenIs(token.RBRACE) && !parse.curTokenIs(token.EOF) {
		stmt := parse.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		parse.nextToken()
	}
	return block
}

func (parse *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}
	if parse.peekTokenIs(token.RPAREN) {
		parse.nextToken()
		return identifiers
	}
	parse.nextToken()
	ident := &ast.Identifier{Token: parse.curToken, Value: parse.curToken.Literal}
	identifiers = append(identifiers, ident)
	for parse.peekTokenIs(token.COMMA) {
		parse.nextToken()
		parse.nextToken()
		ident := &ast.Identifier{Token: parse.curToken, Value: parse.curToken.Literal}
		identifiers = append(identifiers, ident)
	}
	if !parse.expectPeek(token.RPAREN) {
		return nil
	}
	return identifiers
}

func (parse *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: parse.curToken, Function: function}
	exp.Arguments = parse.parseCallArguments()
	return exp
}

func (parse *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}
	if parse.peekTokenIs(token.RPAREN) {
		parse.nextToken()
		return args
	}
	parse.nextToken()
	args = append(args, parse.parseExpression(LOWEST))
	for parse.peekTokenIs(token.COMMA) {
		parse.nextToken()
		parse.nextToken()
		args = append(args, parse.parseExpression(LOWEST))
	}
	if !parse.expectPeek(token.RPAREN) {
		return nil
	}
	return args
}
