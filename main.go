package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/mkorz/b618reboot-go/routerclient"
)

type mandatoryFlags struct {
	RouterURL *string
	Username  *string
	Password  *string
	FlagSet   *flag.FlagSet
}

func newFlagSet(name string) mandatoryFlags {
	mf := mandatoryFlags{}
	mf.FlagSet = flag.NewFlagSet(name, flag.ExitOnError)

	mf.RouterURL = mf.FlagSet.String("url", os.Getenv("ROUTER_URL"), "router url ip or name (http://xxx.xxx.xx.xxx)")
	mf.Username = mf.FlagSet.String("username", os.Getenv("ROUTER_USERNAME"), "username for router account")
	mf.Password = mf.FlagSet.String("password", os.Getenv("ROUTER_PASSWORD"), "password for router account")
	return mf
}

func main() {
	signalStatsCmdFlags := newFlagSet("signal-stats")
	rebootCmdFlags := newFlagSet("reboot")

	if len(os.Args) < 2 || os.Args[1] == "help" {
		fmt.Println("one of the following commands is required: signal-stats, reboot")
		os.Exit(1)
	}

	var client *routerclient.RouterClient
	var err error
	switch os.Args[1] {
	case "signal-stats":
		signalStatsCmdFlags.FlagSet.Parse(os.Args[2:])
		client, err = routerclient.NewRouterClient(*signalStatsCmdFlags.RouterURL, *signalStatsCmdFlags.Username, *signalStatsCmdFlags.Password)
		if err != nil {
			panic(err)
		}
		err = client.Login()
		stats, _ := client.GetSignalStats()
		out, _ := json.Marshal(stats)

		fmt.Println(string(out))

	case "reboot":
		rebootCmdFlags.FlagSet.Parse(os.Args[2:])
		client, err = routerclient.NewRouterClient(*rebootCmdFlags.RouterURL, *rebootCmdFlags.Username, *rebootCmdFlags.Password)
		if err != nil {
			panic(err)
		}
		err = client.Login()
		client.Reboot()

	default:
		fmt.Printf("invalid command: %q\n", os.Args[1])
		os.Exit(1)
	}

}
