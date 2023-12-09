package main

import (
	"os"
	"strings"
	"time"
	"udpStable/config"
	"udpStable/receiver"
	"udpStable/sender"

	"github.com/sirupsen/logrus"
)

func main() {
	args := os.Args

	logFile := config.InitLog()
	defer logFile.Close()
	config.InitPortAndPrivateKey()

	if strings.Compare(args[1], "1") == 0 {
		err := sender.SenderWithoutPort([]byte("Hello world"), config.LOCALHOST_IP, config.LOCALHOST_IP, 7777)
		logrus.Warn(err)
	} else {
		m, err := receiver.Receiver(7777, time.Duration(10)*time.Second)
		logrus.Warn(m, err)
	}
}
