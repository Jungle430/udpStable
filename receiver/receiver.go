package receiver

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
	"udpStable/config"
	"udpStable/dto/message"

	log "github.com/sirupsen/logrus"
)

func Receiver(port int, waitTime time.Duration) ([]byte, error) {
	// 开启端口监听
	receiver, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", config.LOCALHOST_IP, port))
	if err != nil {
		log.Warn(err)
		return nil, err
	}

	startTime := time.Now()
	for {
		if time.Since(startTime) >= waitTime {
			err := fmt.Errorf("超出设置的最长时间:%v", waitTime)
			log.Warn(err)
			return nil, err
		}
		conn, err := net.ListenUDP("udp", receiver)
		if err != nil {
			log.Warn(err)
			return nil, err
		}
		defer conn.Close()

		// 设置等待时间
		err = conn.SetDeadline(time.Now().Add(waitTime))
		if err != nil {
			log.Warn(err)
			return nil, err
		}

		// 接收信息
		buffer := make([]byte, config.BUFFER_SIZE)
		n, _, err := conn.ReadFrom(buffer)

		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				err := fmt.Errorf("超出设置的最长时间:%v", waitTime)
				log.Warn(err)
				return nil, err
			}
			return nil, err
		}

		receiveMessage, err := message.DecodeWithoutSeqNum(buffer[:n])

		// 接收成功
		if err == nil && !receiveMessage.IsAck {
			conn.Close()
			err := send(receiveMessage.DestinationAddress, receiveMessage.DestinationPort, receiveMessage.SourceAddress, receiveMessage.SourcePort, receiveMessage.SeqNum)
			if err != nil {
				log.Warn(err)
				return nil, err
			}
			return receiveMessage.Data, nil
		}

		// 接收失败
		// 接收失败,两种情况
		// 1. 解码错误
		if err != nil {
			log.Warn(err)
			return nil, err
		}
		// 2. 是ACK信息，忽略
		if receiveMessage.IsAck {
			continue
		}
	}
}

// 发送确认消息
func send(sourceAddress net.IP, sourcePort int, destinationAddress net.IP, destinationPort int, seqNum uint64) error {
	// 指定发送地址和端口
	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", destinationAddress.String(), destinationPort))
	if err != nil {
		return err
	}

	// 指定发送地址和端口
	localAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", sourceAddress.String(), sourcePort))
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", localAddr, serverAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	// 生成数据报
	m, err := message.NewAck(
		seqNum,
		localAddr.IP, localAddr.Port,
		serverAddr.IP, serverAddr.Port,
	)

	if err != nil {
		return err
	}

	// 如果为0证明是操作系统随机分配的，需要从conn获取真实信息
	if localAddr.Port == 0 {
		p, err := strconv.Atoi(strings.Split(conn.LocalAddr().String(), ":")[1])
		if err != nil {
			log.Error("操作系统分配端口错误!实际信息:local", localAddr)
			return err
		}
		m.SourcePort = p
	}
	serviceMessage, err := m.Encode()
	if err != nil {
		return err
	}

	// 发送消息
	_, err = conn.Write(serviceMessage)

	if err != nil {
		return err
	}
	return nil
}
