/*
@Time    : 2020/12/29 11:10
@Author  : 郭建宇
@Email   : 276381225@qq.com
@File    : verify.go
*/
package main

import (
	   certtls "crypto/tls"
	    "crypto/x509"
	    "encoding/pem"
	    "fmt"
            "k8s.io/klog"
	    "io/ioutil"
)

func (swarmclient *SwarmClient)verify()(error){
	//parse certificate
	cert,err := parseCert(swarmclient.tlsPath.clientCertFile,swarmclient.tlsPath.clientKeyFile)
	if err != nil{
		return err
	}
	// verify certificate and config tenant
        //fmt.Println("commonname:",cert.Subject.CommonName)
        //fmt.Println("tenant:",swarmclient.config.Tenant)

	if cert.Subject.CommonName != swarmclient.config.Tenant{
		return fmt.Errorf("certificate do not match tenant")
	}
	return nil
}
func parseCert(crt, privateKey string) (*x509.Certificate,error){
    var cert certtls.Certificate

    certPEMBlock, err := ioutil.ReadFile(crt)
    if err != nil {
        return nil,err
    }

    certDERBlock, restPEMBlock := pem.Decode(certPEMBlock)
    if certDERBlock == nil {
        return nil,fmt.Errorf("parse certificate failed")
    }

    cert.Certificate = append(cert.Certificate, certDERBlock.Bytes)

    certDERBlockChain, _ := pem.Decode(restPEMBlock)
    if certDERBlockChain != nil {
        cert.Certificate = append(cert.Certificate, certDERBlockChain.Bytes)
    }

    keyPEMBlock, err := ioutil.ReadFile(privateKey)
    if err != nil {
        return nil,err
    }

    keyDERBlock, _ := pem.Decode(keyPEMBlock)
    if keyDERBlock == nil {
        return nil,fmt.Errorf("parse certificate failed")
    }

    var key interface{}
    var errParsePK error
    if keyDERBlock.Type == "RSA PRIVATE KEY" {
        //RSA PKCS1
        key, errParsePK = x509.ParsePKCS1PrivateKey(keyDERBlock.Bytes)
    } else if keyDERBlock.Type == "PRIVATE KEY" {
        key, errParsePK = x509.ParsePKCS8PrivateKey(keyDERBlock.Bytes)
    }

    if errParsePK != nil {
        return nil,errParsePK
    } else {
        cert.PrivateKey = key
    }

    x509Cert, err := x509.ParseCertificate(certDERBlock.Bytes)
    if err != nil {
        return nil,fmt.Errorf("x509 parse certificate failed ")
    } else {
        switch x509Cert.PublicKeyAlgorithm {
        case x509.RSA:
            {
                klog.V(4).Info("Plublic Key Algorithm:RSA")
            }
        case x509.DSA:
            {
                klog.V(4).Info("Plublic Key Algorithm:DSA")
            }
        case x509.ECDSA:
            {
                klog.V(4).Info("Plublic Key Algorithm:ECDSA")
            }
        case x509.UnknownPublicKeyAlgorithm:
            {
                klog.V(4).Info("Plublic Key Algorithm:Unknow")
            }
        }
    }
    return x509Cert,nil
}
