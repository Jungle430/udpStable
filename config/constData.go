package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

// 通信类型
type CommunicationType string

const (
	// 服务器
	SERVER CommunicationType = "server"
	// 客户端
	CLIENT CommunicationType = "client"
)

const (
	// 缓冲区大小
	BUFFER_SIZE int = 1 << 12

	// 最大等待时间
	MAX_WAIT_TIME time.Duration = time.Duration(5) * time.Second

	// 等待时间
	WAIT_TIME time.Duration = time.Duration(2) * time.Second

	// 接收消息间隔
	READ_INTERVAL_TIME time.Duration = time.Duration(1) * time.Second

	// ACK消息随机编码长度
	ACK_ENCIPHER_LENGTH int = 1 << 5

	// 私有密钥长度
	PRIVATE_KEY_LENGTH int = 10

	// 私有密钥配置文件
	PRIVATE_KEY_FILE string = "resource/private_key.json"

	// 日志文件(客户端)
	CLIENT_LOG_FILE string = "log/udpstable-client.log"

	// 日志文件(服务端)
	SERVER_LOG_FILE string = "log/udpstable-server.log"

	// 启动命令行的模式
	COMMAND_FORMAT string = "go run main.go <server|client> <port>"
)

var (
	// 私有密钥
	PRIVATE_KEY [PRIVATE_KEY_LENGTH]byte

	// LOCALHOST_IP
	LOCALHOST_IP net.IP = net.IPv4(127, 0, 0, 1)
)

var (
	// 序列号不匹配
	ErrSeqMatch error = errors.New("seq number not match")

	// CRC校验失败
	ErrCrcCheck error = errors.New("crc check failed")

	// 数据长度不匹配
	ErrDataLengthMatch error = errors.New("data length does not match")

	// 服务端生成ACK随机编码错误
	ErrAckRandom error = errors.New("error generating ACK random bytes")

	// 服务器长时间未响应
	ErrServerNotRespondLongTime error = errors.New("the server has not responded for a long time")
)

// 初始化配置(私有密钥)
func InitPortAndPrivateKey() {
	// 载入文件
	privateKeyData, err := os.ReadFile(PRIVATE_KEY_FILE)
	if err != nil {
		log.Panic(err)
	}

	// 反序列化私有密钥文件的临时结构体
	type KeyData struct {
		PrivateKey [PRIVATE_KEY_LENGTH]string `json:"private_key"`
	}
	// JSON反序列化
	var keyData KeyData
	err = json.Unmarshal(privateKeyData, &keyData)
	if err != nil {
		log.Panic(fmt.Sprintf("配置文件:%sJSON反序列化失败", PRIVATE_KEY_FILE))
	}

	// 拷贝数据
	for i := 0; i < PRIVATE_KEY_LENGTH; i++ {
		uintValue, err := strconv.ParseUint(keyData.PrivateKey[i], 0, 0)
		if err != nil {
			log.Panic(fmt.Sprintf("配置文件:%sJSON解析失败, 数据不符:%s", PRIVATE_KEY_FILE, err))
		}

		// 必须在byte数据范围内
		if uintValue > 0xFF {
			log.Panic(fmt.Sprintf("配置文件:%sJSON解析失败, 数据%d不是一个byte", PRIVATE_KEY_FILE, uintValue))
		}

		PRIVATE_KEY[i] = byte(uintValue)
	}
	log.Debug("端口与私有密钥配置初始化完成")
}

// 初始化日志配置
func InitLog(fileType CommunicationType) *os.File {
	// 指定输出格式和级别
	log.SetReportCaller(true)
	log.SetLevel(log.InfoLevel)

	// 记录日志的文件
	var (
		logFile *os.File
		err     error
	)

	// 根据参数来选取模式
	switch fileType {
	case CLIENT:
		logFile, err = os.OpenFile(CLIENT_LOG_FILE, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	case SERVER:
		logFile, err = os.OpenFile(SERVER_LOG_FILE, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	default:
		panic(COMMAND_FORMAT)
	}

	if err != nil {
		log.Panic(err)
	}
	// 两个输出地点:缓冲区+日志文件
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	log.Debug("日志配置初始化完成")
	return logFile
}
