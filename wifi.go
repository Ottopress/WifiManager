package main

import (
	"fmt"

	"git.getcoffee.io/ottopress/WifiManager/darwin"
)

func main() {
	systemProfiler := darwin.NewSystemProfiler()
	airport := darwin.NewAirPort()
	fmt.Println(systemProfiler.IsInstalled())
	fmt.Println(systemProfiler.Run())
	fmt.Println(systemProfiler.Get("en1"))
	fmt.Println("-------")
	fmt.Println(airport.IsInstalled())
	fmt.Println(airport.Scan())
}
