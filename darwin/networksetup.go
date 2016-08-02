package darwin

import "os/exec"

// NetworkSetup is a wrapper for the Mac OS X networksetup command.
type NetworkSetup struct{}

// NewNetworkSetup creates a new instance of a NetworkSetup
// command wrapper.
func NewNetworkSetup() NetworkSetup {
	return NetworkSetup{}
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
