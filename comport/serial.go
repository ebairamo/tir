package comport

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procCreateFile      = kernel32.NewProc("CreateFileW")
	procCloseHandle     = kernel32.NewProc("CloseHandle")
	procWriteFile       = kernel32.NewProc("WriteFile")
	procReadFile        = kernel32.NewProc("ReadFile")
	procSetCommState    = kernel32.NewProc("SetCommState")
	procSetCommTimeouts = kernel32.NewProc("SetCommTimeouts")
	procPurgeComm       = kernel32.NewProc("PurgeComm")
)

type DCB struct {
	DCBlength, BaudRate                            uint32
	flags                                          [4]byte
	wReserved, XonLim, XoffLim                     uint16
	ByteSize, Parity, StopBits                     byte
	XonChar, XoffChar, ErrorChar, EofChar, EvtChar byte
	wReserved1                                     uint16
}

type COMMTIMEOUTS struct {
	ReadIntervalTimeout         uint32
	ReadTotalTimeoutMultiplier  uint32
	ReadTotalTimeoutConstant    uint32
	WriteTotalTimeoutMultiplier uint32
	WriteTotalTimeoutConstant   uint32
}

const (
	PURGE_TXABORT        = 0x0001
	PURGE_RXABORT        = 0x0002
	PURGE_TXCLEAR        = 0x0004
	PURGE_RXCLEAR        = 0x0008
	INVALID_HANDLE_VALUE = ^uintptr(0)
)

// OpenPort открывает COM-порт
func OpenPort(portName string) (syscall.Handle, error) {
	path := syscall.StringToUTF16Ptr("\\\\.\\" + portName)
	handle, _, err := procCreateFile.Call(
		uintptr(unsafe.Pointer(path)),
		syscall.GENERIC_READ|syscall.GENERIC_WRITE,
		0,
		0,
		syscall.OPEN_EXISTING,
		0,
		0)

	if handle == INVALID_HANDLE_VALUE {
		return 0, os.NewSyscallError("CreateFile", err)
	}

	return syscall.Handle(handle), nil
}

// ClosePort закрывает COM-порт
func ClosePort(handle syscall.Handle) {
	procCloseHandle.Call(uintptr(handle))
}

// SetCommParams устанавливает параметры COM-порта
func SetCommParams(handle syscall.Handle, baudRate uint32) error {
	var dcb DCB
	dcb.DCBlength = uint32(unsafe.Sizeof(dcb))
	dcb.BaudRate = baudRate
	dcb.ByteSize = 8
	dcb.Parity = 0
	dcb.StopBits = 0

	r, _, err := procSetCommState.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(&dcb)))

	if r == 0 {
		return os.NewSyscallError("SetCommState", err)
	}

	return nil
}

// SetCommTimeouts устанавливает таймауты для COM-порта
func SetCommTimeouts(handle syscall.Handle) error {
	var timeouts COMMTIMEOUTS
	timeouts.ReadIntervalTimeout = 0
	timeouts.ReadTotalTimeoutMultiplier = 1
	timeouts.ReadTotalTimeoutConstant = 2
	timeouts.WriteTotalTimeoutMultiplier = 0
	timeouts.WriteTotalTimeoutConstant = 0

	r, _, err := procSetCommTimeouts.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(&timeouts)))

	if r == 0 {
		return os.NewSyscallError("SetCommTimeouts", err)
	}

	return nil
}

// PurgeComm очищает буферы COM-порта
func PurgeComm(handle syscall.Handle) error {
	r, _, err := procPurgeComm.Call(
		uintptr(handle),
		uintptr(PURGE_RXCLEAR|PURGE_TXCLEAR))

	if r == 0 {
		return os.NewSyscallError("PurgeComm", err)
	}

	return nil
}

// WritePort записывает данные в COM-порт
func WritePort(handle syscall.Handle, buf []byte) (uint32, error) {
	var written uint32
	r, _, err := procWriteFile.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
		uintptr(unsafe.Pointer(&written)),
		0)

	if r == 0 {
		return 0, os.NewSyscallError("WriteFile", err)
	}

	return written, nil
}

// ReadPort читает данные из COM-порта
func ReadPort(handle syscall.Handle, buf []byte) (uint32, error) {
	var read uint32
	r, _, err := procReadFile.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
		uintptr(unsafe.Pointer(&read)),
		0)

	if r == 0 {
		return 0, os.NewSyscallError("ReadFile", err)
	}

	return read, nil
}
