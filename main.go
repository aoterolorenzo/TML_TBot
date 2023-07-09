package main

import "TML_TBot/cmd"

func main() {
	err := cmd.Execute()
	if err != nil {
		return
	}
}
