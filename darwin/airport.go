package darwin

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	// WPA represents the WPA WiFi security protocol
	WPA int = iota
	// WEP represents the WEP WiFi security protocol
	WEP
	// WPA2 represents the WPA2/RSN WiFi security protocol
	WPA2
	// NONE represents an open WiFi network
	NONE
	// AES represents the AES-based CCMP integrity check
	// protocol.
	AES int = iota
	// TKIP represents the TKIP integrity check protocol
	TKIP
	// PSK represents the PSK authentication method for the
	// WiFi network
	PSK int = iota
	// EAP represents the EAP/802.1x authentication method for
	// the WiFi network
	EAP
	// AirPortRE is the regex used to parse the output of the
	// Mac OS X airport command
	// </br>
	// It should be noted that while this may not be the most optimal
	// solution, it is faster than parsing the plist simply due to the
	// considerable amount of data that is provided with the plist
	// format as opposed to running the command normally. As such,
	// this regex will stay until speed becomes a concern or the need
	// arises for the extra data that plist provides.
	AirPortRE = "\\s*([a-zA-Z0-9-_\\s ]*)\\s*([a-fA-F0-9]{2}:[a-fA-F0-9]{2}:[a-fA-F0-9]{2}:[a-fA-F0-9]{2}:[a-fA-F0-9]{2}:[a-fA-F0-9]{2})\\s*([-|+]{1}[0-9]*)\\s*([0-9]*),*[-|+]*[0-9]*\\s*([Y|N]{1})\\s*([A-Z-]*)\\s*(NONE|(?:[a-zA-Z0-9]+))(?:\\((.+?)\\/(.+?)(?:,(.+?))?\\/(.+?)\\))?\\s+?(?:([a-zA-Z0-9]+)\\((.+?)\\/(.+?)(?:,(.+?))?\\/(.+?)\\))?"
)

var (
	// AirPortCompiledRE is the compiled regex of the AirPortRE
	// constant. This is initialized outside any method scope
	// to prevent redundant computing.
	AirPortCompiledRE = regexp.MustCompile(AirPortRE)
	// ProtoConv is a map of the different protocol values to their
	// respective constant values
	ProtoConv = map[string]int{
		"WPA":  WPA,
		"WEP":  WEP,
		"WPA2": WPA2,
		"NONE": NONE,
	}
	// CipherConv is a map of the different available ciphers to
	// their respective constant values
	CipherConv = map[string]int{
		"AES":  AES,
		"TKIP": TKIP,
	}
	// AuthConv is a map of the different available authentication
	// methods to their respective constant values.
	AuthConv = map[string]int{
		"PSK":    PSK,
		"802.1x": EAP,
	}
)

// AirPort is a wrapper for the Mac OS X airport command
type AirPort struct {
	outputCache []AirPortNetwork
}

// AirPortNetwork represents a WiFi network from the output
// of the airport command
type AirPortNetwork struct {
	SSID        string
	BSSID       string
	RSSI        int
	Channel     int
	HT          bool
	CountryCode string
	Security    []AirPortNetworkSecurity
}

// AirPortNetworkSecurity represents a WiFi network's different
// security parameters
type AirPortNetworkSecurity struct {
	Protocol int
	Method   int
	Unicasts []int
	Group    int
}

// NewAirPort creates a new instance of the AirPort
// command wrapper.
func NewAirPort() *AirPort {
	return &AirPort{}
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
	cmd := exec.Command("/System/Library/PrivateFrameworks/Apple80211.framework/Versions/A/Resources/airport", "-s")
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

// Get all networks that match the provided SSID
func (airport *AirPort) Get(ssid string) []AirPortNetwork {
	possibleNetworks := []AirPortNetwork{}
	for index := range airport.outputCache {
		network := airport.outputCache[index]
		if network.SSID == ssid {
			possibleNetworks = append(possibleNetworks, network)
		}
	}
	return possibleNetworks
}

// Disconnect disconnects from the current network without shutting
// down the interface
func (airport *AirPort) Disconnect() error {
	cmd := exec.Command("/System/Library/PrivateFrameworks/Apple80211.framework/Versions/A/Resources/airport", "--disassociate")
	_, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return cmdErr
	}
	return nil
}

func (airport *AirPort) parseOutput(output []byte) ([]AirPortNetwork, error) {
	var networks []AirPortNetwork
	scanner := bufio.NewScanner(bytes.NewReader(output))
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		network, networkErr := airport.parseSingle(scanner.Text())
		if networkErr != nil {
			return networks, networkErr
		}
		if network != nil {
			networks = append(networks, *network)
		}
	}
	return networks, nil
}

// parseSingle item takes a single piece of text and returns
// the most complete possible AirPortNetwork struct, or nil
// if there are no matches found.
// </br>
// parseSingle assumes the format of the item is:
// <SSID> <BSSID> <RSSI> <Channel> <HT> <CC> <SecProto>(<SecMeth>/<Ciphers>/<Group Cipher>)
func (airport *AirPort) parseSingle(item string) (*AirPortNetwork, error) {
	matches := AirPortCompiledRE.FindStringSubmatch(item)
	if len(matches) == 0 {
		return nil, nil
	}
	matches = matches[1:]
	rssiVal, rssiErr := strconv.Atoi(matches[2])
	if rssiErr != nil {
		return nil, rssiErr
	}
	var htVal bool
	if matches[4] == "Y" {
		htVal = true
	} else {
		htVal = false
	}
	channelVal, channelErr := strconv.Atoi(matches[3])
	if channelErr != nil {
		return nil, channelErr
	}
	security := []AirPortNetworkSecurity{}
	if matches[6] != "NONE" {
		for i := 6; i < len(matches); i += 5 {
			unicasts := []int{}
			unicasts = append(unicasts, CipherConv[matches[i+2]])
			if matches[i+2] != "" {
				unicasts = append(unicasts, CipherConv[matches[i+3]])
			}
			security = append(security, AirPortNetworkSecurity{
				Protocol: ProtoConv[matches[i]],
				Method:   AuthConv[matches[i+1]],
				Unicasts: unicasts,
				Group:    CipherConv[matches[i+4]],
			})
		}
	} else {
		security = append(security, AirPortNetworkSecurity{Protocol: ProtoConv[matches[6]]})
	}

	return &AirPortNetwork{
		SSID:        strings.Trim(matches[0], " "),
		BSSID:       strings.Trim(matches[1], " "),
		RSSI:        rssiVal,
		Channel:     channelVal,
		HT:          htVal,
		CountryCode: strings.Trim(matches[5], " "),
		Security:    security,
	}, nil
}
