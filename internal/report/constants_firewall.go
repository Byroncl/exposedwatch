package report

// Firewall table texts.
const (
	headerFirewall          = "FIREWALL"
	firewallSectionTitle    = "Firewall"
	firewallBackendFormat   = "  backend: %s\n"
	firewallAvailableFormat = "  available: %s\n"
	firewallNoteFormat      = "  - %s\n"
	firewallInformationText = "  Firewall verdicts are informational and never change severities.\n"
	availableYes            = "yes"
	availableNo             = "no"
)

// Verdict words. The Docker verdict is shortened in the table; JSON keeps the
// stable engine name.
const verdictWordDocker = "docker"
