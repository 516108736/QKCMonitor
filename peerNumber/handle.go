package peerNumber

import (
	"fmt"
	"time"

	"github.com/516108736/QKCMonitor/qkcClient"

	"github.com/516108736/QKCMonitor/common"
)

type PeerNumber struct {
	param      param
	QkcClients map[string]*qkcClient.Client
	lastTs     int64
}

type param struct {
	Interval int64
	MinPeer  uint64
}

func (b *PeerNumber) SetExtraParams(data interface{}) {
	res := param{}
	res.Interval = int64(data.(map[string]interface{})["Interval"].(float64))
	res.MinPeer = uint64(data.(map[string]interface{})["MinPeer"].(float64))
	b.param = res
}

func New(config common.Config) (*PeerNumber, error) {
	b := &PeerNumber{
		QkcClients: make(map[string]*qkcClient.Client),
	}

	for _, host := range config.TargetHosts {
		b.QkcClients[host] = qkcClient.NewClient(host)
	}

	b.SetExtraParams(config.Params)
	return b, nil
}

func (b *PeerNumber) makeError(host string, errMsg error) string {
	return fmt.Sprintf("host:%v \nerrMsg:%v", host, errMsg)
}

func (b *PeerNumber) Check() []string {
	res := make([]string, 0)
	for host, client := range b.QkcClients {
		ts := time.Now().Unix()
		peerNumber, err := client.GetPeers()
		if err != nil {
			res = append(res, b.makeError(host, err))
		}
		if ts-b.lastTs >= b.param.Interval {
			if peerNumber <= int(b.param.MinPeer) {
				res = append(res, b.makeError(host, fmt.Errorf(
					"上次检查点%d  本次检查点%d  时间间隔%d peer数量%d<=%d",
					b.lastTs, ts, ts-b.lastTs, peerNumber, b.param.MinPeer)))
			}
		}
		fmt.Println("PeerNumber check end", "ip", host, time.Now().Format("2006-01-02 15:04:05"), "peer number", peerNumber)
		b.lastTs = ts
	}
	return res
}

func (b *PeerNumber) Summary() []string {
	res := make([]string, 0)
	for host, client := range b.QkcClients {
		peerNumber, err := client.GetPeers()
		if err != nil {
			res = append(res, b.makeError(host, err))
		} else {
			res = append(res, fmt.Sprintf("ip:%v peerNumber:%d", host, peerNumber))
		}
	}
	return res
}

func (b *PeerNumber) PreCheck() error {
	for host, client := range b.QkcClients {
		h, err := client.GetPeers()
		fmt.Println("BlockTime PreCheck", host, h, err)
		if err != nil {
			return fmt.Errorf("host %v peerNumber err %v", host, err.Error())
		}
	}
	fmt.Println("PeerNumber PreCheck end", len(b.QkcClients))
	return nil
}
