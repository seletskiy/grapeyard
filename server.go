package main

import (
	"fmt"
	"time"
	"github.com/hashicorp/memberlist"
)

type MyBroadcast string

func (mb MyBroadcast) Invalidates(b memberlist.Broadcast) bool {
	return false
}

func (mb MyBroadcast) Message() []byte {
	return []byte{'h', 'e', 'l', 'l', 'o'}
}

func (mb MyBroadcast) Finished() {
	fmt.Println("broadcast finished")
}

type MyDelegate struct {

}

func (m MyDelegate) NodeMeta(limit int) []byte {
	fmt.Println("NodeMeta called")
	return nil
}

func (m MyDelegate) NotifyMsg([]byte) {
	fmt.Println("NotifyMsg called")
}

func (m MyDelegate) GetBroadcasts(overhead, limit int) [][]byte {
	fmt.Println("GetBroadcasts called")
	return [][]byte{
		[]byte{
			'h', 'e', 'l', 'l', 'o',
		},
	}
}

func (m MyDelegate) LocalState(join bool) []byte {
	fmt.Println("LocalState called")
	return nil
}

func (m MyDelegate) MergeRemoteState(buf []byte, join bool) {
	fmt.Println("MergeRemoteState called")
}

func main() {
	/* Create the initial memberlist from a safe configuration.
	Please reference the godoc for other default config types.
	http://godoc.org/github.com/hashicorp/memberlist#Config
	*/
	conf := memberlist.DefaultLocalConfig()
	conf.Delegate = &MyDelegate{}

	list, err := memberlist.Create(conf)
	if err != nil {
		panic("Failed to create memberlist: " + err.Error())
	}

	// Join an existing cluster by specifying at least one known member.
	n, err := list.Join([]string{"127.1:7947"})
	if err != nil {
		panic("Failed to join cluster: " + err.Error())
	} else {
		fmt.Println(n)
	}

	// Ask for members of the cluster
	for _, member := range list.Members() {
		fmt.Printf("Member: %s %s\n", member.Name, member.Addr)
	}

	time.Sleep(10 * time.Second)

	//memberlist.LocalNode()

	// Continue doing whatever you need, memberlist will maintain membership
	// information in the background. Delegates can be used for receiving
	// events when members join or leave.
}
