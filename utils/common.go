package utils

import (
	"log"
	"os"
)

func WarningAsRoot() {
	if os.Getuid() == 0 {
		log.Println("You run warden agent as THE SUPER USER!")
		log.Println("It's not recommonded that running warden as root.")
		log.Println("Try: useradd -g docker warden. Then run warden as user 'warden'.")
	}
}
