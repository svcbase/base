package base

import (
	"os"
)

type SysNoticeT struct {
	Subject string
	Body    string
}
type JsonpickupT struct {
	Identifier        string
	Scene             string
	Path              string
	Clientlanguage_id string
	User_id           string
}

var NoticeChannel chan SysNoticeT

var ExitChannel chan os.Signal
var HeartbeatChannel chan int
var DistributeChannel chan int
var PumpinChannel chan int
var PumpoutChannel chan int
var ImportChannel chan int
var UpgradeChannel chan int
var ZLBPChannel chan int
var ZCYPUMPChannel chan int
var PickupChannel chan JsonpickupT

// var ImportChannel chan JsonimportT
var LabelingChannel chan int /*starting batch labeling tasks at lowest traffic point.*/
var SafegovcnChannel chan int

func init_channel() {
	NoticeChannel = make(chan SysNoticeT, 32)
	HeartbeatChannel = make(chan int, 2)
	ExitChannel = make(chan os.Signal, 1)
	ImportChannel = make(chan int, 1)
	PickupChannel = make(chan JsonpickupT, 2)
	LabelingChannel = make(chan int, 1)
	SafegovcnChannel = make(chan int, 1)
	DistributeChannel = make(chan int, 1)
	PumpinChannel = make(chan int, 1)
	PumpoutChannel = make(chan int, 1)
	ZLBPChannel = make(chan int, 1)
	ZCYPUMPChannel = make(chan int, 1)
	UpgradeChannel = make(chan int, 1)
}
