package command

import (
	"os"
	"strconv"
	"udpStable/config"
)

// 启动命令行数据提取
func CommandLaunch() (config.CommunicationType, int) {
	args := os.Args

	if len(args) != 3 {
		panic(config.COMMAND_FORMAT)
	}

	var communicationType config.CommunicationType

	// client or server ?
	switch args[1] {
	case "server":
		communicationType = config.SERVER
	case "client":
		communicationType = config.CLIENT
	default:
		panic(config.COMMAND_FORMAT)
	}

	// port
	port, err := strconv.Atoi(args[2])
	if err != nil {
		panic(config.COMMAND_FORMAT)
	}

	return communicationType, port
}
