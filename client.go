package main

import (
	"fmt"
	"time"
	"github.com/hashicorp/memberlist"
)

type MyDelegate struct {

}

func (m MyDelegate) NodeMeta(limit int) []byte {
	fmt.Println("NodeMeta called")
	return nil
}

func (m MyDelegate) NotifyMsg(msg []byte) {
	fmt.Println("NotifyMsg called", msg)
}

func (m MyDelegate) GetBroadcasts(overhead, limit int) [][]byte {
	return nil
}

func (m MyDelegate) LocalState(join bool) []byte {
	fmt.Println("LocalState called")
	return nil
}

func (m MyDelegate) MergeRemoteState(buf []byte, join bool) {
	fmt.Println("MergeRemoteState called")
}


func main() {
	conf := memberlist.DefaultLocalConfig()
	conf.BindPort = 7947
	conf.Name = "client"
	conf.Delegate = &MyDelegate{}
	list, err := memberlist.Create(conf)

	if err != nil {
		panic("Failed to create memberlist: " + err.Error())
	}

	// Ask for members of the cluster
	for _, member := range list.Members() {
		fmt.Printf("Member: %s %s:%d\n", member.Name, member.Addr, member.Port)
	}

	for {
		time.Sleep(1 * time.Second)

		fmt.Println("hello")
	}
}
