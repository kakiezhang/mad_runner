package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func writeFile(s string, f string) {
	buf := []byte(s)
	err := ioutil.WriteFile(f, buf, 0644)
	if err != nil {
		fmt.Printf("Write file failed: file[ %s ]\n", f)
		panic(err)
	}
}

func appendFile(s string, f string) {
	fd, err := os.OpenFile(f, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("Append file failed: file[ %s ]\n", f)
		panic(err)
	}
	defer fd.Close()
	buf := []byte(s)
	fd.Write(buf)
}

func readFile(path string) string {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("Read file failed: path[ %s ]\n", path)
		panic(err)
	}
	return string(f)
}

func readLines(path string) []string {
	s := readFile(path)
	lines := strings.Split(s, "\n")
	return lines
}
