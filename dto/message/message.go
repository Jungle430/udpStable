package message

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"net"
	"udpStable/config"

	"github.com/sirupsen/logrus"
)

// 通信消息格式
type Message struct {
	// 序列标识
	SeqNum uint64
	// 数据
	Data []byte
	// 是应答数据还是发送数据
	IsAck bool
	// 数据长度
	Length int
	// CRC校验码
	CRC uint32
	// 源主机IP地址
	SourceAddress net.IP
	// 源主机端口
	SourcePort int
	// 目标主机通信地址
	DestinationAddress net.IP
	// 目标主机端口
	DestinationPort int
}

// 发送数据
func NewNotAck(seqNum uint64, data []byte, sourceAddress net.IP, sourcePort int, destinationAddress net.IP, destinationPort int) *Message {
	m := Message{
		SeqNum:             seqNum,
		Data:               data,
		IsAck:              false,
		Length:             len(data),
		CRC:                0,
		SourceAddress:      sourceAddress,
		SourcePort:         sourcePort,
		DestinationAddress: destinationAddress,
		DestinationPort:    destinationPort,
	}
	m.Encipher()
	return &m
}

// 收到消息的确认信号
func NewAck(seqNum uint64, sourceAddress net.IP, sourcePort int, destinationAddress net.IP, destinationPort int) (*Message, error) {
	// ACK的Data段本来是空的
	// 但是如果用空的进行CRC加密生成的CRC码都一样，私有密钥容易被破译
	// 这里Data段使用随机[]byte来保证数据安全
	randomBytes := make([]byte, config.ACK_ENCIPHER_LENGTH)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, config.ErrAckRandom
	}

	m := Message{
		SeqNum:             seqNum,
		Data:               randomBytes,
		IsAck:              true,
		Length:             len(randomBytes),
		CRC:                0,
		SourceAddress:      config.LOCALHOST_IP,
		SourcePort:         sourcePort,
		DestinationAddress: destinationAddress,
		DestinationPort:    destinationPort,
	}
	m.Encipher()
	return &m, nil
}

// 编码(JSON序列化)
func (m *Message) Encode() ([]byte, error) {
	return json.Marshal(m)
}

// 校验错误
type CheckError struct {
	// 序列号不匹配
	ErrSeqMatch error
	// CRC校验失败
	ErrCrcCheck error
	// 数据长度不匹配
	ErrDataLengthMatch error
}

func (c *CheckError) Error() string {
	return fmt.Sprintf("ErrSeqMath:%v\nErrCrcCheck:%v\nErrDataLengthMatch:%v\n",
		c.ErrSeqMatch, c.ErrCrcCheck, c.ErrDataLengthMatch)
}

// 解码(编码校验,此时需要核验序列码)
func DecodeWithSeqNum(data []byte, seqNum uint64) (*Message, error) {
	checkErr := CheckError{
		ErrSeqMatch:        nil,
		ErrCrcCheck:        nil,
		ErrDataLengthMatch: nil,
	}

	var message Message
	err := json.Unmarshal(data, &message)
	// JSON反序列化失败,解码失败直接返回
	if err != nil {
		return nil, err
	}

	// 序列号校验
	if message.SeqNum != seqNum {
		checkErr.ErrSeqMatch = config.ErrSeqMatch
	}

	// 长度校验
	if len(message.Data) != message.Length {
		checkErr.ErrDataLengthMatch = config.ErrDataLengthMatch
	}

	// CRC校验
	if message.CRC != message.GetCRC() {
		checkErr.ErrCrcCheck = config.ErrCrcCheck
	}

	// 检验是否全部校验都通过
	if !(checkErr.ErrCrcCheck == nil &&
		checkErr.ErrDataLengthMatch == nil &&
		checkErr.ErrSeqMatch == nil) {
		return nil, &checkErr
	}

	return &message, nil
}

// 解码(编码校验,此时不需要核验序列码)
func DecodeWithoutSeqNum(data []byte) (*Message, error) {
	checkErr := CheckError{
		ErrSeqMatch:        nil,
		ErrCrcCheck:        nil,
		ErrDataLengthMatch: nil,
	}

	var message Message
	err := json.Unmarshal(data, &message)
	// JSON反序列化失败,解码失败直接返回
	if err != nil {
		return nil, err
	}
	logrus.Debug(len(message.Data), message.Length)

	// 长度校验
	if len(message.Data) != message.Length {
		checkErr.ErrDataLengthMatch = config.ErrDataLengthMatch
	}

	// CRC校验
	if message.CRC != message.GetCRC() {
		checkErr.ErrCrcCheck = config.ErrCrcCheck
	}

	// 检验是否全部校验都通过
	if !(checkErr.ErrCrcCheck == nil &&
		checkErr.ErrDataLengthMatch == nil &&
		checkErr.ErrSeqMatch == nil) {
		return nil, &checkErr
	}

	return &message, nil
}

// CRC计算，赋给消息的CRC字段
func (m *Message) Encipher() {
	m.CRC = crc32.ChecksumIEEE(append(m.Data, config.PRIVATE_KEY[:]...))
}

// 解码获取CRC
func (m *Message) GetCRC() uint32 {
	return crc32.ChecksumIEEE(append(m.Data, config.PRIVATE_KEY[:]...))
}
