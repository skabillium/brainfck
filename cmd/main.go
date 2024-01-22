package main

import (
	"fmt"
	"os"
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

type Lexer struct {
	pos    int
	source []byte
}

func NewLexer(source []byte) *Lexer {
	return &Lexer{pos: 0, source: source}
}

func (lex *Lexer) nextToken() int {
	for lex.pos < len(lex.source) && !strings.Contains(COMMANDS, string(lex.source[lex.pos])) {
		lex.pos++
	}

	if lex.pos >= len(lex.source) {
		return Eof
	}

	token := strings.Index(COMMANDS, string(lex.source[lex.pos]))
	lex.pos++
	return token
}

func (lex *Lexer) peek() int {
	pos := lex.pos
	token := lex.nextToken()
	lex.pos = pos

	return token
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
	command int // Operation type
	operand int // How many times to perform operation
}

func (op *Operation) print() {
	fmt.Println("Command:", string(COMMANDS[op.command]), "Operand:", op.operand)
}

func parse(lex Lexer) []Operation {
	ops := []Operation{}
	stack := []int{}

	for {
		token := lex.nextToken()
		if token == Eof {
			break
		}

		switch token {
		case RArrow, LArrow, Plus, Minus, Dot, Comma:
			streak := 1
			for {
				next := lex.peek()
				if next == token {
					streak++
					lex.nextToken()
				} else {
					break
				}
			}

			ops = append(ops, Operation{command: token, operand: streak})
		case LBracket:
			ops = append(ops, Operation{command: LBracket, operand: -1})
			stack = append(stack, len(ops)-1)
		case RBracket:
			last := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			ops = append(ops, Operation{command: RBracket, operand: last + 1})
			ops[last].operand = len(ops)
		}
	}

	return ops
}

func interpret(ops []Operation) {
	memory := make([]int, 32)
	head := 0
	ip := 0

	for ip < len(ops) {
		op := ops[ip]
		switch op.command {
		case Right:
			head += op.operand
			if head > MAX_MEMORY_SIZE {
				panic("Overflow")
			}
			if head > len(memory)-1 {
				memory = append(memory, make([]int, head-len(memory)+1)...)
			}
			ip++
		case Left:
			head -= op.operand
			if head < 0 {
				panic("Underflow")
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
					panic("Cannot convert to int")
				}
				memory[head] = val
			}
			ip++
		case JumpIfZero:
			// println("JumpIfZero")
			if memory[head] == 0 {
				ip = op.operand
			} else {
				ip++
			}
		case JumpIfNonZero:
			// println("JumpIfNonZero")
			if memory[head] != 0 {
				ip = op.operand
			} else {
				ip++
			}
		case Eof:
			break
		}

	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Expected file to interpret")
		return
	}
	filepath := os.Args[1]

	source, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Println("Could not find file:", filepath)
	}

	lexer := NewLexer(source)
	ops := parse(*lexer)

	// for _, op := range ops {
	// 	op.print()
	// }

	interpret(ops)
}
