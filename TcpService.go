// TcpService.go
package main

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"

	"github.com/widuu/gojson"
)

var (
	maxRead  = 4096
	msgOK    = []byte(`{"result":"ok"}`)
	msgError = []byte(`{"result":"error"}`)
)

// StartTCPServer start
func StartTCPServer() {

	hostAndPort := ":12100"
	listener := initServer(hostAndPort)
	for {
		conn, err := listener.Accept()
		checkError(err, "Accept: ")
		go connectionHandler(conn)
	}
}
func initServer(hostAndPort string) *net.TCPListener {
	serverAddr, err := net.ResolveTCPAddr("tcp", hostAndPort)
	checkError(err, "Resolving address:port failed: '"+hostAndPort+"'")
	listener, err := net.ListenTCP("tcp", serverAddr)
	checkError(err, "ListenTCP: ")
	println("Listening to: ", listener.Addr().String())
	return listener
}
func connectionHandler(conn net.Conn) {
	connFrom := conn.RemoteAddr().String()
	println("Connection from: ", connFrom)
	for {
		var ibuf []byte = make([]byte, maxRead+1)
		length, err := conn.Read(ibuf[0:maxRead])
		ibuf[maxRead] = 0 // to prevent overflow
		switch err {
		case nil:
			handleMsg(conn, length, err, ibuf)
		default:
			goto DISCONNECT
		}
	}
DISCONNECT:
	err := conn.Close()
	println("Closed connection:", connFrom)
	checkError(err, "Close:")
}
func talktoclients(to net.Conn, msg []byte) {
	wrote, err := to.Write(msg)
	checkError(err, "Write: wrote "+string(wrote)+" bytes.")
}
func handleMsg(conn net.Conn, length int, err error, msg []byte) {
	if length > 0 {
		cmds := gojson.Json(string(msg[:length])).Getdata()
		if val, ok := cmds["cmd"]; ok {
			talktoclients(conn, msgOK)
			//do something here
			if strings.EqualFold(val.(string), "start") {
				var name, folder, sdevname string
				var index, label int
				Is512Sector := false

				if value, ok := cmds["s512"]; ok {
					//Is512Sector = value.(bool)
					switch value.(type) {
					case string:
						Is512Sector, err = strconv.ParseBool(value.(string))
						if err != nil {
							Is512Sector = false
						}
					case bool:
						Is512Sector = value.(bool)

					default:
						Is512Sector = false
					}
					//					Is512Sector, err = strconv.ParseBool(value.(string))
					//					if err != nil {
					//						Is512Sector = false
					//					}
				}

				if value, ok := cmds["name"]; ok {
					name = value.(string)
				}
				if value, ok := cmds["folder"]; ok {
					folder = value.(string)
				}
				if value, ok := cmds["device"]; ok {
					sdevname = value.(string)
					//must run umount if system
					if len(sdevname) > 0 {
						exec.Command("umount", sdevname).Output()
					}
				}
				if value, ok := cmds["index"]; ok {
					//index = int(value.(float64))
					switch value.(type) {
					case string:
						index, _ = strconv.Atoi(value.(string))
					case int:
						index = int(value.(int))
					case int64:
						index = int(value.(int64))
					case float64:
						index = int(value.(float64))
					default:
						index = 1000
					}

				}
				if value, ok := cmds["label"]; ok {
					//label = int(value.(float64))
					//label, _ = strconv.Atoi(value.(string))
					switch value.(type) {
					case string:
						label, _ = strconv.Atoi(value.(string))
					case int:
						label = int(value.(int))
					case int64:
						label = int(value.(int64))
					case float64:
						label = int(value.(float64))
					default:
						label = 1000
					}
				}
				fmt.Printf("%v_%s_%s_%s_%d_%d\n", Is512Sector, name, folder, sdevname, index, label)
				if name == "SecureErase" {
					go RunSecureErase(folder, sdevname, label)
					return
				}
				if Is512Sector && len(sdevname) > 0 {
					profile, err := configxmldata.FindProfileByName(name)
					if err != nil {
						return
					}
					patten := profile.CreatePatten()
					go RunWipe(folder, sdevname, patten, label)
				} else {
					profile, err := configxmldata.FindProfileByName(name)
					if err != nil {
						return
					}
					patten := profile.CreatePatten()
					sdevname = fmt.Sprintf("/dev/sg%d", index)
					go RunWipe(folder, sdevname, patten, label)
				}

			} else if strings.EqualFold(val.(string), "stop") {
				var label int
				if value, ok := cmds["label"]; ok {
					//label = int(value.(float64))
					//label, _ = strconv.Atoi(value.(string))
					switch value.(type) {
					case string:
						label, _ = strconv.Atoi(value.(string))
					case int:
						label = int(value.(int))
					case int64:
						label = int(value.(int64))
					case float64:
						label = int(value.(float64))

					default:
						label = 1000
					}
				}
				processlist.Remove(label)
			}
		} else {
			talktoclients(conn, msgError)
		}

	} else {
		talktoclients(conn, msgError)
	}
}
func checkError(error error, info string) {
	if error != nil {
		fmt.Sprintln("ERROR: " + info + " " + error.Error())
		panic("ERROR: " + info + " " + error.Error()) // terminate program
	}
}
