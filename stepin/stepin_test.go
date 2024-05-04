package stepin

import (
	"log"
	"os"
	"testing"
)

func TestExec(t *testing.T) {
	output, err := Exec("echo", "-n", "hello")
	if err != nil {
		t.Fatal(err)
	}
	if output != "hello" {
		t.Fatalf("unexpected output: %s", output)
	}
	output, err = Exec("echo", "world")
	if err != nil {
		t.Fatal(err)
	}
	if output != "world\n" {
		t.Fatalf("unexpected output: %s", output)
	}
}

func TestNewTmpFile(t *testing.T) {
	f, dispose, err := NewTmpFile("test", []byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = dispose()
	}()
	log.Println(f.Name())

	content, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "hello" {
		t.Fatalf("unexpected content: %s", content)
	}
	err = dispose()
	if err != nil {
		t.Fatal(err)
	}
	_, err = os.Stat(f.Name())
	if !os.IsNotExist(err) {
		t.Fatalf("file %s should be removed", f.Name())
	}
}
