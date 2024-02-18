package main

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
)

const COMMANDS = "><+-.,[]"
const MAX_MEMORY_SIZE = 30_000

// Token types
const (
	RArrow = iota
	LArrow
	Plus
	Minus
	Dot
	Comma
	LBracket
	RBracket
	Eof
)

type Token struct {
	kind int
	pos  int
}

type Lexer struct {
	source []byte
	pos    int
	lines  int
}

func NewLexer(source []byte) *Lexer {
	return &Lexer{source: source, pos: 0, lines: 0}
}

func (lex *Lexer) nextToken() Token {
	for lex.pos < len(lex.source) && !strings.Contains(COMMANDS, string(lex.source[lex.pos])) {
		if lex.source[lex.pos] == '\n' {
			lex.lines++
		}
		lex.pos++
	}

	if lex.pos >= len(lex.source) {
		return Token{kind: Eof, pos: len(lex.source)}
	}

	kind := strings.Index(COMMANDS, string(lex.source[lex.pos]))
	lex.pos++
	return Token{kind: kind, pos: lex.pos - 1}
}

// Operation types
const (
	Right = iota
	Left
	Incr
	Decr
	Output
	Input
	JumpIfZero
	JumpIfNonZero
)

type Operation struct {
	command int
	operand int
}

func getPosition(source []byte, offset int) (int, int) {
	// Count new lines up to that position
	line := 1
	column := 1
	for i := 0; i < offset; i++ {
		if source[i] == '\n' {
			line++
			column = 1
			continue
		}
		column++
	}
	return line, column
}

func parse(lex Lexer) ([]Operation, error) {
	ops := []Operation{}
	stack := []int{}

	token := lex.nextToken()
	for {
		if token.kind == Eof {
			break
		}

		switch token.kind {
		case RArrow, LArrow, Plus, Minus, Dot, Comma:
			streak := 1
			next := lex.nextToken()
			for next.kind == token.kind {
				streak++
				next = lex.nextToken()
			}

			ops = append(ops, Operation{command: token.kind, operand: streak})
			token = next
		case LBracket:
			ops = append(ops, Operation{command: LBracket, operand: -1})
			stack = append(stack, len(ops)-1)
			token = lex.nextToken()
		case RBracket:
			if len(stack) == 0 {
				line, col := getPosition(lex.source, token.pos)
				return nil, fmt.Errorf("%d:%d Loop mismatch", line, col)
			}
			last := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			ops = append(ops, Operation{command: RBracket, operand: last + 1})
			ops[last].operand = len(ops)
			token = lex.nextToken()
		}
	}

	return ops, nil
}

func interpret(ops []Operation) error {
	memory := make([]int, 32)
	head := 0
	ip := 0

run:
	for ip < len(ops) {
		op := ops[ip]
		switch op.command {
		case Right:
			head += op.operand
			if head > MAX_MEMORY_SIZE {
				return fmt.Errorf("memory overflow")

			}
			if head > len(memory)-1 {
				memory = append(memory, make([]int, head-len(memory)+1)...)
			}
			ip++
		case Left:
			head -= op.operand
			if head < 0 {
				return fmt.Errorf("memory underflow")
			}
			ip++
		case Incr:
			memory[head] += op.operand
			ip++
		case Decr:
			memory[head] -= op.operand
			ip++
		case Output:
			for i := 0; i < op.operand; i++ {
				fmt.Printf("%c", memory[head])
			}
			ip++
		case Input:
			for i := 0; i < op.operand; i++ {
				fmt.Println("Input:")
				var input string
				fmt.Scanln(&input)
				val, err := strconv.Atoi(input)
				if err != nil {
					return fmt.Errorf(fmt.Sprintf("Cannot convert \"%s\" to int", input))
				}
				memory[head] = val
			}
			ip++
		case JumpIfZero:
			if memory[head] == 0 {
				ip = op.operand
			} else {
				ip++
			}
		case JumpIfNonZero:
			if memory[head] != 0 {
				ip = op.operand
			} else {
				ip++
			}
		case Eof:
			break run
		}
	}

	return nil
}

func getFilename(filepath string) string {
	i := len(filepath) - 1
	for i > 0 {
		if filepath[i] == '\\' || filepath[i] == '/' {
			i++
			break
		}
		i--
	}

	return filepath[i:]
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Expected file to interpret")
		return
	}
	filepath := os.Args[1]

	source, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Println(err)
	}

	lexer := NewLexer(source)
	ops, err := parse(*lexer)
	if err != nil {
		base := path.Base(filepath)
		fmt.Printf("%s:%s\n", base, err)
	}

	err = interpret(ops)
	if err != nil {
		fmt.Println("Runtime Error:", err)
	}
}
