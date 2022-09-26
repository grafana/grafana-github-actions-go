package main

import (
	"fmt"
	"os"
)

func main() {
	// we need to get all open issue with milestone and remove the milestone from them
	// we need to get all PR opened with milestone and remove the milestone from them
	fmt.Println("Make it workkkkk")
	// Just using something simple to dmeonstrate using the github package here
	argsWithoutProg := os.Args[1:]
	fmt.Println(argsWithoutProg)
}
