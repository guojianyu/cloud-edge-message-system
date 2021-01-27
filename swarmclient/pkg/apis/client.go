/*
@Time    : 2021/1/19 16:21
@Author  : 郭建宇
@Email   : 276381225@qq.com
@File    : client.go
*/
package apis

import (
	"plugin"
	"context"
	"strconv"
)
type Swarmclient struct {
	plugin *plugin.Plugin
	clientID string
	ConfigFilePath string
	SwarmclientApisPath string
}
type SwarmClientOptions func(*Swarmclient)
//rewrite configpath
func ConfigPathOption(configPath string) SwarmClientOptions{
	return func(o *Swarmclient){
		o.ConfigFilePath = configPath
	}
}
//rewrite clientapispath
func ClientApisPathOption(SwarmclientApisPath string) SwarmClientOptions{
	return func(o *Swarmclient){
		o.SwarmclientApisPath = SwarmclientApisPath
	}
}

var(
	durableOption = "durable"
	startWithLastReceived = "startWithLastReceived"
	deliverAllAvailable = "deliverAllAvailable"
	startAtSequence = "startAtSequence"
)

type Options struct {
	OptionName  string
	DurableName string
	StartSequence uint64
}

type Option func(*Options)

func DurableName(durabelname string) Option {
	return func(o *Options) {
		o.OptionName = durableOption
		o.DurableName = durabelname
	}
}
func StartWithLastReceived() Option {
	return func(o *Options) {
		o.OptionName = startWithLastReceived
	}
}

func DeliverAllAvailable() Option {
	return func(o *Options) {
		o.OptionName = deliverAllAvailable
	}
}

func StartAtSequence(seq uint64) Option {
	return func(o *Options) {
		o.OptionName = startAtSequence
		o.StartSequence = seq
	}
}


func (sc *Swarmclient)LocalPub(subject string,data []byte)error{
	function,_:=sc.plugin.Lookup(LocalPubFunc)
	err := function.(func(string,[]byte)(error))(subject,data)
	if err != nil{
		return err
	}
	return nil
}

func (sc *Swarmclient)CloudPub(subject string,data []byte)error{
	function,_:=sc.plugin.Lookup(CloudPubFunc)
	err := function.(func(string,[]byte)(error))(subject,data)
	if err != nil{
		return err
	}
	return nil
}


func ParseSubOpts(opts ...Option)[]string{
	 opt := Options{}
	 for _, o := range opts {
		o(&opt)
	 }
	 result := []string{}
	 if len(opts) ==0{
		 return result
	 }
	 if opt.OptionName == startWithLastReceived{
		 result = []string{startWithLastReceived}
	 }else if opt.OptionName == deliverAllAvailable{
		 result = []string{deliverAllAvailable}
	 }else if opt.OptionName == startAtSequence{
		 seq:=strconv.Itoa(int(opt.StartSequence))
		 result = []string{startAtSequence,seq}
	 }else if opt.OptionName == durableOption{
		 result = []string{durableOption,opt.DurableName}
	 }
	 return result

}
func (sc *Swarmclient)LocalSub(ctx context.Context,subject string,cb func([]byte),opts ...Option)error{
	function,_:=sc.plugin.Lookup(LocalSubFunc)
	err := function.(func(context.Context,string,func([]byte),...string)(error))(ctx,subject,cb,ParseSubOpts(opts...)...)
	if err != nil{
		return err
	}
	return nil
}

func (sc *Swarmclient)CloudSub(ctx context.Context,subject string,cb func([]byte),opts ...Option)error{
	function,_:=sc.plugin.Lookup(CloudSubFunc)
	err := function.(func(context.Context,string,func([]byte),...string)(error))(ctx,subject,cb,ParseSubOpts(opts...)...)
	if err != nil{
		return err
	}
	return nil
}




func NewSwarmClient(clientid string,opts ...SwarmClientOptions)(*Swarmclient,error){
	swarmclient := &Swarmclient{ConfigFilePath:ConfigFilePath,SwarmclientApisPath:SwarmclientApisPath}
	for _, o := range opts {
    		o(swarmclient)
  	}
	p, err := plugin.Open(swarmclient.SwarmclientApisPath)
	if err != nil{
		return swarmclient,err
	}
	function, _ := p.Lookup(NewSwarmClientFunc)
	err = function.(func(string,string)(error))(clientid,swarmclient.ConfigFilePath)
	if err != nil{
		return swarmclient,err
	}
	swarmclient.plugin = p
	swarmclient.clientID = clientid
	return swarmclient,nil
}

func (sc *Swarmclient)Close(){
	function,_:=sc.plugin.Lookup(CloseFunc)
	function.(func())()
}