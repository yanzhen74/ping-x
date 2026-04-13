package main

import "github.com/yanzhen74/ping-x/cmd"

var Version = "dev"

func main() {
	cmd.SetVersion(Version)
	cmd.Execute()
}
