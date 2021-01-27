/*
@Time    : 2020/12/28 13:55
@Author  : 郭建宇
@Email   : 276381225@qq.com
@File    : streaming.go
*/
package main

import (

	"k8s.io/klog"
	"github.com/nats-io/stan.go"
	"context"
	"strconv"
)
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

func getStreamingMethod(opts ...string)stan.SubscriptionOption{
	 var method  stan.SubscriptionOption
	 if len(opts) ==0{
		 return setDefaultStreamingMethod()
	 }

	 if opts[0] == startWithLastReceived{
		 method = stan.StartWithLastReceived()
	 }else if opts[0] == deliverAllAvailable{
		 method = stan.DeliverAllAvailable()
	 }else if opts[0] == startAtSequence{
		 seq,_:=strconv.Atoi(opts[1])
		 method = stan.StartAtSequence(uint64(seq))
	 }else if opts[0] == durableOption{
		 method = stan.DurableName(opts[1])
	 }

	return method
}
func setDefaultStreamingMethod()stan.SubscriptionOption{
	return stan.DurableName("my-durable")
}

func LocalPub(subject string,data []byte)error{
	 sub,err:= swarmclient.GenerateSubject(GenerateLocalSubjectName(subject))
	 if err != nil{
		 return err
	 }
	 klog.V(4).Info("subject: ",sub)
	 err = swarmclient.StreamingCon.Publish(sub,data) // does not return until an ack has been received from NATS Streaming
         if err != nil {
		klog.Error("Error during publish: %v\n", err)
	        return err
         }
	return nil
}

func CloudPub(subject string,data []byte)error{
	 sub,err:= swarmclient.GenerateSubject(GenerateCloudSubjectName(subject))
	 if err != nil{
		 return err
	 }
	 klog.V(4).Info("subject: ",sub)
	 err = swarmclient.StreamingCon.Publish(sub,data) // does not return until an ack has been received from NATS Streaming
         if err != nil {
		klog.Error("Error during publish: %v\n", err)
	        return err
         }
	return nil
}
//https://docs.nats.io/developing-with-nats-streaming/receiving
//stan.StartWithLastReceived() argument more informations
//stan.DurableName("my-durable") record message status for message duration
func LocalSub(ctx context.Context,subject string,cb func([]byte),opts ...string)(error){
	 subname,err:= swarmclient.GenerateSubject(GenerateLocalSubjectName(subject))
	 if err != nil{
		 return err
	 }

	 klog.V(4).Info("subject: ",subname)

	 method := getStreamingMethod(opts...)
	 replica := swarmclient.StreamingCon
	 sub, err := replica.Subscribe(subname,func(m *stan.Msg) {
		 cb(m.Data)

   	 },method) // does not return until an ack has been received from NATS Streaming
         if err != nil {
		klog.Error("Error during Subscribe: %v\n", err)
	        return err
         }
	for  {
		select {
		case <-ctx.Done():
		     klog.V(4).Info("subscribe closed")
		     sub.Close()
		    //sub.Unsubscribe()
		    return nil
		default:
		     if replica != swarmclient.StreamingCon{
			sub.Close()
			replica = swarmclient.StreamingCon
			sub, err = replica.Subscribe(subname,func(m *stan.Msg) {

				cb(m.Data)

			 },method) // does not return until an ack has been received from NATS Streaming
			 if err != nil {
				klog.Error("Error during Subscribe: %v\n", err)
				return err
			 }

		    }
		}
	}

	return nil
}

func CloudSub(ctx context.Context,subject string,cb func([]byte),opts ...string)(error){
	 subname,err:= swarmclient.GenerateSubject(GenerateCloudSubjectName(subject))
	 if err != nil{
		 return err
	 }
	 klog.V(4).Info("subject: ",subname)
	 method := getStreamingMethod(opts...)
	 replica := swarmclient.StreamingCon
	 sub, err := replica.Subscribe(subname,func(m *stan.Msg) {
		 
		 cb(m.Data)

   	 },method) // does not return until an ack has been received from NATS Streaming
         if err != nil {
		klog.Error("Error during Subscribe: %v\n", err)
	        return err
         }
	for {

		select {
		case <-ctx.Done():
		     klog.V(4).Info("subscribe closed")
		    //sub.Unsubscribe()
		    sub.Close()
		    return nil
		default:
		    if replica != swarmclient.StreamingCon{
			sub.Close()
		        replica = swarmclient.StreamingCon
		        sub, err = replica.Subscribe(subname,func(m *stan.Msg) {

		 		cb(m.Data)

			 },method) // does not return until an ack has been received from NATS Streaming
			 if err != nil {
				klog.Error("Error during Subscribe: %v\n", err)
				return err
			 }

		    }
		}
    	}
	return nil
}
