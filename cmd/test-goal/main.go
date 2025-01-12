//go:build linux

package main

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/nyiyui/qrystal/goal"
	"github.com/nyiyui/qrystal/util"
	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"
)

var aMachine string
var bMachine string

func main() {
	util.SetupLog()

	flag.StringVar(&aMachine, "a-path", "", "path to starting state")
	flag.StringVar(&bMachine, "b-path", "", "path to goal state")
	flag.Parse()

	zap.S().Info("parsing machine data…")
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
	zap.S().Info("done parsing machine data.")

	md := goal.DiffMachine(&a, &b)
	zap.S().Info("generated diff.")
	data, err := json.MarshalIndent(md, "  ", "  ")
	if err != nil {
		panic(err)
	}
	zap.S().Infof("machine diff:\n%s\n", data)
	client, err := wgctrl.New()
	if err != nil {
		panic(err)
	}
	handle, err := netlink.NewHandle()
	if err != nil {
		panic(err)
	}

	err = goal.ApplyMachineDiff(a, b, md, client, handle, true)
	if err != nil {
		panic(err)
	}
}
