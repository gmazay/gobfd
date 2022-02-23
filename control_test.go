package gobfd

import (
	"fmt"
	l "log"
	"os"
	"syscall"
	"testing"
	"time"
)

func CallBackBFDState(ip string, pre, cur int) error {
	fmt.Println(ip, pre, cur)
	return nil
}

func TestNewControl(t *testing.T) {
	family := syscall.AF_UNSPEC
	port := 3784

	var local, remote string
	if len(os.Args) > 2 {
		args := os.Args[1:]
		local = args[0]
		remote = args[1]
	} else {
		local = "192.168.43.103"
		remote = "192.168.43.244"
	}
	passive := false
	rxInterval := 400 // milliseconds
	txInterval := 400 // milliseconds
	detectMult := 1
	var logger *l.Logger

	control := NewControl(local, family, port, logger)
	control.Run()

	l.Printf("add bfd check session  remote ip: ", remote)
	control.AddSession(remote, passive, rxInterval, txInterval, detectMult, CallBackBFDState)

	l.Printf("local bfd server running at port: ", controlPort)
	time.Sleep(time.Second * 30)
}
