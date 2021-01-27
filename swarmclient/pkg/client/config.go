/*
@Time    : 2020/12/25 16:09
@Author  : 郭建宇
@Email   : 276381225@qq.com
@File    : config.go
*/
package main

import (
	"os"
	"io/ioutil"
	"fmt"
	"encoding/json"
	"os/exec"
	"encoding/base64"
	"k8s.io/klog"
)


const (
	ConfigFilePath = "/etc/swarm/config"
	RootcaCertFilePath = "/etc/swarm_crt/ca.crt"
	ClientCertFilePath  = "/etc/swarm_crt/tls.crt"
	ClientKeyFilePath  = "/etc/swarm_crt/tls.key"
	ClusterID = "test-cluster"
	ServerDNS = "nats-leaf.nats.svc"
	LocalSubjectHeader = "local."
	CloudSubjectHeader = "cloud."
)

var swarmclient *SwarmClient
func ParseConfigFile(configPath string)(*Config,error){
	  config := &Config{}
	  file, err := os.Open(configPath) // For read access.
	  if err != nil {
	     return config,fmt.Errorf("error: %v", err)
	  }
	  defer file.Close()
	  data, err := ioutil.ReadAll(file)
	  if err != nil {
	    return config,fmt.Errorf("error: %v", err)
	  }
	  err = json.Unmarshal([]byte(string(data)), config)
	  if err != nil {
	    return config,fmt.Errorf("error: %v", err)
	  }
	return config,nil
}

func WriteContentToFile(filepath,content string)error{
	file, err := os.OpenFile(filepath, os.O_RDWR |os.O_TRUNC| os.O_APPEND | os. O_CREATE, 755)
	if err != nil {
		  return err
	}
	defer file.Close()
	 _,err = file.Write([]byte(content))
	if err != nil{
		return err
	}
	 return nil
}

func ExeSysCommand(cmdStr string) (string,error) {
    cmd := exec.Command("sh", "-c", cmdStr)
    opBytes, err := cmd.Output()
    if err != nil {
        return "",err
    }
    return string(opBytes),nil
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func  Base64decodeString(encodeString string)(string,error){
	decodeBytes, err := base64.StdEncoding.DecodeString(encodeString)
	if err != nil {
		klog.Error(err)
		return "",err
	}
	return string(decodeBytes),nil

}

func GenerateLocalSubjectName(subject string)string{
	return LocalSubjectHeader+subject
}

func GenerateCloudSubjectName(subject string)string{
	return CloudSubjectHeader+subject
}