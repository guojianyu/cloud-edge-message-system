/*
@Time    : 2020/12/28 10:44
@Author  : 郭建宇
@Email   : 276381225@qq.com
@File    : main.go
*/
package main

import (
	"k8s.io/klog"
	"swarmManager/webhandler"
)

func main() {
	//get kubernetes client
	client,err := webhandler.NewClusterClient()
	if err != nil{
		klog.Error(err)
		return
	}
	klog.Info("start webserver")
	webhandler.InitWebHandler(client)

}