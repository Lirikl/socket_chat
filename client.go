package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"net"
	"os"
)

type Command struct {
	Name string
	Sarg string
}
type Resp struct {
	Name   string
	Result string
}

func listen(dec *gob.Decoder) {
	var resp Resp
	for {
		err := dec.Decode(&resp)
		if err != nil {
			return
		}
		fmt.Println(resp.Result)
	}
}
func help() {
	fmt.Println("_______________")
	fmt.Println("/rooms_list - list of existing rooms")
	fmt.Println("/participants_list - list of all users in the current room")
	fmt.Println("/disconnect - exit chat")
	fmt.Println("/enter_room <name> - enter room with name <name> and create it if necessary")
	fmt.Println("/leave_room - leave current room")
	fmt.Println("_______________")
}
func main() {
	fmt.Println("enter addres and port of server")
	reader := bufio.NewReader(os.Stdin)
	server, _ := reader.ReadString('\n')
	server = server[:len(server)-1]
	conn, _ := net.Dial("tcp", server)
	enc := gob.NewEncoder(conn)
	dec := gob.NewDecoder(conn)
	defer conn.Close()
	var resp Resp
	//var com Command
	for {
		fmt.Println("Enter your name")
		name, _ := reader.ReadString('\n')
		name = name[:len(name)-1]
		if len(name) == 0 {
			continue
		}
		enc.Encode(Command{Name: "connect", Sarg: name})
		dec.Decode(&resp)
		if resp.Result != "Succes" {
			fmt.Println(resp.Result)
		} else {
			break
		}
	}
	fmt.Println("Print /help to get list of coomands")

	go listen(dec)
	for {
		text, _ := reader.ReadString('\n')
		text = text[:len(text)-1]

		if text == "/rooms_list" {
			enc.Encode(Command{Name: "room_list"})
		} else if text == "/help" {
			help()
		} else if text == "/participants_list" {
			enc.Encode(Command{Name: "participants_list"})
		} else if text == "/diconnect" {
			enc.Encode(Command{Name: "disconnect"})
			return
		} else if text == "/leave_room" {
			enc.Encode(Command{Name: "disconnect_room"})
		} else if len(text) > len("/enter_room ") && text[:len("/enter_room ")] == "/enter_room " {
			room := text[len("/enter_room "):]
			enc.Encode(Command{Name: "connect_room", Sarg: room})
			//dec.Decode(&resp)
			//fmt.Println(resp.Name)
			//fmt.Println(resp.Result)
		} else {
			enc.Encode(Command{Name: "text", Sarg: text})
		}

	}

}

//func chk(err error) {
//if err != nil {
//		panic(err)
//	}
//}
