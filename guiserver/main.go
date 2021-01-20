package main

import (
	"chandler.com/gogen/models"
	"chandler.com/gogen/utils"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
)

var (
	invalidRequestJSON=gin.H{
		"msg":"invalid request",
	}
	internalErrorJSON=gin.H{
		"msg":"internal error",
	}
	notFoundJSON=gin.H{
		"msg":"not found",
	}

)

var topo [][]bool
func init()  {
	topo=make([][]bool,100)
	for i:=0;i<100;i++{
		topo[i]=make([]bool,100)
	}
	//load topo
	log.Println("load topo in current dir/topo.json")
	fn:="./topo.json"
	f,err:=ioutil.ReadFile(fn)
	if err!=nil{
		log.Fatalln(err)
	}
	log.Println("load done")
	obj:=make(map[string][][][]int)
	err=json.Unmarshal([]byte(f),&obj)
	if err!=nil{
		log.Fatalln(err)
	}
	//log.Println(obj)
	links,ok:=obj["topo"]
	if !ok{
		log.Fatalf("invalid file")
	}

	count:=0
	for i:=0;i<100;i++{
		for j:=0;j<100;j++{
			if links[i][j][0]!=-1{
				count++
				topo[i][j]=true
				topo[j][i]=true
			}
		}
	}
	if count!=566{
		panic(count)
	}

}



const N=10
func getDelay(c *gin.Context)  {
	res:=make([]*models.FlowDesc,0)
	rN:=rand.Float64()
	src,srcExists:=c.GetQuery("src")
	if !srcExists{
		c.JSON(http.StatusBadRequest,invalidRequestJSON)
		return
	}
	srcId,_:=strconv.Atoi(src)
	srcIp,_:=utils.GenerateIP(srcId)
	dst,dstExists:=c.GetQuery("dst")
	if !dstExists {
		c.JSON(http.StatusBadRequest, invalidRequestJSON)
		return
	}

	dstId,_:=strconv.Atoi(dst)
	dstIp,_:=utils.GenerateIP(dstId)
	if rN<0.8{
		for i:=0;i<N;i++{
			res=append(res,&models.FlowDesc{
				SrcIP:           srcIp,
				SrcPort:         0,
				DstIP:           dstIp,
				DstPort:         0,
				Proto:           "TCP",
				TxStartTs:       0,
				TxEndTs:         0,
				RxStartTs:       0,
				RxEndTs:         0,
				FlowType:        0,
				ReceivedPackets: 0,
				Loss:            0,
				PeriodPackets:   0,
				PeriodLoss:      0,
				MinDelay:        int64(rand.Intn(3)),
				MaxDelay:        int64(10+rand.Intn(10)),
				MeanDelay:       float64(5+rand.Intn(4)),
				StdVarDelay:     float64(3+rand.Intn(5)),
			})
		}
	}else{
		for i:=0;i<N;i++{
			res=append(res,&models.FlowDesc{
				SrcIP:           srcIp,
				SrcPort:         0,
				DstIP:           dstIp,
				DstPort:         0,
				Proto:           "TCP",
				TxStartTs:       0,
				TxEndTs:         0,
				RxStartTs:       0,
				RxEndTs:         0,
				FlowType:        0,
				ReceivedPackets: 0,
				Loss:            0,
				PeriodPackets:   0,
				PeriodLoss:      0,
				MinDelay:        int64(rand.Intn(10)),
				MaxDelay:        int64(30+rand.Intn(10)),
				MeanDelay:       float64(15+rand.Intn(4)),
				StdVarDelay:     float64(10+rand.Intn(5)),
			})
		}
	}
	c.JSON(http.StatusOK,gin.H{
		"num":len(res),
		"data":res,
	})
	return
}

func getLoss(c *gin.Context)  {
	res:=make([]*models.FlowDesc,0)
	rN:=rand.Float64()
	src,srcExists:=c.GetQuery("src")
	if !srcExists{
		c.JSON(http.StatusBadRequest,invalidRequestJSON)
		return
	}
	srcId,_:=strconv.Atoi(src)
	srcIp,_:=utils.GenerateIP(srcId)
	dst,dstExists:=c.GetQuery("dst")
	if !dstExists {
		c.JSON(http.StatusBadRequest, invalidRequestJSON)
		return
	}

	dstId,_:=strconv.Atoi(dst)
	dstIp,_:=utils.GenerateIP(dstId)
	if rN<0.8{
		for i:=0;i<N;i++{
			res=append(res,&models.FlowDesc{
				SrcIP:           srcIp,
				SrcPort:         0,
				DstIP:           dstIp,
				DstPort:         0,
				Proto:           "UDP",
				TxStartTs:       0,
				TxEndTs:         0,
				RxStartTs:       0,
				RxEndTs:         0,
				FlowType:        0,
				ReceivedPackets: 100,
				Loss:            0.1*rand.Float64(),
				PeriodPackets:   100,
				PeriodLoss:      0.2*rand.Float64(),
				MinDelay:        0,
				MaxDelay:        0,
				MeanDelay:       0,
				StdVarDelay:     0,
			})
		}
	}else{
		for i:=0;i<N;i++{
			res=append(res,&models.FlowDesc{
				SrcIP:           srcIp,
				SrcPort:         0,
				DstIP:           dstIp,
				DstPort:         0,
				Proto:           "UDP",
				TxStartTs:       0,
				TxEndTs:         0,
				RxStartTs:       0,
				RxEndTs:         0,
				FlowType:        0,
				ReceivedPackets: 0,
				Loss:            0,
				PeriodPackets:   0,
				PeriodLoss:      0.3+0.7*rand.Float64(),
				MinDelay:        0,
				MaxDelay:        0,
				MeanDelay:       0,
				StdVarDelay:     0,
			})
		}
	}
	c.JSON(http.StatusOK,gin.H{
		"num":len(res),
		"data":res,
	})
	return
}

func getTelemetryDelay(c *gin.Context)  {
	res:=make(map[string]float64)
	for i:=0;i<100;i++{
		for j:=0;j<100;j++{
			if i>=j{
				continue
			}
			if topo[i][j]{
				delay:=20*rand.Float64()
				link:=fmt.Sprintf("%d-%d",i,j)
				//revLink:=fmt.Sprintf("%d-%d",j,i)
				res[link]=delay
				//res[revLink]=delay
			}
		}
	}
	c.JSON(http.StatusOK,res)
	return
}

func getTelemetryLoss(c *gin.Context)  {
	res:=make(map[string]float64)
	for i:=0;i<100;i++{
		for j:=0;j<100;j++{
			if i>=j{
				continue
			}
			if topo[i][j]{
				loss :=rand.Float64()
				link:=fmt.Sprintf("%d-%d",i,j)
				//revLink:=fmt.Sprintf("%d-%d",j,i)
				res[link]= loss
				//res[revLink]= loss
			}
		}
	}
	c.JSON(http.StatusOK,res)
	return
}


func getMaxLinkRate(c *gin.Context){
	var res=0
	if rand.Float64()<0.5{
		res=100+rand.Intn(20)
	}else{
		res=20+rand.Intn(20)
	}

	f:=float64(res)/120
	c.JSON(http.StatusOK,gin.H{
		"res":f,
	})

	//resp:=make([]gin.H,0)
	//resp=append(resp,gin.H{"res":res})
	//c.JSON(http.StatusOK,resp)
	return
}

func getTrafficMatrix(c *gin.Context)  {
	res:=make(map[string]float64)
	for i:=0;i<100;i++{
		for j:=0;j<100;j++{
			if i==j{
				continue
			}
			key:=fmt.Sprintf("%d-%d",i,j)
			res[key]=15*rand.Float64()
		}
	}
	c.JSON(http.StatusOK,res)
	return
}

func getAllLinkRate(c *gin.Context)  {
	res:=make(map[string]float64)
	for i:=0;i<100;i++{
		for j:=0;j<100;j++{
			if topo[i][j]{
				k:=fmt.Sprintf("%d-%d",i,j)
				res[k]=rand.Float64()*100
			}
		}
	}
	c.JSON(http.StatusOK,res)
	return
}

func main()  {
	rand.Seed(utils.NowInNano())
	log.Println("this is gui test server")
	r:=gin.Default()
	r.GET("/delay",getDelay)
	r.GET("/loss",getLoss)
	r.GET("/maxrate",getMaxLinkRate)
	r.GET("/telemetry/loss",getTelemetryLoss)
	r.GET("/telemetry/delay",getTelemetryDelay)
	r.GET("/linkrate",getAllLinkRate)

	r.GET("/traffic",getTrafficMatrix)
	server:=&http.Server{
		Addr: ":10086",
		Handler: r,
	}
	log.Println("server start")
	server.ListenAndServe()
	log.Println("server stop")
}
