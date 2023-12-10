package main

import (
	"fmt"
	"time"
	"udpStable/command"
	"udpStable/config"
	"udpStable/receiver"
	"udpStable/sender"

	log "github.com/sirupsen/logrus"
)

func main() {
	communicationType, port := command.CommandLaunch()

	// 初始化日志
	logFile := config.InitLog(communicationType)
	defer logFile.Close()

	// 初始化私有密钥
	config.InitPortAndPrivateKey()

	switch communicationType {
	case config.SERVER:
		server(port, time.Duration(10)*time.Second)
	case config.CLIENT:
		client(port)
	}
}

func server(port int, waitTime time.Duration) {
	for {
		data, err := receiver.Receiver(port, waitTime)
		if err != nil {
			log.Warn(err)
		} else {
			log.Info(data)
		}
	}
}

func client(port int) {
	fmt.Println("Enter your send message")
	var data string
	_, err := fmt.Scan(&data)

	if err != nil {
		log.Panic(err)
	}

	for i := 0; i < 10; i++ {
		err := sender.SenderWithoutPort([]byte(data), config.LOCALHOST_IP, config.LOCALHOST_IP, port)
		if err != nil {
			log.Warn(err)
		} else {
			log.Info("server send message success")
		}
	}
}
