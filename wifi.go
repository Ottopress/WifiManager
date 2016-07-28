package main

import (
	"fmt"

	"git.getcoffee.io/ottopress/WifiManager/darwin"
)

func main() {
	systemProfiler := darwin.NewSystemProfiler()
	fmt.Println(systemProfiler.IsInstalled())
	fmt.Println(systemProfiler.Run())
}
