package main

import (
	"io"
	"os"
	"reflect"
	"testing"
)

// Test Lexer
func TestLexer(t *testing.T) {
	source := "><+-.,[]"
	lex := NewLexer([]byte(source))

	expected := [8]int{
		RArrow, LArrow, Plus,
		Minus, Dot, Comma,
		LBracket, RBracket,
	}

	i := 0
	for {
		token := lex.nextToken()
		if token.kind == Eof {
			break
		}

		if token.kind != expected[i] {
			t.Fatalf("Expected '%s', received %d", string(source[i]), token.kind)
		}

		i++
	}
}

// Test Parser
func TestParser(t *testing.T) {
	source := "...><<+"
	lex := NewLexer([]byte(source))

	expected := []Operation{
		{Output, 3},
		{Right, 1},
		{Left, 2},
		{Incr, 1},
	}

	ops, err := parse(*lex)

	if !reflect.DeepEqual(ops, expected) || err != nil {
		t.Fatal("Expected other than received")
	}
}

// Test Interpreter
func TestInterpreter(t *testing.T) {
	helloWorld := "++++++++[>++++[>++>+++>+++>+<<<<-]>+>+>->>+[<]<-]>>.>---.+++++++..+++.>>.<-.<.+++.------.--------.>>+.>++."

	lex := NewLexer([]byte(helloWorld))
	ops, _ := parse(*lex)

	// Capture Stdout
	stdOut := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	interpret(ops)

	_ = w.Close()
	result, _ := io.ReadAll(r)
	out := string(result)
	os.Stdout = stdOut

	expected := "Hello World!\n"

	if out != expected {
		t.Fatalf("Expected '%s' received '%s'", expected, out)
	}

}

// GetFilename
func TestGetFilename(t *testing.T) {
	paths := []string{
		// Files
		"main.go",
		// Relative
		"./main.go",
		// Unix
		"/path/to/main.go",
		// Windows
		"C:\\\\Users\\main.go",
	}

	for _, path := range paths {
		file := getFilename(path)
		if file != "main.go" {
			t.Fatalf("Failed case for path %s, expected %s", path, file)
		}
	}
}

// GetPosition
func TestGetPosition(t *testing.T) {
	source := `...
...
...`

	line, col := getPosition([]byte(source), 1)
	if line != 1 || col != 2 {
		t.Fatalf("Expected line %d, column %d. Received line %d, column %d", 1, 2, line, col)
	}

	line, col = getPosition([]byte(source), 10)
	if line != 3 || col != 3 {
		t.Fatalf("Expected line %d, column %d. Received line %d, column %d", 3, 2, line, col)
	}
}
