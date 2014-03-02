package gossip

import (
	"encoding/json"
	"io"
	"net/http"
	"bufio"
	"log"
	"fmt"
	"github.com/hashicorp/memberlist"
)

var _ = fmt.Println

type UpdateExecutor interface {
	Run(binStream io.Reader, repoOffset int64, net *Network)
}

type Config struct {
	RootNodes []string
	LocalPort int
	LocalVersion int64
	Name string
}

type Broadcast struct {
	rate int
	msg []byte
}

type Network struct {
	members *memberlist.Memberlist
	broadcasts []*Broadcast
	executor UpdateExecutor
	version int64
}

func (d *Network) GetVersion() int64 {
	return d.version
}

func (d *Network) NodeMeta(limit int) []byte {
	return nil
}

func (d *Network) NotifyMsg(msg []byte) {
	log.Printf("got broadcast: %s", msg)
	var possibleUpdate UpdateMsg
	if err := json.Unmarshal(msg, &possibleUpdate); err == nil {
		update := possibleUpdate
		if update.Version > d.version {
			//log.Println("replicate broadcast due new version")
			d.AddBroadcast(msg)
			//d.broadcasts = append(d.broadcasts, msg)
			d.version = update.Version
			//d.updating = true

			resp, err := http.Get(update.URI)
			if err != nil {
				log.Fatal(err)
				return
			}

			d.executor.Run(bufio.NewReader(resp.Body), update.RepoOffset, d)
		}
	}
}

func (d *Network) GetBroadcasts(overhead, limit int) [][]byte {
	toSend := make([][]byte, 0)
	keep := make([]*Broadcast, 0)
	for _, b := range d.broadcasts {
		if b.rate <= 0 {
			continue
		}

		limit -= len(b.msg)
		if limit < 0 {
			break
		}

		b.rate -= 1

		toSend = append(toSend, b.msg)
		keep = append(keep, b)
	}

	if len(toSend) == 0 {
		return nil
	}

	d.broadcasts = keep

	log.Printf("sending broadcasts: %s", toSend)

	return toSend
}

func (d *Network) LocalState(join bool) []byte {
	return nil
}

func (d *Network) MergeRemoteState(buf []byte, join bool) {

}

func (d *Network) AddBroadcast(buf []byte) {
	// @TODO Hardcoded retransmit rate!
	d.broadcasts = append(d.broadcasts, &Broadcast{
		rate: 3,
		msg: buf,
	})
}

type UpdateMsg struct {
	Version int64 `json:"v"`
	URI string `json:"uri"`
	RepoOffset int64 `json:off`
}

func NewGossipNetwork(netConf Config, executor UpdateExecutor) *Network {
	conf := memberlist.DefaultLocalConfig()
	network := &Network{
		executor: executor,
		version: netConf.LocalVersion,
	}

	conf.BindPort = netConf.LocalPort
	conf.Name = netConf.Name
	conf.Delegate = network
	
	list, err := memberlist.Create(conf)
	if err != nil {
		panic("Failed to create memberlist: " + err.Error())
	}

	n := 0
	for i := 0; i < 3; i++ {
		n, err = list.Join(netConf.RootNodes)
		if n > 0 {
			break
		}
	}

	if n == 0 {
		panic("Can't connect to any of the root nodes: " + err.Error())
	}

	network.members = list

	return network
}

func (net *Network) SendUpdateMsg(version int64, uri string, repoOffset int64) {
	net.version = version
	bin, _ := json.Marshal(UpdateMsg{version, uri, repoOffset})
	net.AddBroadcast(bin)
}

func (net *Network) GetMembers() []*memberlist.Node {
	return net.members.Members()
}

func (net *Network) Join(host string) (int, error) {
	return net.members.Join([]string{host})
}
