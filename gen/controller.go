package main

import (
	"chandler.com/gogen/common"
	"chandler.com/gogen/models"
	"chandler.com/gogen/utils"
	"fmt"
	"github.com/google/gopacket"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

type controller struct {
	id          int
	flowCounter uint64

	waiter      *sync.WaitGroup
	numWorkers int
	workers     []*generator
	specifiedPktFn string
	mtu                int
	emptySize          int
	selfID             int64
	destinationIDs     []int
	pktsDir            string
	intf               string
	winSize            int
	controllerIP       string
	controllerPort     int
	sleep              bool
	report             bool
	delay              bool
	delayDuration      int
	debug              bool
	forceTarget        bool
	target             int
	enablePktLossStats bool
	pktLossDir         string
	flowType           int
	storeFlowCounter bool

	counterWriter *models.FCounterWriter

	rip   string
	rport int
	volumeStats bool
	volumeStatsDir string
	lines []string

}

func (c *controller) Init() error {
	c.waiter = &sync.WaitGroup{}
	c.waiter.Add(c.numWorkers)
	c.workers = make([]*generator, 0)

	//random readlines
	var pktFileCount int

	files, err := ioutil.ReadDir(c.pktsDir)
	pktFns := make([]string, 0)
	if err != nil {
		return err
	}
	for _, f := range files {
		if strings.Contains(f.Name(), "pkts") {
			pktFileCount++
			pktFns = append(pktFns, f.Name())
		}
	}
	if pktFileCount == 0 {
		log.Fatalf("there is no pkt file in %s", c.pktsDir)
	}
	var pktFile string
	if len(c.specifiedPktFn)==0{
		pktFile=path.Join(c.pktsDir,pktFns[rand.Intn(pktFileCount)])
	}else{
		pktFile=c.specifiedPktFn
	}
	lines,err:=utils.ReadLines(pktFile)

	if err!=nil{
		log.Fatalf("Error reading pkt file %s\n", pktFile)
	}
	c.lines=lines

	for i := 0; i < c.numWorkers; i++ {
		if len(c.specifiedPktFn)!=0{
			log.Printf("Specified a pkts file %s\n",c.specifiedPktFn)
		}
		c.workers = append(c.workers, &generator{
			ID:                 c.id,
			MTU:                c.mtu,
			EmptySize:          c.emptySize,
			SelfID:             i,
			DestinationIDs:     c.destinationIDs,
			PktsDir:            c.pktsDir,
			Int:                c.intf,
			WinSize:            c.winSize,
			ControllerIP:       c.controllerIP,
			ControllerPort:     c.controllerPort,
			Sleep:              c.sleep,
			Report:             c.report,
			Delay:              c.delay,
			DelayTime:          c.delayDuration,
			Debug:              c.debug,
			FlowType:           0,
			ForceTarget:        c.forceTarget,
			Target:             c.target,
			enablePktLossStats: c.enablePktLossStats,
			pktLossDir:         c.pktLossDir,
			waiter:             c.waiter,
			options:            gopacket.SerializeOptions{},
			fType:              c.flowType,
			specifiedPktFn: c.specifiedPktFn,
			selfLoadPkt: false,
			lines: c.lines,
		})
	}

	for _,g:=range c.workers{
		g.Init()
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	stopTickerChan := make(chan common.Signal, 1)
	go func() {
		sig := <-sigs
		log.Printf("generator received signal %s\n", sig)
		stopTickerChan <- common.StopSignal
		for _,g:=range c.workers{
			g.stopChannel<-struct{}{}
		}
	}()

	if c.storeFlowCounter {
		c.counterWriter = models.NewDefaultCounterWriter(c.rip, c.rport)
		err = c.counterWriter.Init()
		if err != nil {
			log.Fatalf("Error connect to redis instance,flow counter won't work\n")
		}

		//start ticker
		ticker := time.NewTicker(time.Duration(5) * time.Second)
		go func() {

			//c.counterWriter = common.NewDefaultCounterWriter(c.rip, c.rport)
			//err = c.counterWriter.init()
			//if err != nil {
			//	log.Println("Error connect to redis instance,flow counter won't work")
			//	return
			//}
			for {
				select {
				case <-ticker.C:
					//collect flow counter and write to redis
					var res int64
					for _, g := range c.workers {
						res += atomic.LoadInt64(&g.flowCounter)
					}
					err := c.counterWriter.Write(res)
					if err != nil {
						log.Println("Error write flow counter to redis")
					}
					continue
				case <-stopTickerChan:
					ticker.Stop()
					return
				}
			}
		}()
	}


	return nil
}

func (c *controller) Start() error {
	for _, g := range c.workers {
		go g.Start()
	}

	if c.volumeStats{
		go c.printVolumeStats()
	}

	c.waiter.Wait()
	if c.storeFlowCounter{
		err := c.counterWriter.Destroy()
		if err != nil {
			log.Println("Error write flow counter to redis")
		}
	}

	log.Println("All work done, controller exits")
	return nil
}

func (c *controller)printVolumeStats()  {
	if !utils.DirExists(c.volumeStatsDir){
		log.Fatalf("no such directory for stats printing %s",c.volumeStatsDir)
	}
	fn:=path.Join(c.volumeStatsDir,fmt.Sprintf("%d.%d.volume",c.id,utils.NowInMilli()))
	log.Printf("start print volume stats to %s",fn)

	if utils.FileExist(fn){
		log.Printf("file exists %s,now delete\n",fn)
		utils.RMFile(fn)
	}
	//volumes:=make([]int,0)
	current_sum:=0
	current_interval:=0.0

	for {
		for _, l := range c.lines {
			contents := strings.Split(l, " ")
			interval, err := strconv.ParseFloat(contents[0], 64)
			if err != nil {
				log.Fatalf("error parse file %s interval\n", fn)
			}
			size, err := strconv.Atoi(contents[1])
			if err != nil {
				log.Fatalf("error parse file %s size\n", fn)
			}
			current_interval += interval
			current_sum += size
			if current_interval >= 5*1e9 {
				//we need to flush
				f, err := os.OpenFile(fn, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					log.Fatalf("error when open file %s\n", fn)
				}
				defer f.Close()

				//todo this is ridiculous since worker is responsible for traffic generation,
				// not controller
				// so it's worker that is responsible for stats printing
				// this piece of code is just a pile of shit
				target:="random"
				if c.forceTarget{
					target=strconv.Itoa(c.target)
				}
				content:=fmt.Sprintf("src:%d,dst:%s,volume:%d\n",c.id,target,current_sum*c.numWorkers)
				if _, err = f.WriteString(content); err != nil {
					log.Fatalf("error write to file %s\n", fn)
				}
				current_sum = 0
				current_interval = 0
				log.Println("flush once")
				time.Sleep(5 * time.Second)
			}
		}
	}
}
