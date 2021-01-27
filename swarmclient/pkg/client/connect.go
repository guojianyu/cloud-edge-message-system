/*
@Time    : 2020/12/25 16:10
@Author  : 郭建宇
@Email   : 276381225@qq.com
@File    : connect.go
*/
package main

import (
	 "github.com/nats-io/stan.go"
	"github.com/nats-io/nats.go"
	"time"
	"k8s.io/klog"
	"fmt"
)
func (swarmclient *SwarmClient)Connect()error{
	 //fmt.Println(swarmclient.tlsPath.rootCACertFile)
	 //fmt.Println(swarmclient.tlsPath.clientCertFile)
	 //fmt.Println(swarmclient.tlsPath.clientKeyFile)
	 //fmt.Println(swarmclient.natsServer)
	 rootCA := nats.RootCAs(swarmclient.tlsPath.rootCACertFile)
         clientCert := nats.ClientCert(swarmclient.tlsPath.clientCertFile, swarmclient.tlsPath.clientKeyFile)
         alwaysReconnect := nats.MaxReconnects(-1)
	 for {
		 nc, err := nats.Connect(swarmclient.natsServer, rootCA, clientCert, alwaysReconnect)
		 if err != nil {
		    klog.Error("Error while connecting to NATS, backing off for a sec... (error: %s)", err)
		    time.Sleep(1 * time.Second)
		    continue
		 }
		 swarmclient.NatsCon = nc
		 break
    	}
	sc, err := stan.Connect(swarmclient.clusterID, swarmclient.clientID, stan.NatsConn(swarmclient.NatsCon),stan.Pings(10, 5),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			fmt.Println("Connection lost, reason: ", reason)
			swarmclient.ConStatus = false
		}))
	if err != nil {
            klog.Error("Can't connect: %v.\nMake sure a NATS Streaming Server is running at: %s", err, swarmclient.natsServer)
	    return fmt.Errorf("Can't connect: %v.\nMake sure a NATS Streaming Server is running at: %s", err, swarmclient.natsServer)
	}
	swarmclient.StreamingCon = sc
	return nil
}

func (swarmclient *SwarmClient) CheckoutConnect(){
	for ; ; {

		for ; ; {

			if !swarmclient.ConStatus{
				break

			} else {
				time.Sleep(10 * time.Second)
			}

		}
		swarmclient.StreamingCon.Close()
		err := swarmclient.Connect()
		if err != nil{
			klog.Error("Failed to connect server %s",err)
			time.Sleep(10*time.Second)
		}else {
			swarmclient.ConStatus = true
		}

	}

}
