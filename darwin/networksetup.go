package darwin

import (
	"errors"
	"os/exec"
	"regexp"
	"strconv"
)

// NetworkSetup is a wrapper for the Mac OS X networksetup command.
type NetworkSetup struct{}

// NewNetworkSetup creates a new instance of a NetworkSetup
// command wrapper.
func NewNetworkSetup() *NetworkSetup {
	return &NetworkSetup{}
}

// IsInstalled returns whether or not the networksetup executable
// can be found in the current PATH environment variable.
func (networkSetup *NetworkSetup) IsInstalled() bool {
	_, err := exec.LookPath("networksetup")
	if err != nil {
		return false
	}
	return true
}

// Connect initializes a connection on the provided interface to the given
// network.
func (networkSetup *NetworkSetup) Connect(iface, ssid, password string) error {
	cmd := exec.Command("networksetup", "-setairportnetwork", iface, ssid, password)
	_, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return cmdErr
	}
	return nil
}

// Up turns on the provided interface
func (networkSetup *NetworkSetup) Up(iface string) error {
	cmd := exec.Command("networksetup", "-setirportpower", iface, "on")
	_, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return cmdErr
	}
	return nil
}

// Down turns off the provided interface
func (networkSetup *NetworkSetup) Down(iface string) error {
	cmd := exec.Command("networksetup", "-setirportpower", iface, "off")
	_, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return cmdErr
	}
	return nil
}

// GetMTU returns the MTU value of the provided interface
func (networkSetup *NetworkSetup) GetMTU(iface string) (int, error) {
	cmd := exec.Command("networksetup", "-getMTU", iface)
	cmdOut, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return 0, cmdErr
	}
	regexMtu, regexErr := regexp.Compile(`.*: (\d+).*`)
	if regexErr != nil {
		return 0, regexErr
	}
	mtuVal := regexMtu.FindAllStringSubmatch(string(cmdOut), -1)
	if mtuVal != nil {
		conv, convErr := strconv.Atoi(mtuVal[0][1])
		if convErr != nil {
			return 0, convErr
		}
		return conv, nil
	}
	return 0, errors.New("networksetup: mtu couldn't be parsed")
}
