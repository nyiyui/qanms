//go:build linux

package main

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/nyiyui/qrystal/goal"
	"github.com/nyiyui/qrystal/util"
	"go.uber.org/zap"
)

var mPath string

func main() {
	util.SetupLog()

	flag.StringVar(&mPath, "m-path", "", "path to goal state")
	flag.Parse()

	zap.S().Info("parsing machine data…")
	raw, err := os.ReadFile(mPath)
	if err != nil {
		panic(err)
	}

	var m goal.Machine
	err = json.Unmarshal(raw, &m)
	if err != nil {
		panic(err)
	}
	zap.S().Info("done parsing machine data.")

	applier, err := goal.NewApplier(goal.ApplierOptions{})
	if err != nil {
		panic(err)
	}
	err = applier.ApplyMachine(m)
	if err != nil {
		panic(err)
	}
}
