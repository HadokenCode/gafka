package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/funkygao/gafka/cmd/gk/command/protos"
	"github.com/funkygao/gocli"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type Sniff struct {
	Ui  cli.Ui
	Cmd string
}

func (this *Sniff) Run(args []string) (exitCode int) {
	var (
		device     string
		filter     string
		protocol   string
		serverPort int
	)
	cmdFlags := flag.NewFlagSet("sniff", flag.ContinueOnError)
	cmdFlags.Usage = func() { this.Ui.Output(this.Help()) }
	cmdFlags.StringVar(&device, "i", "", "")
	cmdFlags.StringVar(&filter, "f", "", "")
	cmdFlags.StringVar(&protocol, "p", "ascii", "")
	cmdFlags.IntVar(&serverPort, "sp", 0, "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	if validateArgs(this, this.Ui).
		require("-i", "-f").
		invalid(args) {
		return 2
	}

	prot := protos.New(protocol, serverPort)
	if prot == nil {
		this.Ui.Error("unkown protocol")
		this.Ui.Outputf(this.Help())
		return 2
	}

	this.Ui.Outputf("starting sniff on interface %s", device)
	snaplen := int32(1 << 20) // max number of bytes to read per packet
	handle, err := pcap.OpenLive(device, snaplen, true, pcap.BlockForever)
	swallow(err)
	defer handle.Close()

	swallow(handle.SetBPFFilter(filter))

	// Use the handle as a packet source to process all packets
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	this.Ui.Output("starting to read packets...")
	for packet := range packetSource.Packets() {
		this.handlePacket(packet, prot)
	}

	return
}

func (this *Sniff) handlePacket(packet gopacket.Packet, prot protos.Protocol) {
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer == nil {
		return
	}
	ip, _ := ipLayer.(*layers.IPv4)

	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpLayer == nil {
		return
	}
	tcp, _ := tcpLayer.(*layers.TCP)

	applicationLayer := packet.ApplicationLayer()
	if applicationLayer == nil {
		return
	}

	output := prot.Unmarshal(applicationLayer.Payload())
	if len(output) == 0 {
		return
	}

	this.Ui.Info(fmt.Sprintf("%s:%s -> %s:%s %dB",
		ip.SrcIP, tcp.SrcPort, ip.DstIP, tcp.DstPort, len(applicationLayer.Payload())))
	this.Ui.Output(output)
}

func (this *Sniff) Synopsis() string {
	return fmt.Sprintf("Sniff traffic on a network with libpcap")
}

func (this *Sniff) Help() string {
	help := fmt.Sprintf(`
Usage: %s sniff [options]

    %s

Options:

    -i interface

    -f filter
      e,g. tcp and port 80

    -p ascii|zk|kafka
      Default protocol: ascii

    -sp port
      Server port

`, this.Cmd, this.Synopsis())
	return strings.TrimSpace(help)
}
