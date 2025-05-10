package main
import (
	"fmt"
	"os"
	"monkey_kd/repl"
)

func main() {
	fmt.Printf("Insert commands:\n")
	repl.Start(os.Stdin, os.Stdout)
}