package consts

const (
	// TestNamespace is the default testing namespace
	TestNamespace = "sno-testing"
	// WPCDeviceType is the type of the tested WPC NIC
	WPCDeviceType = "E810-XXV-4T"
	// NVMFirmwareMinVersion is the minimal firmware version required for WPC NIC
	NVMFirmwareMinVersion = 4.01
	// DPLLLockedHOACQState is the required DPLL state -> 3: DPLL Locked Holdover Acquired
	DPLLLockedHOACQState = "3"
	// DPLLMaxAbsOffset is the maximum absolute DPLL offset in nanoseconds
	DPLLMaxAbsOffset = 30
)
