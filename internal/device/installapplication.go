package device

import (
	"log"
)

func (device *Device) installApplication(cmd *InstallApplicationCommand) error {
	log.Println("call installApplication. mock installation. do nothing")
	return nil
}