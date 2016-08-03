package wifimanager

import (
	"errors"
	"net"

	"git.getcoffee.io/ottopress/wifimanager/darwin"
)

const (
	// IfaceConnected state indicates the itnerface is on and
	// connected.
	IfaceConnected int = iota
	// IfaceDisassociated state indicates the interface is on but
	// doesn't have an active connection
	IfaceDisassociated
	// IfaceOff state indicates the interface is off.
	IfaceOff
	// SecurityWPA represents the WPA WiFi security protocol
	SecurityWPA int = iota
	// SecurityWEP represents the WEP WiFi security protocol
	SecurityWEP
	// SecurityWPA2 represents the WPA2/RSN WiFi security protocol
	SecurityWPA2
	// SecurityNone represents the lack of any WiFi security protocol
	SecurityNone
)

var (
	airport        = darwin.NewAirPort()
	networkSetup   = darwin.NewNetworkSetup()
	systemProfiler = darwin.NewSystemProfiler()
)

// WifiInterface represents a physical WiFi interface
type WifiInterface struct {
	net.Interface
	Model      string
	Vendor     string
	Connection WifiNetwork
}

// WifiNetwork represents a discovered WiFi network
type WifiNetwork struct {
	SSID         string
	Channel      int
	SecurityType int
	SecurityKey  string
	QualityLevel int
	NoiseLevel   int
}

// GetWifiInterfaces returns a list of all active Wifi interfaces
func GetWifiInterfaces() ([]WifiInterface, error) {
	wifiInterfaces := []WifiInterface{}

	netInterfaces, netErr := net.Interfaces()
	if netErr != nil {
		return nil, netErr
	}

	systemProfiler.Run(networkSetup)
	for _, iface := range netInterfaces {
		_, spErr := systemProfiler.Get(iface.Name)
		if spErr == nil {
			wifiInterface, _ := NewWifiInterface(iface)
			wifiInterfaces = append(wifiInterfaces, wifiInterface)
		}
	}

	if len(wifiInterfaces) < 1 {
		return nil, errors.New("wifi: no wifi interfaces found")
	}
	return wifiInterfaces, nil
}

// NewWifiInterface builds a WifiInterface instance off of the
// "net" package's interface.
func NewWifiInterface(iface net.Interface) (WifiInterface, error) {
	wifiInterface := WifiInterface{Interface: iface}

	spInfo, spErr := systemProfiler.Get(iface.Name)
	if spErr != nil {
		return WifiInterface{}, spErr
	}
	wifiInterface.Model = spInfo.ID
	wifiInterface.MTU = spInfo.MTU
	wifiInterface.Vendor = spInfo.Vendor
	return wifiInterface, nil
}

// Scan returns a list of all reachable WiFi networks
func (wifiInterface *WifiInterface) Scan() ([]WifiNetwork, error) {
	airportNetworks, airportErr := airport.Scan()
	if airportErr != nil {
		return nil, airportErr
	}

	wifiNetworks := []WifiNetwork{}
	for _, network := range airportNetworks {
		wifiNetwork := WifiNetwork{}
		wifiNetwork.SSID = network.SSID
		wifiNetwork.Channel = network.Channel
		wifiNetwork.SecurityType = network.Security
		wifiNetwork.QualityLevel = network.QualityLevel
		wifiNetwork.NoiseLevel = network.NoiseLevel
		wifiNetworks = append(wifiNetworks, wifiNetwork)
	}
	return wifiNetworks, nil
}

// GetAPs returns all networks under the same SSID
func GetAPs(ssid string, networks []WifiNetwork) ([]WifiNetwork, error) {
	accessPoints := []WifiNetwork{}
	for _, network := range networks {
		if network.SSID == ssid {
			accessPoints = append(accessPoints, network)
		}
	}
	if len(accessPoints) < 1 {
		return accessPoints, errors.New("wifi: no access points found with SSID " + ssid)
	}
	return accessPoints, nil
}

// GetBestAP returns the access point with the provided SSID that
// has the best quality connection
func GetBestAP(accessPoints []WifiNetwork) (WifiNetwork, error) {
	if len(accessPoints) == 1 {
		return accessPoints[0], nil
	}
	var bestAP WifiNetwork
	for _, accessPoint := range accessPoints {
		if accessPoint.QualityLevel > bestAP.QualityLevel {
			bestAP = accessPoint
		}
	}
	return bestAP, nil
}

// UpdateNetwork updates the connection of the interface
func (wifiInterface *WifiInterface) UpdateNetwork(network WifiNetwork) {
	wifiInterface.Connection = network
}

// Connect the interface to the current WiFi connection
func (wifiInterface *WifiInterface) Connect() error {
	connectErr := networkSetup.Connect(wifiInterface.Name, wifiInterface.Connection.SSID, wifiInterface.Connection.SecurityKey)
	if connectErr != nil {
		return connectErr
	}
	return nil
}

// Disconnect disconnects from the current network without shutting
// down the interface
func (wifiInterface *WifiInterface) Disconnect() error {
	disconnectErr := airport.Disconnect()
	if disconnectErr != nil {
		return disconnectErr
	}
	return nil
}

// UpdateSecurityKey updates the security key of the network
func (wifiNetwork *WifiNetwork) UpdateSecurityKey(key string) {
	wifiNetwork.SecurityKey = key
}

// Prerequisites returns whether or not all the required
// commands are installed
func Prerequisites() bool {
	commandList := map[string]bool{
		"airport":        airport.IsInstalled(),
		"networkSetup":   networkSetup.IsInstalled(),
		"systemProfiler": systemProfiler.IsInstalled(),
	}

	needList := []string{}

	for command, installed := range commandList {
		if !installed {
			needList = append(needList, command)
		}
	}

	if len(needList) > 0 {
		return false
	}
	return true
}
