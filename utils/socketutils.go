package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
)

func SendByte(ip string,port int,content []byte) (err error)  {
	conn,err:=net.Dial("tcp",fmt.Sprintf("%s:%d",ip,port))
	if err!=nil{
		return err
	}
	defer conn.Close()
	_, err = conn.Write(content)
	return err
}

func SendStr(ip string,port int,content *string)(err error)  {
	return SendByte(ip,port,[]byte(*content))
}

func SendMap(ip string,port int,content map[string]interface{}) error  {
	jsonBytes,err:=json.Marshal(content)
	if err!=nil{
		return err
	}
	err= SendByte(ip,port,jsonBytes)
	return err
}

// send and wait for response
func SendAndRecv(ip string,port int,content []byte,delim byte) (string,error){
	conn,err:=net.Dial("tcp",fmt.Sprintf("%s:%d",ip,port))
	if err!=nil{
		log.Println("error when dial")
		return "",err
	}
	//length:=len(content)
	//there should be a for loop
	_, err = conn.Write(content)
	if err!=nil{
		return "",err
	}
	respone:=bufio.NewReader(conn)
	resp,err:=respone.ReadBytes(delim)
	if err!=nil{
		return "",err
	}
	tmp:=resp[0:len(resp)-1]
	return string(tmp),nil
}
