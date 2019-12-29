package main

import (
	"fmt"
	"github.com/creasty/defaults"
	"github.com/mikepb/go-serial"
	log "github.com/sirupsen/logrus"
	"github.com/vipally/binary"
	"time"
)

const (
	START_MARKER    byte = 0xCC
	END_MARKER      byte = 0xCD
	T_REG_READ      byte = 0xF8
	T_REG_READ_RESP byte = 0xF6
	T_REG_WRITE     byte = 0xF9
)

type RegT byte

const (
	REG_MODE_SELECTION         RegT = 0x02
	REG_MAIN_CONTROL           RegT = 0x03
	REG_STREAMING_CONTROL      RegT = 0x05
	REG_STATUS                 RegT = 0x06
	REG_BAUDRATE               RegT = 0x07
	REG_POWER_MODE             RegT = 0x0A
	REG_PRODUCT_IDENTIFICATION RegT = 0x10
	REG_PRODUCT_VERSION        RegT = 0x11
	REG_MAX_BAUDRATE           RegT = 0x12
	REG_OUTPUT_BUFFER_LENGTH   RegT = 0xE9
)

const (
	MODE_POWER_BINS          uint32 = 0x0001
	MODE_ENVELOPE            uint32 = 0x0002
	MODE_IQ                  uint32 = 0x0003
	MODE_SPARSE              uint32 = 0x0004
	MODE_DISTANCE_FIXED_PEAK uint32 = 0x0100
	MODE_OBSTACLE_DETECTION  uint32 = 0x0300
	MODE_PRESENCE_DETECTOR   uint32 = 0x0400
)

const (
	MAIN_STOP              uint32 = 0
	MAIN_CREATE_FLAG_ERR   uint32 = 0x0001
	MAIN_ACTIVATE_FLAG_ERR uint32 = 0x0002
	MAIN_CREATE_ACTIVATE   uint32 = 0x0003
	MAIN_CLEAR             uint32 = 0x0004
)

//Distance Peak Fixed RegisterS
const (
	DIST_RANGE_START            RegT = 0x20
	DIST_RANGE_LENGTH           RegT = 0x21
	DIST_REPETITION_MODE        RegT = 0x22
	DIST_UPDATE_RATE            RegT = 0x23
	DIST_GAIN                   RegT = 0x24
	DIST_SENSOR_POWER_MODE      RegT = 0x25
	DIST_TX_DISABLE             RegT = 0x26
	DIST_DECREASE_TX_EMISSION   RegT = 0x27
	DIST_PROFILE_SELECTION      RegT = 0x28
	DIST_DOWNSAMPLING_FACTOR    RegT = 0x29
	DIST_HW_ACC_AVERAGE_SAMPLES RegT = 0x30
	DIST_RUN_FACTOR             RegT = 0x40
	DIST_THR_AMPLITUDE          RegT = 0x42
	DIST_SORT_BY                RegT = 0x44
	DIST_START                  RegT = 0x81
	DIST_LENGTH                 RegT = 0x82
	DIST_DATA_SATURATED         RegT = 0xA0
	DIST_MISSED_DATA            RegT = 0xA1
	DIST_REFL_COUNT             RegT = 0xB0
	DIST_REFL_1_DISTANCE        RegT = 0xB1
	DIST_REFL_1_AMPLITUDE       RegT = 0xB2
	DIST_REFL_2_DISTANCE        RegT = 0xB3
	DIST_REFL_2_AMPLITUDE       RegT = 0xB4
	DIST_REFL_3_DISTANCE        RegT = 0xB5
	DIST_REFL_3_AMPLITUDE       RegT = 0xB6
	DIST_REFL_4_DISTANCE        RegT = 0xB7
	DIST_REFL_5_AMPLITUDE       RegT = 0xB8
)

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

		checkStatus(p)

		maxBaud := readRegister(p, REG_MAX_BAUDRATE)
		log.Info("MaxBaud ", maxBaud)

		writeRegister(p, REG_POWER_MODE, 0)
		checkStatus(p)

		writeRegister(p, REG_MODE_SELECTION, MODE_DISTANCE_FIXED_PEAK)
		checkStatus(p)

		writeRegister(p, DIST_RANGE_START, 41)
		checkStatus(p)

		writeRegister(p, DIST_RANGE_LENGTH, 96)
		checkStatus(p)

		writeRegister(p, REG_MAIN_CONTROL, MAIN_CREATE_FLAG_ERR)
		checkStatus(p)

		writeRegister(p, REG_MAIN_CONTROL, MAIN_ACTIVATE_FLAG_ERR)
		checkStatus(p)

		for {
			time.Sleep(1 * time.Second)
			if checkStatus(p)&StatusDataReady != 0 {
				reflCount := readRegister(p, DIST_REFL_COUNT)
				log.Debug("refl count ", reflCount)
			}

		}
		defer p.Close()
	}
}

type readRegRequest struct {
	StartMarker   byte   `default:"204"`
	PayloadLength uint16 `default:"0001"`
	RequestType   byte   `default:"248"`
	Register      RegT
	EndMarker     byte `default:"205"`
}

type writeRegRequest struct {
	StartMarker   byte   `default:"204"`
	PayloadLength uint16 `default:"0005"`
	RequestType   byte   `default:"249"`
	Register      RegT
	Value         uint32
	EndMarker     byte `default:"205"`
}

type writeRegResponse struct {
	StartMarker   byte
	PayloadLength uint16
	RequestType   byte
	Register      RegT
	Value         uint32
	EndMarker     byte
}

type readRegResponse struct {
	StartMarker   byte
	PayloadLength uint16
	RequestType   byte
	Register      RegT
	Value         uint32
	EndMarker     byte
}

func writeRegister(p *serial.Port, reg RegT, val uint32) uint32 {
	req := &writeRegRequest{Value: val, Register: reg}
	err := defaults.Set(req)
	if err != nil {
		log.Fatal(err)
	}

	if bbuf, err := binary.Encode(req, nil); err == nil {
		_, err := p.Write(bbuf)
		if err != nil {
			log.Panic(err)
		}
		log.Trace(bbuf)
	} else {
		log.Panic(err)
	}

	log.Debug("set reg ", reg, " to ", val)

	resp := &writeRegResponse{}
	sz := binary.Size(resp)
	buffer := make([]byte, sz)
	numRead, err := p.Read(buffer)
	if numRead != sz {
		log.Warn("did not recieve expected data, got vs expected: ", numRead, " ", sz)
	}
	if err != nil {
		log.Error(err)
	}
	err = binary.Decode(buffer, resp)
	if err != nil {
		log.Error(err)
	}
	return resp.Value

}

func readRegister(p *serial.Port, reg RegT) uint32 {
	req := &readRegRequest{Register: reg}
	err := defaults.Set(req)
	if err != nil {
		log.Fatal(err)
	}

	if bbuf, err := binary.Encode(req, nil); err == nil {
		_, err := p.Write(bbuf)
		if err != nil {
			log.Panic(err)
		}
		log.Trace(bbuf)
	} else {
		log.Panic(err)
	}

	resp := &readRegResponse{}
	sz := binary.Size(resp)
	buffer := make([]byte, sz)
	numRead, err := p.Read(buffer)
	if numRead != sz {
		log.Warn("did not recieve expected data, got vs expected: ", numRead, " ", sz)
	}
	if err != nil {
		log.Error(err)
	}
	err = binary.Decode(buffer, resp)
	if err != nil {
		log.Error(err)
	}
	return resp.Value
}

const (
	StatusErrActivating  = 0x00100000
	StatusErrCreating    = 0x00080000
	StatusInvalidMode    = 0x00040000
	StatusInvalidCommand = 0x00020000
	StatusDataReady      = 0x00000100
	StatusError          = 0x00010000
	StatusServActivated  = 0x00000002
	StatusServ           = 0x00000001
)

func checkStatus(p *serial.Port) uint32 {
	regValue := readRegister(p, REG_STATUS)
	log.Trace(fmt.Sprintf("Status 0x%x", regValue))
	if 0 != regValue&StatusErrActivating {
		log.Debug("STATUS: Error activating the requested service or detector")
	}
	if 0 != regValue&StatusErrCreating {
		log.Debug("STATUS: Error creating the requested service or detector.")
	}
	if 0 != regValue&StatusInvalidMode {
		log.Debug("STATUS: Invalid Mode.")
	}
	if 0 != regValue&StatusInvalidCommand {
		log.Debug("STATUS: Invalid command or parameter received..")
	}
	if 0 != regValue&StatusError {
		log.Debug("STATUS: An error occurred in the module.")
	}
	if 0 != regValue&StatusDataReady {
		log.Debug("STATUS: Data is ready to be read from the buffer")
	}
	if 0 != regValue&StatusServActivated {
		log.Debug("STATUS: Service or detector is activated.")
	}
	if 0 != regValue&StatusServ {
		log.Debug("STATUS: Service or detector is created.")
	}

	return regValue
}
