package main

import (
	"fmt"
	"net"

	"git.getcoffee.io/ottopress/WifiManager/darwin"
)

const (
	// IfaceConnected state indicates the itnerface is on and
	// connected.
	IfaceConnected WifiInterfaceStatus = iota
	// IfaceDisassociated state indicates the interface is on but
	// doesn't have an active connection
	IfaceDisassociated
	// IfaceOff state indicates the interface is off.
	IfaceOff
)

// WifiInterfaceStatus represents the state of interface
type WifiInterfaceStatus int

// WifiInterface represents a physical WiFi interface
type WifiInterface struct {
	net.Interface
	Name         string
	HardwareAddr string
	Status       bool
}

func main() {
	systemProfiler := darwin.NewSystemProfiler()
	airport := darwin.NewAirPort()
	networkSetup := darwin.NewNetworkSetup()
	fmt.Println("Network: ", networkSetup.IsInstalled())
	fmt.Println("SystemProfiler: ", systemProfiler.IsInstalled())
	fmt.Println("AirPort: ", airport.IsInstalled())
	fmt.Println("-------")
	fmt.Println(systemProfiler.Run(&networkSetup))
	fmt.Println(systemProfiler.Get("en1"))
	fmt.Println("-------")
	wifiList, wifiErr := airport.Scan()
	if wifiErr != nil {
		fmt.Println(wifiErr)
	}
	fmt.Println(wifiList[1].SSID)
}
