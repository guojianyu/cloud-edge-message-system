/*
@Time    : 2020/12/28 15:06
@Author  : 郭建宇
@Email   : 276381225@qq.com
@File    : main.go
*/
package main

import (
	"fmt"
	"k8s.io/klog"
	"swarmclient/pkg/apis"
	"flag"
	"context"
	"time"
)

func main() {

	//clientID is the client unique flag,record message status
	//clientID should contain only alphanumeric characters, - or _
	ctx, cancel:= context.WithCancel(context.Background())
	var (
		configPath string
		subject string
		clientID  string

    	)
	flag.StringVar(&configPath, "config", apis.ConfigFilePath, "config file path")
        flag.StringVar(&subject, "subject", "channel", "subject")
        flag.StringVar(&clientID, "client", "client-123", "clientID is the client unique flag")
	flag.Parse()
	swarmclient,err := apis.NewSwarmClient(
		clientID,

		//******optional

		//apis.ConfigPathOption("/etc/swarm/config"),
		//apis.ClientApisPathOption("/etc/swarm/apis.so"),
	)
	if err != nil{
		klog.Error("Failed to initialize swarmclient:",err)
		return
	}

	go  func() {
		//if fourth parameter is not filled,default set DurableName("my-durable")
		err = swarmclient.LocalSub(ctx,subject,func(data []byte) {
			//todo your job
			fmt.Printf("Received a message: %s\n", string(data))
		},
		//******Choose one or no

		//apis.StartWithLastReceived(),
		//apis.DeliverAllAvailable(),
		//apis.StartAtSequence(22),
		//apis.DurableName("my-durable"),
		)
		if err != nil {
			klog.Error("Error during Subscribe: ", err)
		}
	}()

	//make sure that my durable setup is finished in first.
	time.Sleep(1)

	for i := 0;i < 5; i++{

		message := fmt.Sprintf("Hello World %v",i)
		err = swarmclient.LocalPub(subject,[]byte(message))
		if err != nil {
			klog.Error("Error during publish:", err)
		}
	}

	time.Sleep(3*time.Second)
	//stop to sub
	cancel()

     	// Close connection
    	swarmclient.Close()

	fmt.Println("finish")

}