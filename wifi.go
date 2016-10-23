package wifimanager

import (
	"errors"
	"fmt"
	"net"

	"github.com/ottopress/WifiManager/darwin"
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

	// ErrMissingIface should be returned if no interfaces could be found
	// while getting available interfaces
	ErrMissingIface = errors.New("wifi: no wifi interfaces found")
	// ErrMissingAP should be returned if scanning for an access point with
	// a specific SSID yielded no results
	ErrMissingAP = errors.New("wifi: no access point found with provided name")
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
	SSID        string
	BSSID       string
	RSSI        int
	HT          bool
	Channel     int
	Security    []WifiNetworkSecurity
	SecurityKey string
}

// WifiNetworkSecurity represents the security configuration of
// a WiFi network.
type WifiNetworkSecurity struct {
	Protocol int
	Method   int
	Unicasts []int
	Group    int
}

// GetWifiInterfaces returns a list of all active Wifi interfaces
func GetWifiInterfaces() ([]WifiInterface, error) {
	wifiInterfaces := []WifiInterface{}

	netInterfaces, netErr := net.Interfaces()
	if netErr != nil {
		return wifiInterfaces, netErr
	}

	_, runErr := systemProfiler.Run(networkSetup)
	if runErr != nil {
		return wifiInterfaces, runErr
	}
	for _, iface := range netInterfaces {
		_, spErr := systemProfiler.Get(iface.Name)
		if spErr == nil {
			wifiInterface, _ := NewWifiInterface(iface)
			wifiInterfaces = append(wifiInterfaces, wifiInterface)
		}
	}

	if len(wifiInterfaces) < 1 {
		return wifiInterfaces, ErrMissingIface
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
	fmt.Println("Starting scan")
	airportNetworks, airportErr := airport.Scan()
	fmt.Println("Middle part")
	if airportErr != nil {
		return nil, airportErr
	}
	fmt.Println("Ending scan")
	wifiNetworks := []WifiNetwork{}
	for _, network := range airportNetworks {
		security := []WifiNetworkSecurity{}
		for _, airSecurity := range network.Security {
			security = append(security, WifiNetworkSecurity{
				Protocol: airSecurity.Protocol,
				Method:   airSecurity.Method,
				Unicasts: airSecurity.Unicasts,
				Group:    airSecurity.Group,
			})
		}
		wifiNetworks = append(wifiNetworks, WifiNetwork{
			SSID:     network.SSID,
			BSSID:    network.BSSID,
			RSSI:     network.RSSI,
			Channel:  network.Channel,
			Security: security,
			HT:       network.HT,
		})
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
		return accessPoints, ErrMissingAP
	}
	return accessPoints, nil
}

// GetBestAP returns the access point with the provided SSID that
// has the best quality connection
func GetBestAP(accessPoints []WifiNetwork) (WifiNetwork, error) {
	if len(accessPoints) == 1 {
		return accessPoints[0], nil
	}
	bestAP := accessPoints[0]
	for _, accessPoint := range accessPoints {
		if accessPoint.RSSI < bestAP.RSSI {
			bestAP = accessPoint
		}
	}
	return bestAP, nil
}

// UpdateNetwork updates the connection of the interface
func (wifiInterface *WifiInterface) UpdateNetwork(network WifiNetwork) {
	wifiInterface.Connection = network
}

// Up turns on the WiFi interface
func (wifiInterface *WifiInterface) Up() error {
	upErr := networkSetup.Up(wifiInterface.Name)
	if upErr != nil {
		return upErr
	}
	return nil
}

// Down turns off the WiFi interface
func (wifiInterface *WifiInterface) Down() error {
	downErr := networkSetup.Down(wifiInterface.Name)
	if downErr != nil {
		return downErr
	}
	return nil
}

// Connect the interface to the current WiFi connection
func (wifiInterface *WifiInterface) Connect() error {
	connectErr := networkSetup.Connect(wifiInterface.Name, wifiInterface.Connection.SSID, wifiInterface.Connection.SecurityKey)
	if connectErr != nil {
		return connectErr
	}
	return nil
}

// Status returns the power state of the WiFi interface
func (wifiInterface *WifiInterface) Status() (bool, error) {
	status, statusErr := networkSetup.Status(wifiInterface.Name)
	if statusErr != nil {
		return false, statusErr
	}
	return status, nil
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
