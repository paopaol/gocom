package gocom

//#include "win_com.h"
import "C"

import "unsafe"

import (
	"fmt"
	"strconv"
)

//串口可设置参数
const (
	//波特率s
	B115200 = 115200
	B19200  = 19200
	B9600   = 9600
	//8个数据位
	AN8 = 8
	//1个停止位
	SB1 = 1
	//无奇偶校验位
	NPBit = 0
)
const (
	STX = 0x02
	ETX = 0x03
)

type WinComErr struct {
	ErrOpt  string
	ErrInfo string
}

func (err *WinComErr) Error() string {
	return fmt.Sprintf("%s:%s", err.ErrOpt, err.ErrInfo)
}

type Com interface {
	//从串口读数据，
	//成功:返回读到的数据及数据长度，err 为nil
	//失败:返回-1，and err not nil
	Read(out []byte) (n int, err error)
	//同Read
	//从串口读数据，
	//成功:返回读到的实际数据长度或期望长度，err 为nil
	//失败:返回-1，and err not nil
	Readn(out []byte, l int) (n int, err error)

	//同Read
	//带格式的从串口读一条记录,
	//格式:STX[1byte] + len[2byte] + data + ETX[1byte] + lrc[1byte]
	//成功返回数据段内容,以及数据段长度, nil
	//失败：返回-1，and err not nil
	ReadRecord(out []byte) (n int, err error)

	//从串口写数据，
	//成功:返回写入的数据及数据长度，err 为nil
	//失败:返回-1，and err not nil
	Write(in []byte) (n int, err error)
	Writen(in []byte, l int) (n int, err error)

	//同Write
	//带格式的串口写入一条记录,
	//格式:STX[1byte] + len[2byte] + data + ETX[1byte] + lrc[1byte]
	//成功返回写入数据段长度, nil
	//失败：返回-1，and err not nil
	WriteRecord(in []byte) (n int, err error)

	//设置串口参数，可设置内容见ComInfo
	Set(info *ComInfo) (err error)

	//设置串口的读写超时时间,单位(秒)
	//-1为永久阻塞
	//0为非阻塞
	SetDeadline(timeout int) (err error)

	//同SetDeadline
	//设置串口的读超时时间
	//-1为永久阻塞
	//0为非阻塞
	SetReadDeadline(timeout int) (err error)

	//同SetDeadline
	//设置串口的写超时时间
	//-1为永久阻塞
	//0为非阻塞
	SetWriteDeadline(timeout int) (err error)

	//关闭串口
	Close() (err error)
}

type ComInfo struct {
	//波特率
	BaudRate int
	//每个字节有多少位
	DateBit int
	//有无奇偶校验位
	Parity int
	//几个停止位
	StopBits int
}

type WinCom struct {
	fd          int
	read_block  bool
	write_block bool
}

func (com *WinCom) Read(out []byte) (int, error) {
	n := C.com_read(C.int(com.fd),
		unsafe.Pointer(&out), C.int(cap(out)))
	if n < 0 {
		return -1, &WinComErr{"读串口", "失败"}
	}
	if n == 0 && com.read_block == true {
		return -1, &WinComErr{"读串口", "超时"}
	}
	return int(n), nil
}
func (com *WinCom) Write(in []byte) (int, error) {
	n := C.com_write(C.int(com.fd),
		unsafe.Pointer(&in), C.int(len(in)))
	if n < 0 {
		return -1, &WinComErr{"写串口", "失败"}
	}
	if n == 0 && com.write_block == true {
		return -1, &WinComErr{"写串口", "超时"}
	}
	return int(n), nil
}

func (com *WinCom) Readn(out []byte, l int) (int, error) {
	var nread int = 0
	var n int = 0
	var err error

	if l <= 0 {
		return 0, nil
	}

	for {
		n, err = com.Read(out[nread:])
		if err != nil {
			return -1, err
		}
		nread += n
		//如果读够指定数量，则返回
		if nread == l {
			return nread, nil
		}
		//数据没有了，返回实际读到的长度
		if n == 0 {
			return nread, nil
		}
	}
}
func (com *WinCom) Writen(in []byte, l int) (int, error) {
	var nwrite int = 0
	var n int = 0
	var err error

	if l <= 0 {
		return 0, nil
	}

	for {
		n, err = com.Write([]byte(in[nwrite:l]))
		if err != nil {
			return -1, err
		}
		
		//正好
		if n == l {
			
			return l, nil
		}
		//没了，写完了,返回实际写入的数据
		if n == 0 {
			return nwrite, nil
		}
		//已经写了这么多数据
		nwrite += n
	}
}

func lrc() byte {
	return 'a'
}
func (com *WinCom) ReadRecord(out []byte) (int, error) {
	buf := make([]byte, 2)

	//读STX
	n, err := com.Readn(buf, 1)
	if err != nil {
		return -1, err
	}
	if n != 1 || buf[0] != STX {
		return -1, &WinComErr{"读串口", "头部字段有误"}
	}

	//读len
	n, err = com.Readn(buf, 2)
	if err != nil {
		return -1, err
	}
	if n != 2 {
		return -1, &WinComErr{"读串口", "长度字段有误"}
	}

	//读data
	l, _ := strconv.Atoi(fmt.Sprintf("%x", buf[0:2]))
	data := make([]byte, l)
	n, err = com.Readn(data, l)
	if err != nil {
		return -1, err
	}
	if n != l {
		return -1, &WinComErr{"读串口", "数据字段有误"}
	}

	//读ETX
	n, err = com.Readn(buf, 1)
	if err != nil {
		return -1, err
	}
	if n != 1 || buf[0] != ETX {
		return -1, &WinComErr{"读串口", "结束字段有误"}
	}

	//读lrc
	n, err = com.Readn(buf, 1)
	if err != nil {
		return -1, err
	}
	if n != 1 {
		return -1, &WinComErr{"读串口", "校验字段字段丢失"}
	}
	//if lrc() != buf[0] {
	//	return -1, &WinComErr{"读串口", "校验字段字段有误"}
	//}
	//返回数据段数据
	ncopy := copy(out, data[0:l])
	if ncopy != l {
		return -1, &WinComErr{"读串口", "接收缓冲区过小"}
	}
	//返回数据段长度
	return l, nil
}

func (com *WinCom) WriteRecord(in []byte) (int, error) {
	//STX + len + data + ETX + lrc
	record_len := 1 + 2 + len(in) + 1 + 1
	record := make([]byte, record_len)

	record[0] = STX
	record[1] = byte(len(in) / 256)
	record[2] = byte(len(in) % 256)
	copy(record[3:], in)
	record[3+len(in)] = ETX
	record[4+len(in)] = lrc()

	n, err := com.Writen(record[0:record_len], record_len)
	if err != nil {
		return -1, nil
	}
	if n != record_len {
		return -1, &WinComErr{"写串口", "写入数据不完整"}
	}
	//返回数据段长度
	return len(in), nil
}

//注：此函数还不完善，目前仅支持设置的参数，见const()串口可设置参数
func (com *WinCom) Set(info *ComInfo) error {
	var dcb C.DCB
	dcb.BaudRate = C.DWORD(info.BaudRate)
	dcb.ByteSize = C.BYTE(info.DateBit)
	if info.Parity == NPBit {
		dcb.Parity = C.NOPARITY
	}
	if info.StopBits == SB1 {
		dcb.StopBits = C.ONESTOPBIT
	}
	ret := C.com_set(C.int(com.fd), &dcb)
	if ret < 0 {
		return &WinComErr{"设置串口", "失败"}
	}
	return nil
}

//设置掉线临界点(即，设置读写的超时时间)
func (com *WinCom) SetDeadline(timeout int) error {
	//阻塞
	if timeout < 0 || timeout > 0 {
		com.read_block = true
		com.write_block = true
	}
	//非阻塞
	if timeout == 0 {
		com.read_block = false
		com.write_block = false
	}
	if timeout < 0 {
		timeout = -1
	}

	ret := C.com_set_deadline(C.int(com.fd), C.int(timeout))
	if ret < 0 {
		return &WinComErr{"设置串口读写超时时间", "失败"}
	}
	return nil
}

//设置读超时时间
func (com *WinCom) SetReadDeadline(timeout int) error {
	//阻塞
	if timeout < 0 || timeout > 0 {
		com.read_block = true
	}
	//非阻塞
	if timeout == 0 {
		com.read_block = false
	}
	if timeout < 0 {
		timeout = -1
	}
	ret := C.com_set_read_deadline(C.int(com.fd), C.int(timeout))
	if ret < 0 {
		return &WinComErr{"设置串口读超时时间", "失败"}
	}
	return nil
}

//设置写超时时间
func (com *WinCom) SetWriteDeadline(timeout int) error {
	//阻塞
	if timeout < 0 || timeout > 0 {
		com.write_block = true
	}
	//非阻塞
	if timeout == 0 {
		com.write_block = false
	}
	if timeout < 0 {
		timeout = -1
	}
	ret := C.com_set_write_deadline(C.int(com.fd), C.int(timeout))
	if ret < 0 {
		return &WinComErr{"设置串口写超时时间", "失败"}
	}
	return nil
}

func (com *WinCom) Close() (err error) {
	ret := C.com_close(C.int(com.fd))
	if ret == 0 {
		return nil
	}
	return &WinComErr{"关闭串口", "失败"}
}

//默认读写无超时时间，永久阻塞
func Open(cnum int) (Com, error) {
	com := &WinCom{}
	fd := C.com_open(C.int(cnum))
	if fd < 0 {
		return nil, &WinComErr{"打开串口", "端口不存在或已被占用"}
	}
	com.fd = int(fd)
	//永久阻塞
	com.read_block = true
	com.write_block = true
	return com, nil
}
