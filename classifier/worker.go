package main

import (
	"chandler.com/gogen/common"
	"chandler.com/gogen/utils"
	"gonum.org/v1/gonum/stat"
	"log"
	"math"
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
	pktSizes:=stats["pkts_size"]
	idts:=stats["idt"]

	idts=utils.FilterFloat(idts, func(f float64) bool {
		return f>=0
	})
	if len(idts)==0{
		return
	}
	min_idt,max_idt,mean_idt:= getStats(idts)
	stdv_idt:=stat.StdDev(idts,nil)



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
		MaxInterval:  max_idt,
		MinInterval:  min_idt,
		MeanInterval: mean_idt,
		DevInterval:  stdv_idt,
		TrueLabel:    defaultLoader.getFlowType(fn),
		Pred:         0,
	}
}

func (w *worker)start()  {
	defer w.wg.Done()
	stopped:=false
	//shuffle
	l:=defaultLoader
	for{
		if stopped{
			return
		}
		//read line
		fn:=l.randomPick()
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
				log.Println(line)
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
						//todo
						//process
						w.processedFlowIdSet.Add(flowId)
						flowstats:=make(map[string][]float64)
						flowstats["pkt_size"]=w.flowIdToPktSize[flowId]
						flowstats["idt"]=w.flowIdToPktIdt[flowId]
						go w.processStats(fn,flowstats)
					}else{
						w.flowIdToPktSize[flowId]=append(w.flowIdToPktSize[flowId],float64(size))
						w.flowIdToPktIdt[flowId]=append(w.flowIdToPktIdt[flowId], tsDiffInFlow)
					}

				}
				time.Sleep(100*time.Nanosecond)
			}
		}

	}
}

func (w *worker)reset(){
	w.processedFlowIdSet =utils.NewIntSet()
	w.flowIdToPktIdt=make(map[int][]float64)
	w.flowIdToPktSize=make(map[int][]float64)
	w.windowSize=10
}

