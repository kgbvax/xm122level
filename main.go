package main

//import "github.com/jacobsa/go-serial/serial"
import (
	"fmt"
	"github.com/creasty/defaults"
	"github.com/mikepb/go-serial"
	log "github.com/sirupsen/logrus"
	"github.com/vipally/binary"
)

const START_MARKER byte = 0xCC
const END_MARKER byte = 0xCD
const T_REG_READ byte = 0xF8
const T_REG_READ_RESP byte = 0xF6
const T_REG_WRITE byte = 0xF9

const REG_MODE_SELECTION byte = 0x02
const REG_MAIN_CONTROL byte = 0x03
const REG_STREAMING_CONTROL byte = 0x05
const REG_STATUS byte = 0x06
const REG_BAUDRATE byte = 0x07
const REG_POWER_MODE byte = 0x0A
const REG_PRODUCT_IDENTIFICATION byte = 0x10
const REG_PRODUCT_VERSION byte = 0x11
const REG_MAX_BAUDRATE byte = 0x12
const REG_OUTPUT_BUFFER_LENGTH byte = 0xE9

func findPort() *string {
	ports, err := serial.ListPorts()

	if err != nil {
		log.Panic(err)
	}

	for _, info := range ports {
		if info.Description() == "XB122" {
			log.Info("Found XB122 at " + info.Name())
			name := info.Name()
			return &name
		}
	}
	return nil
}

func main() {
	log.SetLevel(log.DebugLevel)

	portName := findPort()
	if portName != nil {
		options := serial.RawOptions
		options.BitRate = 115200
		options.Mode = serial.MODE_READ_WRITE

		p, err := options.Open(*portName)
		if err != nil {
			log.Panic(err)
		} else {
			log.Debug("opened port")
		}

		readRegisterRequest(p, REG_STATUS)
		regValue := readRegisterResponse(p)
		log.Info(fmt.Sprintf("Status 0x%x", regValue))
		if 0 != regValue&0x00100000 {
			log.Info("STATUS: Error activating the requested service or detector")
		}
		if 0 != regValue&0x00080000 {
			log.Info("STATUS: Error creating the requested service or detector.")
		}
		if 0 != regValue&0x00040000 {
			log.Info("STATUS: Invalid Mode.")
		}
		if 0 != regValue&0x00020000 {
			log.Info("STATUS: Invalid command or parameter received..")
		}
		if 0 != regValue&0x00010000 {
			log.Info("STATUS: An error occurred in the module.")
		}
		if 0 != regValue&0x00000100 {
			log.Info("STATUS: Data is ready to be read from the buffer")
		}
		if 0 != regValue&0x00000002 {
			log.Info("STATUS: Service or detector is activated.")
		}
		if 0 != regValue&0x00000001 {
			log.Info("STATUS: Service or detector is created.")
		}

		readRegisterRequest(p, REG_MAX_BAUDRATE)
		maxBaud := readRegisterResponse(p)
		log.Info("MaxBaud ", maxBaud)

		defer p.Close()
	}

}

type ResponseStateT byte

const (
	START ResponseStateT = iota
	PAYLOAD_LEN_LOW
	PAYLOAD_LEN_HIGH
	TYPE
	ADDRESS
	VALUE1
	VALUE2
	VALUE3
	VALUE4
	END
	DONE
)

type readRegRequest struct {
	StartMarker   byte   `default:"204"`
	PayloadLength uint16 `default:"0001"`
	RequestType   byte   `default:"248"`
	Register      byte
	EndMarker     byte `default:"205"`
}

func writeRegisterRequest(p *serial.Port, reg byte, val uint32) {
	var reqMsg [10] byte
	reqMsg[0] = START_MARKER
	reqMsg[1] = 0x05 //payload Length 2 bytes
	reqMsg[2] = 0x00
	reqMsg[3] = T_REG_WRITE
	reqMsg[4] = reg
	reqMsg[5] = byte(val >> 24)
	reqMsg[6] = byte(val >> 16)
	reqMsg[7] = byte(val >> 8)
	reqMsg[8] = byte(val & 0xff)
	reqMsg[9] = END_MARKER
}

func readRegisterRequest(p *serial.Port, reg byte) {

	req := &readRegRequest{}
	err := defaults.Set(req)
	if err != nil {
		log.Fatal(err)
	}

	req.Register = reg

	if bbuf, err := binary.Encode(req, nil); err == nil {
		_, err := p.Write(bbuf)
		if err != nil {
			log.Panic(err)
		}
		log.Trace(bbuf)
	}

	log.Debug("Send readRegisterRequest for reg ", reg)
}

func readRegisterResponse(p *serial.Port) uint32 {
	var buf [1] byte

	var value uint32

	expectState := START
	for expectState != DONE {
		_, err := p.Read(buf[:])
		if err != nil {
			log.Panic(err)
		}
		val := buf[0]
		log.Trace("ExpectState: ", expectState)
		switch expectState {

		case START:
			if val != START_MARKER {
				log.Error("expected start marker, got ", val)
			}
			expectState++
		case PAYLOAD_LEN_LOW:
			expectState++
		case PAYLOAD_LEN_HIGH:
			expectState++
		case TYPE:
			{
				if val != T_REG_READ_RESP {
					log.Error("expected start Register Read response, got ", val)
				}
				expectState++
			}
		case ADDRESS:
			{
				expectState++
			}

		case VALUE1:
			{
				value = uint32(val)
				expectState++
			}
		case VALUE2:
			{
				value |= uint32(val) << 8
				expectState++
			}
		case VALUE3:
			{
				value |= uint32(val) << 16
				expectState++
			}
		case VALUE4:
			{
				value |= uint32(val) << 24
				expectState++
			}
		case END:
			{
				if val != END_MARKER {
					log.Error("expected END MARKER   got", val)
				}
				expectState++
			}

		}
	}
	return value
}
