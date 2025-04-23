package main

import (
	"encoding/hex"
	"fmt"
)

const (
	keyRaw = "978bbefb" // random hex key
	flag   = "24HIUT{C'est lorsqu'on ne cherche pas la clef qu'on la trouve}"
)

var (
	key, _ = hex.DecodeString(keyRaw)
)

func main() {
	out := make([]byte, len(flag))
	for i := range flag {
		out[i] = flag[i] ^ key[i%len(key)]
	}
	fmt.Println(hex.EncodeToString(out))
}
