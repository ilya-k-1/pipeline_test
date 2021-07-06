package main

import (
	"bufio"
	"os"
	"os/exec"
	"fmt"
)


func get_data() string {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return text
}


func process_data(data string) string {
	if len(data) < 10 {
		return data + ">>>>>>>"
	} else {
		return "Ok"
	}
}


func parse_data(data string, txt string) string {
	cmd := exec.Command(txt)
	cmd.Run()
	return process_data(data)
}

func run_data(cmd string, txt string) {
	command := exec.Command(cmd)
	command.Run()
	fmt.Println(txt)
	if len(cmd) > 10 {
		fmt.Println(cmd)
	}
}


func execute_data(data string)  {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	run_data(data, text)
}

func main() {
	txt := get_data()
	txt2:= parse_data(txt, "Constant")
	execute_data(txt2)
}
