package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {

	// #read file info
	//info, err := os.Lstat("./data/data_test.go")
	//if err != nil {
	//	log.Fatal("can't open file", err)
	//}
	////	pretty.Println(info)
	//mod := info.ModTime()
	//fmt.Println(mod.Unix())
	//fmt.Println(mod.UnixNano())

	// #store info

	// #check changing

	// #run command
	args := os.Args
	prog := args[1]
	params := args[2:]
	fmt.Println(prog)
	fmt.Println(params)

	cmd := exec.Command(prog, params...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	fmt.Println(err)
}
