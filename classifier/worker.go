package main

import (
	"chandler.com/gogen/common"
	"chandler.com/gogen/utils"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"gonum.org/v1/gonum/stat"
	"log"
	"math"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type worker struct{
	id                 int
	doneChan           chan common.Signal
	processedFlowIdSet *utils.IntSet
	wg                 *sync.WaitGroup
	flowDescs          chan *flowDesc

	flowIdToPktSize map[int][]float64
	flowIdToPktIdt map[int][]float64
	windowSize int

	intf string
	ether *layers.Ethernet
	vlan *layers.Dot1Q
	vlanId uint16
	ipv4 *layers.IPv4
	tcp *layers.TCP
	udp *layers.UDP
	flowIdToPorts map[int][2]int
	options gopacket.SerializeOptions
	buffer gopacket.SerializeBuffer

	handle *pcap.Handle
	rawData []byte
	payloadPerPacketSize int

	selfId int
	targetId int

}



func (w *worker)Init()  {

	w.options.FixLengths = true
	//w.options.ComputeChecksums=true

	w.buffer=gopacket.NewSerializeBuffer()
	w.rawData=make([]byte,1600)
	w.ether = &layers.Ethernet{
		EthernetType: layers.EthernetTypeDot1Q,
	}

	macStr,err:=utils.GenerateMAC(w.selfId)
	if err!=nil{
		log.Fatalf("invalid generater id %d\n",w.selfId)
	}
	mac,err:=net.ParseMAC(macStr)
	if err!=nil{
		log.Fatal(err)
	}
	w.ether.SrcMAC=mac


	dstmacStr,err:=utils.GenerateMAC(w.targetId)
	if err!=nil{
		log.Fatalf("invalid target id %d\n",w.targetId)
	}
	dstMac,err:=net.ParseMAC(dstmacStr)
	if err!=nil{
		log.Fatalf("invalid dst mac %s\n",dstmacStr)
	}
	w.ether.DstMAC=dstMac

	w.vlan=&layers.Dot1Q{
		VLANIdentifier: w.vlanId,
		Type: layers.EthernetTypeIPv4,
	}

	w.ipv4 = &layers.IPv4{
		Version:    4,   //uint8
		IHL:        5,   //uint8
		TOS:        0,   //uint8
		Id:         0,   //uint16
		Flags:      0,   //IPv4Flag
		FragOffset: 0,   //uint16
		TTL:        255, //uint8
	}

	ipStr,err:=utils.GenerateIP(w.selfId)
	if err!=nil{
		log.Fatalf("invalid generate id %d\n",w.selfId)
	}
	ip:=net.ParseIP(ipStr)
	if nil==ip{
		log.Fatalf("invalid ip %s\n",ipStr)
	}
	w.ipv4.SrcIP=ip
	dstIpStr,err:=utils.GenerateIP(w.targetId)
	if err!=nil{
		log.Fatalf("invalid target id %d\n",w.targetId)
	}
	dstIp:=net.ParseIP(dstIpStr)
	if nil==dstIp{
		log.Fatalf("invalid target dst ip  %s\n",dstIpStr)
	}
	w.ipv4.DstIP=dstIp

	w.tcp = &layers.TCP{}
	w.udp = &layers.UDP{}
	rand.Seed(time.Now().UnixNano())
}

func (g  *worker)sendPkt(payloadSize int,isTCP bool) (err error) {
	count:=payloadSize/g.payloadPerPacketSize

	for ;count>0;count--{
		payLoadPerPacket:=g.rawData[:g.payloadPerPacketSize]

		payloadSize-=g.payloadPerPacketSize
		if isTCP{
			err=gopacket.SerializeLayers(g.buffer,g.options,g.ether,g.vlan,g.ipv4,g.tcp,gopacket.Payload(payLoadPerPacket))
			if err!=nil{
				return err
			}
			err=g.handle.WritePacketData(g.buffer.Bytes())
			if err!=nil{
				return err
			}
		}else{
			err=gopacket.SerializeLayers(g.buffer,g.options,g.ether,g.vlan,g.ipv4,g.udp,gopacket.Payload(payLoadPerPacket))
			if err!=nil{
				return err
			}
			err=g.handle.WritePacketData(g.buffer.Bytes())
			if err!=nil{
				return err
			}
		}
	}
	// send all we return
	if payloadSize==0{
		return nil
	}


	if payloadSize<9{
		payloadSize=9
	}

	leftPayload :=g.rawData[:payloadSize]

	if isTCP{
		err=gopacket.SerializeLayers(g.buffer,g.options,g.ether,g.vlan,g.ipv4,g.tcp,gopacket.Payload(leftPayload))
		if err!=nil{
			return err
		}
		err=g.handle.WritePacketData(g.buffer.Bytes())
		if err!=nil{
			return err
		}
	}else{
		err=gopacket.SerializeLayers(g.buffer,g.options,g.ether,g.vlan,g.ipv4,g.udp,gopacket.Payload(leftPayload))
		if err!=nil{
			return err
		}
		err=g.handle.WritePacketData(g.buffer.Bytes())
		if err!=nil{
			return err
		}
	}

	return nil
}



func getStats(stats []float64)(min,max,mean float64)  {
	min=math.MaxFloat64
	max=math.MinInt64
	sum:=float64(0)
	for _,v:=range stats{
		if v>max{
			max=v
		}
		if v<min{
			min=v
		}
		sum+=v
	}
	return min,max,sum/float64(len(stats))
}



func (w *worker)processStats(fn string,stats map[string][]float64)  {
	pktSizes:=stats["pkt_size"]
	idts:=stats["idt"]

	idts=utils.FilterFloat(idts, func(f float64) bool {
		return f>=0
	})
	if len(idts)==0{
		return
	}
	minIdt, maxIdt, meanIdt := getStats(idts)
	stdvIdt :=stat.StdDev(idts,nil)



	pktSizes=utils.FilterFloat(pktSizes, func(f float64) bool {
		return f>=0
	})
	if len(pktSizes)==0{
		return
	}
	minPktsize, maxPktsize, meanPktsize := getStats(pktSizes)
	stdvPktsize :=stat.StdDev(pktSizes,nil)

	w.flowDescs<-&flowDesc{
		FileName:     fn,
		MaxPktSize:   maxPktsize,
		MinPktSize:   minPktsize,
		MeanPktSize:  meanPktsize,
		DevPktSize:   stdvPktsize,
		MaxInterval:  maxIdt,
		MinInterval:  minIdt,
		MeanInterval: meanIdt,
		DevInterval:  stdvIdt,
		TrueLabel:    defaultLoader.getFlowType(fn),
		Pred:         0,
	}
	//log.Println("send")
}

func (w *worker)start()  {
	log.Printf("workerId:%d start\n",w.id)
	defer w.wg.Done()
	handle,err:=pcap.OpenLive(w.intf,1500,false,0)
	if err!=nil{
		log.Fatalf("cannot open interface:%s\n",w.intf)
	}
	w.handle=handle
	defer w.handle.Close()

	stopped:=false
	//shuffle
	l:=defaultLoader
	for{
	loop:	if stopped{
			return
		}
		//read line
		fn:=l.randomPick()
		log.Printf("random pick %s\n",fn)
		lines,err:=utils.ReadLines(fn)
		if nil!=err{
			log.Printf("invalid pkts file %s\n",fn)
			continue
		}
		w.reset()
		for _,line:=range lines{
			if stopped{
				break
			}
			select {
			case <-w.doneChan:
				log.Printf("Worker %d exit",w.id)
				stopped=true
				return
			default:
				//log.Println(line)
				content:=strings.Split(line," ")
				if len(content)!=6{
					log.Fatalf("invalid pkts file %s\n",fn)
				}
				toSleep, err := strconv.ParseFloat(content[0], 64)
				if toSleep < 0 && int(toSleep) != -1 ||err!=nil {
					log.Fatalf("Invalid sleep time in pkt file %s\n", fn)
				}
				size, err := strconv.Atoi(content[1])
				if err != nil {
					log.Fatalf("Invalid pkt size in pkt file %s\n", fn)
				}

				proto:=content[2]
				tsDiffInFlow, err := strconv.ParseFloat(content[4], 64)
				if tsDiffInFlow < 0 && int(tsDiffInFlow) != -1 {
					log.Fatalln("Invalid ts diff in flow")
				}
				if err != nil {
					log.Fatalf("Invalid ts diff in flow in pkt file %s\n", fn)
				}

				flowId, err := strconv.Atoi(content[3])
				if err != nil {
					log.Fatalf("Invalid flow id in pkt file %s\n", fn)
				}

				//last, err := strconv.ParseInt(content[5], 10, 64)
				//if err != nil {
				//	log.Fatalf("Invalid last payload indicator in pkt file %s\n", fn)
				//}
				//isLastPayload:= last>0

				if !w.processedFlowIdSet.Contains(flowId){
					//store
					if len(w.flowIdToPktSize[flowId])==w.windowSize{
						//process
						w.processedFlowIdSet.Add(flowId)
						flowstats:=make(map[string][]float64)
						flowstats["pkt_size"]=w.flowIdToPktSize[flowId]
						flowstats["idt"]=w.flowIdToPktIdt[flowId]
						//log.Println("process")
						go w.processStats(fn,flowstats)
						goto loop
					}else{
						w.flowIdToPktSize[flowId]=append(w.flowIdToPktSize[flowId],float64(size))
						w.flowIdToPktIdt[flowId]=append(w.flowIdToPktIdt[flowId], tsDiffInFlow)
					}

				}
				//decide ports
				_,ok:=w.flowIdToPorts[flowId]
				if !ok{
					//random generate
					sp:=rand.Intn(1350)
					sp+=10
					dp:=rand.Intn(1350)
					dp+=10
					w.flowIdToPorts[flowId]=[2]int{
						sp,dp,
					}
				}
				ports:=w.flowIdToPorts[flowId]
				isTcp:=proto=="TCP"
				if isTcp{
					w.ipv4.Protocol=6
					w.tcp.SrcPort=layers.TCPPort(ports[0])
					w.tcp.DstPort=layers.TCPPort(ports[1])
				}else{
					w.ipv4.Protocol=17
					w.udp.SrcPort=layers.UDPPort(ports[0])
					w.udp.DstPort=layers.UDPPort(ports[1])
				}
				//send
				err = w.sendPkt(size, proto == "TCP")
				if err!=nil{
					log.Printf("error:%s\n",err)
				}
				time.Sleep(time.Duration(int(toSleep/2))*time.Nanosecond)
			}
		}

	}
}

func (w *worker)reset(){
	w.processedFlowIdSet =utils.NewIntSet()
	w.flowIdToPktIdt=make(map[int][]float64)
	w.flowIdToPktSize=make(map[int][]float64)
	w.windowSize=10
	w.flowIdToPorts=make(map[int][2]int)
}

