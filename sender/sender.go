package sender

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
	"udpStable/config"
	"udpStable/dto/message"

	log "github.com/sirupsen/logrus"
)

// 发送消息API(不指明端口)
func SenderWithoutPort(data []byte, sourceAddress net.IP, destinationAddress net.IP, destinationPort int) error {
	m, err := senderWithoutPort(data, sourceAddress, destinationAddress, destinationPort)
	if err != nil {
		log.Warn("udp发送信息失败,错误:", err)
		return err
	}

	// 开启端口监听
	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", sourceAddress.String(), m.SourcePort))
	if err != nil {
		log.Warn("监听失败,错误:", err)
		return err
	}
	conn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		log.Warn("监听失败,错误:", err)
		return err
	}
	defer conn.Close()

	// 开始计时
	startTime := time.Now()
	for {
		// 超时了直接返回
		if time.Since(startTime) >= config.MAX_WAIT_TIME {
			log.Error("超时,接收方信息:IP:", destinationAddress, "port:", destinationPort)
			return config.ErrServerNotRespondLongTime
		}

		// 防止堵塞设置最长接收信息时间,本轮不行就下一轮
		conn.SetDeadline(time.Now().Add(config.WAIT_TIME))

		// 接收信息
		buffer := make([]byte, config.BUFFER_SIZE)
		n, _, err := conn.ReadFromUDP(buffer)

		if err != nil {
			// 超时就下一轮,同时在发一次
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Warn(err)
				// 再次发送信息
				conn.Close()
				m, _ = senderWithoutPort(data, sourceAddress, destinationAddress, destinationPort)

				// 发送之后更新监听端口
				serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", sourceAddress.String(), m.SourcePort))
				if err != nil {
					log.Warn(err)
					return err
				}
				conn, err = net.ListenUDP("udp", serverAddr)
				if err != nil {
					log.Warn("监听失败,错误:", err)
					return err
				}
			} else {
				log.Warn(err)
				return err
			}
			continue
		}

		receiveMessage, err := message.DecodeWithSeqNum(buffer[:n], m.SeqNum)

		// 接收成功
		if err == nil && receiveMessage.IsAck {
			break
		}

		// 接收失败,两种情况
		// 1. 解码错误
		if err != nil {
			log.Warn(err)
			// 再次发送信息
			m, _ = senderWithoutPort(data, sourceAddress, destinationAddress, destinationPort)
			conn, err = net.ListenUDP("udp", serverAddr)
			if err != nil {
				log.Warn("监听失败,错误:", err)
				return err
			}
		}
		// 2. 不是ACK信息，忽略
		if !receiveMessage.IsAck {
			continue
		}
	}
	return nil
}

// 发送消息API(指明端口)
func SenderWithPort(data []byte, sourceAddress net.IP, sourcePort int, destinationAddress net.IP, destinationPort int) error {
	// 根据端口发送消息
	m, err := senderWithPort(data, sourceAddress, sourcePort, destinationAddress, destinationPort)
	if err != nil {
		log.Warn("udp发送信息失败,错误:", err)
		return err
	}

	// 开启端口监听
	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", sourceAddress.String(), sourcePort))
	if err != nil {
		log.Warn(err)
		return err
	}
	conn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		log.Warn("监听失败,错误:", err)
		return err
	}
	defer conn.Close()

	// 开始计时
	startTime := time.Now()
	for {
		// 超时了直接返回
		if time.Since(startTime) >= config.MAX_WAIT_TIME {
			log.Error("超时,接收方信息:IP:", destinationAddress, "port:", destinationPort)
			return config.ErrServerNotRespondLongTime
		}

		// 防止堵塞设置最长接收信息时间,本轮不行就下一轮
		conn.SetDeadline(time.Now().Add(config.WAIT_TIME))

		// 接收信息
		buffer := make([]byte, config.BUFFER_SIZE)
		n, _, err := conn.ReadFromUDP(buffer)

		if err != nil {
			// 超时就下一轮,同时在发一次
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Warn(err)

				// 重新发送
				conn.Close()
				m, _ = senderWithoutPort(data, sourceAddress, destinationAddress, destinationPort)

				// 继续监听
				conn, err = net.ListenUDP("udp", serverAddr)
				if err != nil {
					log.Warn("监听失败,错误:", err)
					return err
				}
			} else {
				log.Warn(err)
				return err
			}
			continue
		}

		receiveMessage, err := message.DecodeWithSeqNum(buffer[:n], m.SeqNum)

		// 接收成功
		if err == nil && receiveMessage.IsAck {
			break
		}

		// 接收失败,两种情况
		// 1. 解码错误
		if err != nil {
			log.Warn(err)
			// 再次发送信息
			m, _ = senderWithPort(data, sourceAddress, sourcePort, destinationAddress, destinationPort)
			conn, err = net.ListenUDP("udp", serverAddr)
			if err != nil {
				log.Warn("监听失败,错误:", err)
				return err
			}
		}
		// 2. 不是ACK信息，忽略
		if !receiveMessage.IsAck {
			continue
		}

	}

	return nil
}

// 发送UDP消息(非API,需要指定发送端口)
func senderWithPort(data []byte, sourceAddress net.IP, sourcePort int, destinationAddress net.IP, destinationPort int) (*message.Message, error) {
	// 指定发送地址和端口
	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", destinationAddress.String(), destinationPort))
	if err != nil {
		return nil, err
	}

	// 指定发送地址和端口
	localAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", sourceAddress.String(), sourcePort))
	if err != nil {
		return nil, err
	}

	// 发送消息
	m, err := send(data, serverAddr, localAddr)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// 发送UDP消息(非API, 发送端口由操作系统调度)
func senderWithoutPort(data []byte, sourceAddress net.IP, destinationAddress net.IP, destinationPort int) (*message.Message, error) {
	// 指定发送地址和端口
	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", destinationAddress.String(), destinationPort))
	if err != nil {
		return nil, err
	}

	// 指定发送端口(操作系统随机分配)
	localAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:0", sourceAddress.String()))

	if err != nil {
		return nil, err
	}

	// 发送消息
	m, err := send(data, serverAddr, localAddr)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// 底层API,更改了参数
func send(data []byte, serverAddr *net.UDPAddr, localAddr *net.UDPAddr) (*message.Message, error) {
	// 建立连接
	conn, err := net.DialUDP("udp", localAddr, serverAddr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// 生成序列码
	source := rand.NewSource(time.Now().UnixNano())
	randomGenerator := rand.New(source)
	seqNum := randomGenerator.Uint64()

	// 封装为数据报
	m := message.NewNotAck(
		seqNum, data,
		localAddr.IP, localAddr.Port,
		serverAddr.IP, serverAddr.Port,
	)

	// 如果为0证明是操作系统随机分配的，需要从conn获取真实信息
	if localAddr.Port == 0 {
		p, err := strconv.Atoi(strings.Split(conn.LocalAddr().String(), ":")[1])
		if err != nil {
			log.Error("操作系统分配端口错误!实际信息:local", localAddr)
			return nil, err
		}
		m.SourcePort = p
	}

	serviceMessage, err := m.Encode()
	if err != nil {
		return nil, err
	}

	// 发送消息
	_, err = conn.Write(serviceMessage)
	if err != nil {
		return nil, err
	}
	return m, nil
}
