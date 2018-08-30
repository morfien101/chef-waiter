package config

var defaultFileLocation = "C:\\Program Files\\chefwaiter\\config.json"

func (vc *ValuesContainer) writeConfigFileOSDefaults() {
	vc.InternalLogLocation = "c:\\logs\\chefwaiter"
	vc.InternalStateFileLocation = "C:\\Program Files\\chefwaiter"
}
