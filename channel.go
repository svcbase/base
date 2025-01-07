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
var DataexportCommandChannel chan string

var ExitChannel chan os.Signal
var HeartbeatChannel chan int
var DistributeChannel chan int
var PumpinChannel chan int
var PumpoutChannel chan int
var ImportChannel chan int
var UpgradeChannel chan int
var ZLBPChannel chan int
var ZCYPUMPChannel chan int
var APIretryChannel chan int
var BatchemailChannel chan int
var PickupChannel chan JsonpickupT
var TxtImgChannel chan string
var TotalChannel chan int
var BizopenedChannel chan int

// var ImportChannel chan JsonimportT
var LabelingChannel chan int /*starting batch labeling tasks at lowest traffic point.*/
var SafegovcnChannel chan int

func init_channel() {
	NoticeChannel = make(chan SysNoticeT, 32)
	DataexportCommandChannel = make(chan string, 32)
	HeartbeatChannel = make(chan int, 2)
	ExitChannel = make(chan os.Signal, 1)
	ImportChannel = make(chan int, 1)
	PickupChannel = make(chan JsonpickupT, 2)
	TxtImgChannel = make(chan string, 1)
	LabelingChannel = make(chan int, 1)
	SafegovcnChannel = make(chan int, 1)
	DistributeChannel = make(chan int, 1)
	PumpinChannel = make(chan int, 1)
	PumpoutChannel = make(chan int, 1)
	ZLBPChannel = make(chan int, 1)
	ZCYPUMPChannel = make(chan int, 1)
	APIretryChannel = make(chan int, 1)
	UpgradeChannel = make(chan int, 1)
	BatchemailChannel = make(chan int, 1)
	TotalChannel = make(chan int, 1)
	BizopenedChannel = make(chan int, 1)
}
