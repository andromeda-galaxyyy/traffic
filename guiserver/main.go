package main

import (
	"chandler.com/gogen/models"
	"chandler.com/gogen/utils"
	"github.com/gin-gonic/gin"
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


func getMaxLinkRate(c *gin.Context){
	var res=0
	if rand.Float64()<0.5{
		res=100+rand.Intn(20)
	}else{
		res=20+rand.Intn(20)
	}
	//c.JSON(http.StatusOK,gin.H{
	//	"res":res,
	//})

	resp:=make([]gin.H,0)
	resp=append(resp,gin.H{"res":res})
	c.JSON(http.StatusOK,resp)
	return
}

func main()  {
	rand.Seed(utils.NowInNano())
	log.Println("this is gui test server")
	r:=gin.Default()
	r.GET("/delay",getDelay)
	r.GET("/loss",getLoss)
	r.GET("/maxrate",getMaxLinkRate)
	server:=&http.Server{
		Addr: ":10086",
		Handler: r,
	}
	log.Println("server start")
	server.ListenAndServe()
	log.Println("server stop")
}
