package aux

import (
	"testing"
	"net"
	"fmt"
)

func TestDecodeIpPort(t *testing.T) {
	ip := Netiptoipnr(net.ParseIP("127.0.0.1"))

	port := uint32(20012)

	fmt.Println("EncodeIpPort", EncodeIpPort(ip, port))
}
