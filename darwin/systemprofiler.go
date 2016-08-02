package darwin

import (
	"errors"
	"os/exec"
	"regexp"

	"github.com/DHowett/go-plist"
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
)

// SystemProfiler is a wrapper for the Mac OS X system_profiler command.
type SystemProfiler struct {
	outputCache *SystemProfilerOutput
}

// SystemProfilerOutput represents the output of the system_profiler command
type SystemProfilerOutput struct {
	SPCompletionInterval float64              `plist:"_SPCompletionInterval"`
	SPResponseTime       float64              `plist:"_SPResponseTime"`
	DataType             string               `plist:"_dataType"`
	DetailLevel          int                  `plist:"_detailLevel"`
	ParentDataType       string               `plist:"_parentDataType"`
	SystemProfilerItems  []SystemProfilerItem `plist:"_items"`
}

// SystemProfilerItem represents a dict in the SystemProfiler output
type SystemProfilerItem struct {
	SystemProfilerInterfaces []SystemProfilerInterface `plist:"spairport_airport_interfaces"`
	SystemProfilerSoftware   SystemProfilerSoftware    `plist:"spairport_software_information"`
}

// SystemProfilerInterface represents a Wifi Interface in the SystemProfiler
// output.
type SystemProfilerInterface struct {
	Name          string `plist:"_name"`
	Vendor        string
	ID            string
	Status        int
	SPAirDrop     string `plist:"spairport_caps_airdrop"`
	SPWoW         string `plist:"spairport_caps_wow"`
	SPStatus      string `plist:"spairport_status_connected"`
	SPChannels    []int  `plist:"spairport_supported_channels"`
	SPPhyModes    string `plist:"spairport_supported_phymodes"`
	SPCardType    string `plist:"spairport_wireless_card_type"`
	SPCountryCode string `plist:"spairport_wireless_country_code"`
	SPFirmware    string `plist:"spairport_wireless_firmware_version"`
	SPLocale      string `plist:"spairport_wireless_locale"`
	SPMacAddr     string `plist:"spairport_wireless_mac_address"`
}

// SystemProfilerSoftware represents the software information
// returned in the SystemProfiler output
type SystemProfilerSoftware struct {
	SPCoreWlan    string `plist:"spairport_corewlan_version"`
	SPCoreWlanKit string `plist:"spairport_corewlankit_version"`
	SPDiagnostics string `plist:"spairport_diagnostics_version"`
	SPExtra       string `plist:"spairport_extra_version"`
	SPFamily      string `plist:"spairport_family_version"`
	SPProfiler    string `plist:"spairport_profiler_version"`
	SPUtility     string `plist:"spairport_utility_version"`
}

// NewSystemProfiler creates a new instance of a SystemProfiler
// command wrapper.
func NewSystemProfiler() SystemProfiler {
	return SystemProfiler{}
}

// IsInstalled returns whether or not the system_profiler executable
// can be found in the current PATH environment variable.
func (systemProfiler *SystemProfiler) IsInstalled() bool {
	_, err := exec.LookPath("system_profiler")
	if err != nil {
		return false
	}
	return true
}

// Run the system_profiler command and both cache and return the output
func (systemProfiler *SystemProfiler) Run() (*SystemProfilerOutput, error) {
	cmd := exec.Command("system_profiler", "-detailLevel", "mini", "SPAirPortDataType", "-xml")
	cmdOut, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return nil, cmdErr
	}
	parseOut, parseErr := systemProfiler.parseOutput(cmdOut)
	if parseErr != nil {
		return nil, parseErr
	}
	systemProfiler.outputCache = parseOut
	return systemProfiler.outputCache, nil
}

// Get returns an instance of a wireless interface if
// one exists with the provided name
func (systemProfiler *SystemProfiler) Get(iface string) (SystemProfilerInterface, error) {
	for _, spIface := range systemProfiler.outputCache.SystemProfilerItems[0].SystemProfilerInterfaces {
		if spIface.Name == iface {
			return spIface, nil
		}
	}
	return SystemProfilerInterface{}, errors.New("systemprofiler: no wireless interface found with name " + iface)
}

func (systemProfiler *SystemProfiler) parseOutput(output []byte) (*SystemProfilerOutput, error) {
	var marshal []SystemProfilerOutput
	_, marshalErr := plist.Unmarshal(output, &marshal)
	if marshalErr != nil {
		return nil, marshalErr
	}
	systemProfilerOutput := &marshal[0]
	// regexCardAttr captures the vendor and ID from the card type attribute
	regexCardAttr, regexErr := regexp.Compile(`.*\((.+), (.+)\).*`)
	if regexErr != nil {
		return nil, regexErr
	}
	for index := range systemProfilerOutput.SystemProfilerItems[0].SystemProfilerInterfaces {
		systemProfilerInterface := &systemProfilerOutput.SystemProfilerItems[0].SystemProfilerInterfaces[index]
		cardAttr := regexCardAttr.FindAllStringSubmatch(systemProfilerInterface.SPCardType, -1)
		if cardAttr != nil {
			systemProfilerInterface.Vendor = cardAttr[0][1]
			systemProfilerInterface.ID = cardAttr[0][2]
		}
		switch systemProfilerInterface.SPStatus {
		case "spairport_status_connected":
			systemProfilerInterface.Status = IfaceConnected
		case "spairport_status_disassociated":
			systemProfilerInterface.Status = IfaceDisassociated
		case "spairport_status_off":
			systemProfilerInterface.Status = IfaceOff
		}
	}
	return systemProfilerOutput, nil
}
