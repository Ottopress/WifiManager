package darwin

import (
	"os"
	"os/exec"

	"github.com/DHowett/go-plist"
)

const (
	// WPA represents the WPA WiFi security protocol
	WPA WifiSecurity = iota
	// WEP represents the WEP WiFi security protocol
	WEP
	// WPA2 represents the WPA2/RSN WiFi security protocol
	WPA2
	// None represents an open WiFi network
	None
)

// AirPort is a wrapper for the Mac OS X airport command
type AirPort struct {
	outputCache []AirPortNetwork
}

// AirPortNetwork represents a WiFi network from the output
// of the airport command
type AirPortNetwork struct {
	SSID           string `plist:"SSID_STR"`
	Security       WifiSecurity
	BSSID          string          `plist:"BSSID"`
	Channel        int             `plist:"CHANNEL"`
	TxRate         []int           `plist:"RATES"`
	NoiseLevel     int             `plist:"NOISE"`
	QualityLevel   int             `plist:"RSSI"`
	APMode         int             `plist:"AP_MODE"`
	BeaconInterval int             `plist:"BEACON_INT"`
	SecurityRSN    *WifiSecurityIE `plist:"RSN_IE"`
	SecurityWPA    *WifiSecurityIE `plist:"WPA_IE"`
	SecurityWEP    *WifiSecurityIE `plist:"WEP_IE"`
}

// WifiSecurityIE represents the different possible WiFi security types
// represented in the XML as dicts. This is empty since we care only about
// the presence of the dicts and not about the contents.
type WifiSecurityIE struct{}

// WifiSecurity is an enum for the different WiFi security protocols
type WifiSecurity int

// NewAirPort creates a new instance of the AirPort
// command wrapper.
func NewAirPort() AirPort {
	return AirPort{}
}

// IsInstalled returns whether or not the airport executable
// can be found in its specialized location.
func (airport *AirPort) IsInstalled() bool {
	if _, statErr := os.Stat("/System/Library/PrivateFrameworks/Apple80211.framework/Versions/A/Resources/airport"); statErr != nil {
		if os.IsNotExist(statErr) {
			return false
		}
	}
	return true
}

// Scan using the airport command and both cache and return the output
func (airport *AirPort) Scan() ([]AirPortNetwork, error) {
	cmd := exec.Command("/System/Library/PrivateFrameworks/Apple80211.framework/Versions/A/Resources/airport", "-s", "-x")
	cmdOut, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return nil, cmdErr
	}

	parseOut, parseErr := airport.parseOutput(cmdOut)
	if parseErr != nil {
		return nil, parseErr
	}
	airport.outputCache = parseOut
	return parseOut, nil
}

func (airport *AirPort) parseOutput(output []byte) ([]AirPortNetwork, error) {
	var networks []AirPortNetwork
	_, marshalErr := plist.Unmarshal(output, &networks)
	if marshalErr != nil {
		return nil, marshalErr
	}
	for index := range networks {
		network := &networks[index]
		if network.SecurityRSN != nil {
			network.Security = WPA2
		} else if network.SecurityWPA != nil {
			network.Security = WPA
		} else if network.SecurityWEP != nil {
			network.Security = WEP
		} else {
			network.Security = None
		}
	}
	return networks, nil
}
