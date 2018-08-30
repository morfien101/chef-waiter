package config

var defaultFileLocation = "/etc/chefwaiter/config.json"

func (vc *ValuesContainer) writeConfigFileOSDefaults() {
	vc.InternalLogLocation = "/var/log/chefwaiter"
	vc.InternalStateFileLocation = "/etc/chefwaiter"
}
