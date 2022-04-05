package main

import (
	"os"

	"github.com/orochi-network/orochimaru/keypair"
	"github.com/orochi-network/orochimaru/network"
)

func main() {
	keyfile := AppConfig.GetKeyFile()
	var nodeKey *keypair.KeyPair
	if _, err := os.Stat(keyfile); err != nil {
		// Create a new key pair
		nodeKey, err = keypair.New()
		if err != nil {
			log.Panic(err)
		}
		log.Debugf("save key to file: %s", keyfile)
		nodeKey.SaveToFile(keyfile)
	} else {
		// Load key from json file if existed
		nodeKey, err = keypair.LoadFromFile(keyfile)
		log.Debugf("load key from file: %s", keyfile)
		if err != nil {
			log.Panic(err)
		}
	}

	net := network.New(AppConfig.GetBindHost(), AppConfig.GetBindPort(), AppConfig.GetDomain(), nodeKey)
	net.Announce()
	net.Join()

}
