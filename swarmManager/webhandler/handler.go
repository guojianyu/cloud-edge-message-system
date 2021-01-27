/*
@Time    : 2020/12/23 10:44
@Author  : 郭建宇
@Email   : 276381225@qq.com
@File    : handler.go
*/

package webhandler

import (
	"net/http"
	"github.com/emicklei/go-restful"
	"k8s.io/klog"
	"fmt"
	"encoding/json"
	"encoding/base64"

)

var clusterclient *ClusterClient

func Register(ws *restful.WebService) {
	ws.Route(
		ws.POST("/login").
			To(handlerLogin))
	ws.Route(
		ws.POST("/cert/tenant").
			To(handlerCreateTenantCert))
	ws.Route(
		ws.POST("/mec/register").
			To(handlerRegisterEdge))
	ws.Route(
		ws.GET("/mec").
			To(handlerShowMec))
	ws.Route(
		ws.GET("/cert/tenant/{tenant}/{mecid}").
			To(handlerGetTenantCert))
	ws.Route(
		ws.GET("/cert/edgecert").
			To(handlerGetCloudNatsLeafAndLocalCert))
}

func handlerLogin(request *restful.Request, response *restful.Response) {
        var err error
	if err != nil{
	    klog.Error(err)
	    response.WriteHeaderAndEntity(500,"datas save failed ")
	}else{
	    response.WriteHeaderAndEntity(200,"datas save successed")
	}

}

func handlerCreateTenantCert(request *restful.Request, response *restful.Response) {
	parameter,err := parseRequestParameter(request)
	tenant := parameter["tenant"].(string)
	if err == nil{
		err = clusterclient.CreateTenantCert(tenant)
	}
	if err != nil{
		klog.Error(err)
		response.WriteHeaderAndEntity(500,"create tenant certificate failed")
	}else {
		response.WriteHeaderAndEntity(200,"create tenant certificate")
	}

}

func handlerRegisterEdge(request *restful.Request, response *restful.Response) {
	parameter,err := parseRequestParameter(request)
	mecname := parameter["mec"].(string)
	mecip := parameter["ip"].(string)
	if err == nil{
		//write data
		err = clusterclient.CreateMecClusterSecret(mecname,mecip)
	}
	if err != nil{
		klog.Error(err)
		response.WriteHeaderAndEntity(500,"register edge failed")
	}else {
		response.WriteHeaderAndEntity(200,"register edge sucessed")
	}

}

func handlerShowMec(request *restful.Request, response *restful.Response) {
	secrets,err :=clusterclient.GetAllMecClusterSecrets()
	if err != nil{
		klog.Error(err)
		response.WriteHeaderAndEntity(500,"list mec id failed")
	}else {
		response.WriteHeaderAndEntity(200,secrets)
	}

}


func handlerGetTenantCert(request *restful.Request, response *restful.Response) {
	config:=map[string]string{}
	tenant := request.PathParameter("tenant")
	mecid := request.PathParameter("mecid")
	if len(tenant)==0 || len(mecid) ==0{
		response.WriteHeaderAndEntity(500,"tenant or mecid  cannot be empty")
	}
	cert,err :=clusterclient.FindTenantCert(tenant)
	if err != nil{
		klog.Error(err)
		response.WriteHeaderAndEntity(500,"get tenant certificate failed")
		return
	}
	mecidsecret,err :=clusterclient.GetMecClusterSecret(mecid)
	if err != nil{
		klog.Error(err)
		response.WriteHeaderAndEntity(500,"mecid do not register")
		return
	}
	secret,err := clusterclient.FindTenantCertContent(tenant)
	if err != nil{
		klog.Error(err)
		response.WriteHeaderAndEntity(500,"get tenant certificate failed")
		return
	}else {
		crt := secret.Data
		//fmt.Println("crt:",crt)
		config["ca.crt"] =  base64.StdEncoding.EncodeToString(crt["ca.crt"])
		config["tls.crt"] = base64.StdEncoding.EncodeToString(crt["tls.crt"])
		config["tls.key"] = base64.StdEncoding.EncodeToString(crt["tls.key"])
		config["tenant"] = tenant
		config["mec"] = string(mecidsecret.Data["mec"])
		duration,_:= cert.Spec.Duration.MarshalJSON()
		config["duration"] = string(duration)
		config["creationTimestamp"] = cert.CreationTimestamp.String()
		config["notAfter"] = cert.Status.NotAfter.String()
		response.WriteHeaderAndEntity(200,config)
	}

}


func handlerGetCloudNatsLeafAndLocalCert(request *restful.Request, response *restful.Response) {
        certset,err := clusterclient.GetCloudNatsLeafAndLocalCert()
	if err != nil{
		klog.Error(err)
		response.WriteHeaderAndEntity(500,"from cloud certificates failed")
	}else {

		response.WriteHeaderAndEntity(200,certset)
	}


}

func InitWebHandler(client *ClusterClient) {
	clusterclient = client
	apiV1Ws := new(restful.WebService)
	apiV1Ws.Path("/api/v1").Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	Register(apiV1Ws)
	wsContainer := restful.NewContainer()
	wsContainer.EnableContentEncoding(true)
	wsContainer.Add(apiV1Ws)
	fmt.Println("start web")
	// Run a HTTP server that serves static public files from './public' and handles API calls.
	http.ListenAndServe(":9990",wsContainer)

}

func parseRequestParameter(request *restful.Request)(map[string]interface{}, error) {
	request.Request.ParseForm()
	fmt.Println(request.Request.Body)
	formData := make(map[string]interface{})
   	err := json.NewDecoder(request.Request.Body).Decode(&formData)
	if err != nil{
		return formData,err
	}

	return formData,nil
}