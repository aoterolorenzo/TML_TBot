package main

import (
	"TML_TBot/cmd"
	"TML_TBot/config"
	"fmt"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			config.Log.Error(fmt.Sprintf("Panic!: %+v\n", r))
			main()
		}
	}()

	err := cmd.Execute()
	if err != nil {
		return
	}
}
