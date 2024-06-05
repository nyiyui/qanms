package main

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/nyiyui/qrystal/goal"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
)

var aMachine string
var bMachine string

func main() {
	flag.StringVar(&aMachine, "a-path", "", "path to starting state")
	flag.StringVar(&bMachine, "b-path", "", "path to goal state")
	flag.Parse()

	aMachineData, err := os.ReadFile(aMachine)
	if err != nil {
		panic(err)
	}
	bMachineData, err := os.ReadFile(bMachine)
	if err != nil {
		panic(err)
	}

	var a goal.Machine
	var b goal.Machine
	err = json.Unmarshal(aMachineData, &a)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(bMachineData, &b)
	if err != nil {
		panic(err)
	}
	md := goal.DiffMachine(&a, &b)
	client, err := wgctrl.New()
	if err != nil {
		panic(err)
	}
	handle, err := netlink.NewHandle()
	if err != nil {
		panic(err)
	}
	err = goal.ApplyMachineDiff(a, b, md, client, handle)
	if err != nil {
		panic(err)
	}
}
