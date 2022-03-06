package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

var mutex sync.Mutex = sync.Mutex{}

const sampleRate = 44100
const seconds = 2

type Command struct {
	Name string
	Sarg string
}
type Resp struct {
	Name   string
	Result string
}

type Frame struct {
	Buff [sampleRate * seconds]float32
}

var Rooms map[string]map[string]bool = make(map[string]map[string]bool, 0)

func RoomsList(m map[string]map[string]bool) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return "[" + strings.Join(keys, ", ") + "]"
}

func RoomsParticipants(m map[string]bool) string {
	keys := make([]string, 0, len(m))
	for k, val := range m {
		if val {
			keys = append(keys, k)
		}
	}
	return "[" + strings.Join(keys, ", ") + "]"
}

var Nicknames map[string]bool = make(map[string]bool, 0)
var encoders map[string]*gob.Encoder = make(map[string]*gob.Encoder, 0)

func handle_conn(conn net.Conn) {
	defer conn.Close()
	dec := gob.NewDecoder(conn)
	enc := gob.NewEncoder(conn)
	var nick string = ""
	var room string = ""
	var com Command
	for {
		err := dec.Decode(&com)
		if err != nil {
			com.Name = "disconnect"
		}
		mutex.Lock()
		switch com.Name {
		case "room_list":
			enc.Encode(Resp{
				Name:   "room_list",
				Result: "list of existing rooms: " + RoomsList(Rooms),
			})
		case "participants_list":
			pref := "list of users in room " + room + ": "
			res := pref + "[]"
			if room != "" && len(Rooms[room]) != 0 {
				res = pref + RoomsParticipants(Rooms[room])
			}
			if room == "" {
				res = "Enter the room first to ask participants list"
			}
			enc.Encode(Resp{
				Name:   "participants_list",
				Result: res,
			})

		case "disconnect":
			if nick != "" {
				Nicknames[nick] = false
				if room != "" {
					Rooms[room][nick] = false
				}
			}
			mutex.Unlock()
			return

		case "connect":
			if Nicknames[com.Sarg] {
				enc.Encode(Resp{
					Name:   "connect",
					Result: "Name already in use",
				})

			} else {
				fmt.Println(com.Sarg)
				Nicknames[com.Sarg] = true
				nick = com.Sarg
				encoders[nick] = enc
				enc.Encode(Resp{
					Name:   "connect",
					Result: "Succes",
				})
			}

		case "connect_room":
			if Rooms[com.Sarg] == nil {
				Rooms[com.Sarg] = map[string]bool{}
			}
			room = com.Sarg
			Rooms[com.Sarg][nick] = true
			fmt.Println("also in the room: " + RoomsParticipants(Rooms[com.Sarg]))
			enc.Encode(Resp{
				Name:   "connect",
				Result: "currently in the room: " + RoomsParticipants(Rooms[com.Sarg]),
			})
			for key, val := range Rooms[room] {
				if val {
					encoders[key].Encode(Resp{Name: "text", Result: nick + " entered the room"})
				}
			}

		case "disconnect_room":
			if nick != "" {
				if room != "" {
					Rooms[room][nick] = false

					for key, val := range Rooms[room] {
						if val {
							encoders[key].Encode(Resp{Name: "text", Result: nick + " left the room"})
						}
					}
				}
				room = ""
			}
		case "text":
			if nick != "" {
				if room != "" {
					for key, val := range Rooms[room] {
						if val {
							encoders[key].Encode(Resp{Name: "text", Result: nick + ":" + com.Sarg})
						}
					}
				}
			}

		case "mute":

		case "unmute":
		}
		mutex.Unlock()
	}
}

func main() {
	//portaudio.Initialize()
	//defer portaudio.Terminate()
	//buffer := make([]float32, sampleRate*seconds)

	fmt.Println("Enter port to listen on")
	reader := bufio.NewReader(os.Stdin)
	server, _ := reader.ReadString('\n')
	server = server[:len(server)-1]
	ln, _ := net.Listen("tcp", server)
	defer ln.Close()

	for {
		conn, _ := ln.Accept()
		go handle_conn(conn)
	}
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}
