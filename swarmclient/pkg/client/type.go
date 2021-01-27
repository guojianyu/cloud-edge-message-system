/*
@Time    : 2020/12/25 16:31
@Author  : 郭建宇
@Email   : 276381225@qq.com
@File    : type.go
*/
package main

import (
	"k8s.io/klog"
	"github.com/nats-io/stan.go"
	"github.com/nats-io/nats.go"
	"fmt"
	"strings"
	//"os"
	"context"
)

type SwarmClient struct {
	configPath string
	natsServer string
	clusterID string
	clientID string
	tlsPath      tls
	config   *Config
	NatsCon  *nats.Conn
	ConStatus bool
	StreamingCon stan.Conn
	client   interface{}
}

type tls struct {
	rootCACertFile string
	clientCertFile string
	clientKeyFile string
}

type Config struct {
     RootCa string `json:"ca.crt,omitempty"`
     ClientCert string `json:"tls.crt,omitempty"`
     ClientKey string `json:"tls.key,omitempty"`
     Tenant string `json:"tenant,omitempty"`
     MecId string `json:"mec,omitempty"`
     CreationTimestamp string `json:"creationTimestamp,omitempty"`
     Duration string `json:"duration,omitempty"`
     NotAfter string `json:"notAfter,omitempty"`
}


type SwarmClientOptions func(*SwarmClient)
func ConfigPathOption(configPath string) SwarmClientOptions{
	return func(o *SwarmClient){
		o.configPath = configPath
	}
}
func NatsServerOption(natsServer string) SwarmClientOptions{
	return func(o *SwarmClient){
		o.natsServer = natsServer
	}
}
func ClusterIDOption(clusterID string) SwarmClientOptions{
	return func(o *SwarmClient){
		o.clusterID = clusterID
	}
}



func (swarmclient *SwarmClient)GenerateSubject(subject string)(string,error){

	hostname := swarmclient.config.MecId
	tenant := swarmclient.config.Tenant
	sub := ""
	if  strings.HasPrefix(subject,LocalSubjectHeader){
		channel := strings.Replace(subject,LocalSubjectHeader,"",-1)
		sub = strings.Join([]string{hostname,tenant,channel},".")
	}else if strings.HasPrefix(subject,CloudSubjectHeader) {
		channel := strings.Replace(subject,CloudSubjectHeader,"",-1)
		sub = strings.Join([]string{strings.Replace(CloudSubjectHeader,".","",-1),tenant,channel},".")
	}else{
		return "",fmt.Errorf("subject name is error!")
	}
	sub = strings.Replace(sub,"\n","",-1)
	sub = strings.Replace(sub," ","",-1)
	return sub,nil
}

func NewSwarmClient(clientID string,configpath string)(error){

	swarmclient = &SwarmClient{
		configPath:configpath,
		natsServer:ServerDNS,
		clusterID:ClusterID,
		clientID: clientID,
		ConStatus: true,
		tlsPath:tls{
			rootCACertFile:RootcaCertFilePath,
			clientCertFile:ClientCertFilePath,
			clientKeyFile:ClientKeyFilePath,
		},

	}

	if len(swarmclient.clientID) == 0 {
		return fmt.Errorf("clientID do not empty!")
	}
	config,err := ParseConfigFile(swarmclient.configPath)
	if err != nil{
		klog.Error(err)
		return err
	}
	swarmclient.clientID = config.Tenant + "-" + swarmclient.clientID
	swarmclient.config = config
	//write certificates to file
	if exit,_ :=PathExists(swarmclient.tlsPath.rootCACertFile);!exit{
		split := strings.Split(swarmclient.tlsPath.rootCACertFile,"/")
		_,err = ExeSysCommand("mkdir -p "+strings.Join(split[:len(split)-1],"/"))
		if err != nil {
			klog.Error("create cert documents error %s",err)
		}
		tmp,err := Base64decodeString(config.RootCa)
		if err != nil{
			return err
		}
		err = WriteContentToFile(swarmclient.tlsPath.rootCACertFile,tmp)
		if err != nil{
			klog.Error(err)
			return err
		}
	}

	if exit,_ :=PathExists(swarmclient.tlsPath.clientCertFile);!exit{
		split := strings.Split(swarmclient.tlsPath.clientCertFile,"/")
		_,err = ExeSysCommand("mkdir -p "+strings.Join(split[:len(split)-1],"/"))
		if err != nil {
			klog.Error("create cert documents error %s",err)
		}
		tmp,err := Base64decodeString(config.ClientCert)
		if err != nil{
			return err
		}
		err = WriteContentToFile(swarmclient.tlsPath.clientCertFile,tmp)
		if err != nil{
			klog.Error(err)
			return err
		}
	}

	if exit,_ :=PathExists(swarmclient.tlsPath.clientKeyFile);!exit{
		split := strings.Split(swarmclient.tlsPath.clientKeyFile,"/")
		_,err = ExeSysCommand("mkdir -p "+strings.Join(split[:len(split)-1],"/"))
		if err != nil {
			klog.Error("create cert documents error %s",err)
		}
		tmp,err := Base64decodeString(config.ClientKey)
		if err != nil{
			return err
		}
		err = WriteContentToFile(swarmclient.tlsPath.clientKeyFile,tmp)
		if err != nil{
			klog.Error(err)
			return err
		}
	}
	//verify certificate
	err = swarmclient.verify()
	if err != nil{
		klog.Error("Failed to verify certificate %s",err)
		return err
	}

	err = swarmclient.Connect()
	if err != nil{
		klog.Error("Failed to connect server %s",err)
		return err
	}

	//checkout connection status
	go func() {
		swarmclient.CheckoutConnect()
	}()
	return nil
}

func Close(){
	sc := swarmclient.StreamingCon
	sc.Close()
}

func Unsubscribe(cancel context.CancelFunc){
	cancel()
}