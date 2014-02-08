package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	bitstring "github.com/ymotongpoo/go-bitstring"
)

const UtmpxFile = "/var/run/utmpx"
const EntrySize = uint64(628)

type Utmpx struct {
	User []byte `binary:"256"`
	Id   []byte `binary:"4"`
	Line []byte `binary:"32"`
}

func (u *Utmpx) String() string {
	return string(u.User) + " " + string(u.Id) + " " + string(u.Line)
}

func main() {
	file, err := os.Open(UtmpxFile)
	if err != nil {
		log.Fatalln("Error occured during opening file: ", err.Error())
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalln(err.Error())
	}
	dataBuf := bytes.NewBuffer(data)
	buf := bitstring.NewBuffer(dataBuf)

	for {
		data, err := buf.PopBytes(EntrySize)
		if err != nil {
			if err != io.EOF {
				log.Fatalln("Error: ", err.Error())
			}
		}

		b := bitstring.NewBuffer(bytes.NewBuffer(data))
		u := &Utmpx{}

		marshalErr := bitstring.Unmarshal(b, u)
		if marshalErr != nil {
			if marshalErr != io.EOF {
				log.Fatalln("Error: ", err.Error())
			}
		}
		fmt.Println(u.String())
		if err == io.EOF {
			break
		}
	}
}
