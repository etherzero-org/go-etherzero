package aux

import (
	"net"
	"strings"
	"strconv"
	"fmt"
	"github.com/etherzero/go-etherzero/common"
	"github.com/etherzero/go-etherzero/p2p/enode"
)

func Ipnrtonetip(ipnr uint32) net.IP {
	var bytes [4]byte
	bytes[0] = byte(ipnr & 0xFF)
	bytes[1] = byte((ipnr >> 8) & 0xFF)
	bytes[2] = byte((ipnr >> 16) & 0xFF)
	bytes[3] = byte((ipnr >> 24) & 0xFF)
	return net.IPv4(bytes[3], bytes[2], bytes[1], bytes[0])
}

// mask code
// t.Log((InetTobton(net.IPv4(192,168,1,190))))
func Netiptoipnr(ipnr net.IP) uint32 {
	bits := strings.Split(ipnr.String(), ".")

	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])

	var sum uint32

	sum += uint32(b0) << 24
	sum += uint32(b1) << 16
	sum += uint32(b2) << 8
	sum += uint32(b3)

	return sum
}

// high 32 is ip
// low 32 port
func EncodeIpPort(ip, port uint32) (ret uint64) {

	ipTmp := uint64(ip)

	ret |= (ipTmp << 32)
	ret |= uint64(port)
	return
}

// DecodeIpPort
func DecodeIpPort(decode uint64) (ip, port uint32) {

	ip = uint32(decode >> 32)
	port = uint32(decode & 0xFFFFFFFF)
	return
}

func PrefixZeroString(count uint32) (zeroString string) {
	for {
		if count == 0 {
			break
		}

		count -= 1
		zeroString = fmt.Sprintf("%v%v", "0", zeroString)
	}
	fmt.Printf("zeroNumberzeroNumberzeroNumber %v\n", zeroString)
	return
}

func NewDiscoverNode(id1, id2 [32]byte, ipPort uint64) (node *enode.Node) {

	ip, port := DecodeIpPort(ipPort)

	netip := Ipnrtonetip(ip)

	nodeid := make([]byte, 64)
	nodeid = append(id1[:], id2[:]...)

	nodeidStr := common.Bytes2Hex(nodeid)
	fmt.Printf("nodeidStr is %v\n", nodeidStr)
	enodeStr := fmt.Sprintf("enode://%s@%s:%d", nodeidStr, netip.String(), port)
	fmt.Printf("enodeStr is %v\n", enodeStr)
	node = enode.MustParseV4(enodeStr)
	fmt.Printf("NewDiscoverNode is %v\n", node.String())
	return
}
