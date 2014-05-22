package gocom

//#include "win_com.h"
import "C"

import "unsafe"

import (
	"fmt"
	"strconv"
)

const (
	B115200 = 115200
	B19200  = 19200
	B9600   = 9600
	//8 data bits
	AN8 = 8
	//1stop bit
	SB1 = 1
	//no Parity Check
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
	//read from com，
	//return data len and err. if ok err is nil
	Read(out []byte) (n int, err error)
	//read n bytes from com，
	//return data len and err. if ok err is nil
	Readn(out []byte, l int) (n int, err error)
	//read one record from com，
	//return data len and err. if ok err is nil
	//the record fmt:STX[1byte] + len[2byte] + data + ETX[1byte] + lrc[1byte]
	//retuen data len(not record len)
	ReadRecord(out []byte) (n int, err error)
	//write to com，
	//return data len and err. if ok err is nil
	Write(in []byte) (n int, err error)
	//write n bytes to com，
	//return data len and err. if ok err is nil
	Writen(in []byte, l int) (n int, err error)

	//write one record to com，
	//return data len and err. if ok err is nil
	//the record fmt:STX[1byte] + len[2byte] + data + ETX[1byte] + lrc[1byte]
	//retuen data len(not record len)
	WriteRecord(in []byte) (n int, err error)

	//set com ,see ComInfo
	Set(info *ComInfo) (err error)

	//set RDWR timeout (seconds)
	//-1 timeout forever
	//0 nonblock
	SetDeadline(timeout int) (err error)

	//set RD timeout (seconds)
	//-1 timeout forever
	//0 nonblock
	SetReadDeadline(timeout int) (err error)

	//set WR timeout (seconds)
	//-1 timeout forever
	//0 nonblock
	SetWriteDeadline(timeout int) (err error)

	Close() (err error)
}

type ComInfo struct {
	//BaudRate
	BaudRate int
	//DataBit
	DataBit int
	//Parity
	Parity int
	//StopBits
	StopBits int
}

type WinCom struct {
	fd          int
	read_block  bool
	write_block bool
}

func (com *WinCom) Read(out []byte) (int, error) {

	n := C.com_read(C.int(com.fd),
		unsafe.Pointer(&out), C.int(len(out)))

	if n < 0 {
		return -1, &WinComErr{"Read Com", "fail"}
	}
	if n == 0 && com.read_block == true {
		return -1, &WinComErr{"Read Com", "timeout"}
	}
	return int(n), nil
}
func (com *WinCom) Write(in []byte) (int, error) {
	n := C.com_write(C.int(com.fd),
		unsafe.Pointer(&in), C.int(len(in)))
	if n < 0 {
		return -1, &WinComErr{"Write Com", "fail"}
	}
	if n == 0 && com.write_block == true {
		return -1, &WinComErr{"Write Com", "timeout"}
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
		n, err = com.Read(out[nread:l])
		if err != nil {
			return -1, err
		}
		if n == 0 {
			return nread, nil
		}
		nread += n
		if nread == l {
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
		if n == l {

			return l, nil
		}
		if n == 0 {
			return nwrite, nil
		}
		nwrite += n
	}
}

func lrc() byte {
	return 'a'
}
func (com *WinCom) ReadRecord(out []byte) (int, error) {
	buf := make([]byte, 2)

	//STX
	n, err := com.Readn(buf, 1)
	if err != nil {
		return -1, err
	}

	if n != 1 || buf[0] != STX {
		return -1, &WinComErr{"Read Com", "STX not found"}
	}

	//len
	n, err = com.Readn(buf, 2)
	if err != nil {
		return -1, err
	}
	if n != 2 {
		return -1, &WinComErr{"Read Com", "LEN not found"}
	}

	//data
	l, _ := strconv.Atoi(fmt.Sprintf("%x", buf[0:2]))
	data := make([]byte, l)
	n, err = com.Readn(data, l)
	if err != nil {
		return -1, err
	}
	if n != l {
		return -1, &WinComErr{"Read Com", "DATA err"}
	}

	//ETX
	n, err = com.Readn(buf, 1)
	if err != nil {
		return -1, err
	}
	if n != 1 || buf[0] != ETX {
		return -1, &WinComErr{"Read Com", "ETX not found"}
	}

	//lrc
	n, err = com.Readn(buf, 1)
	if err != nil {
		return -1, err
	}
	if n != 1 {
		return -1, &WinComErr{"Read Com", "LRC not found"}
	}
	//if lrc() != buf[0] {
	//	return -1, &WinComErr{"Read Com", "LRC err"}
	//}
	ncopy := copy(out, data[0:l])
	if ncopy != l {
		return -1, &WinComErr{"Read Com", "Recvbuf too small"}
	}

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
		return -1, &WinComErr{"Write Com", "Write record err"}
	}
	return len(in), nil
}

func (com *WinCom) Set(info *ComInfo) error {
	var dcb C.DCB
	dcb.BaudRate = C.DWORD(info.BaudRate)
	dcb.ByteSize = C.BYTE(info.DataBit)
	if info.Parity == NPBit {
		dcb.Parity = C.NOPARITY
	}
	if info.StopBits == SB1 {
		dcb.StopBits = C.ONESTOPBIT
	}
	ret := C.com_set(C.int(com.fd), &dcb)
	if ret < 0 {
		return &WinComErr{"Set Com", "fail"}
	}
	return nil
}

func (com *WinCom) SetDeadline(timeout int) error {
	if timeout < 0 || timeout > 0 {
		com.read_block = true
		com.write_block = true
	}
	if timeout == 0 {
		com.read_block = false
		com.write_block = false
	}
	if timeout < 0 {
		timeout = -1
	}

	ret := C.com_set_deadline(C.int(com.fd), C.int(timeout))
	if ret < 0 {
		return &WinComErr{"SetDeadline", "fail"}
	}
	return nil
}

func (com *WinCom) SetReadDeadline(timeout int) error {
	if timeout < 0 || timeout > 0 {
		com.read_block = true
	}
	if timeout == 0 {
		com.read_block = false
	}
	if timeout < 0 {
		timeout = -1
	}
	ret := C.com_set_read_deadline(C.int(com.fd), C.int(timeout))
	if ret < 0 {
		return &WinComErr{"SetReadDeadline", "fail"}
	}
	return nil
}

func (com *WinCom) SetWriteDeadline(timeout int) error {
	if timeout < 0 || timeout > 0 {
		com.write_block = true
	}
	if timeout == 0 {
		com.write_block = false
	}
	if timeout < 0 {
		timeout = -1
	}
	ret := C.com_set_write_deadline(C.int(com.fd), C.int(timeout))
	if ret < 0 {
		return &WinComErr{"SetWriteDeadline", "fail"}
	}
	return nil
}

func (com *WinCom) Close() (err error) {
	ret := C.com_close(C.int(com.fd))
	if ret == 0 {
		return nil
	}
	return &WinComErr{"Close Com", "fail"}
}

func Open(cnum int) (Com, error) {
	com := &WinCom{}
	fd := C.com_open(C.int(cnum))
	if fd < 0 {
		return nil, &WinComErr{"Open Com", "Port not found or be occupied"}
	}
	com.fd = int(fd)
	com.read_block = true
	com.write_block = true
	return com, nil
}
