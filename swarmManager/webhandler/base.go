/*
@Time    : 2020/12/23 11:13
@Author  : 郭建宇
@Email   : 276381225@qq.com
@File    : base.go
*/
package webhandler

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/kubernetes"
	"github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"

	//"time"
	"time"
	"k8s.io/klog"
	"strings"
)
var (
	//ConfigurationPath = "/etc/kube/config"
	ConfigurationPath = ""
	ThirdpartyResourcesName = "certificates"
	CertNamespace = "nats"
	EdgeNatsServerDns = "nats-leaf.nats.svc"
	EdgeNatsServerCaName = "edge-nats-ca"
	EdgeNatsServerCaKind = "Issuer"
	EdgeNatsServerTlsCertSecretName = "edge-nats-server-tls"
	EdgeNatsStreamingTlsCertSecretName = "edge-nats-client-tls"
)

var (
	NatsServerLeafTlsCertSecretName = "nats-streaming-leaf-tls"

)


type ClusterClient struct{
	csClient *kubernetes.Clientset
	restClient *rest.RESTClient
}

func generateCertName(tenant string) string{
	return "nats-tenant-"+tenant+"-tls"
}

func (client *ClusterClient)FindTenantCertContent(tenant string)(secret *corev1.Secret,err error){
	certname := generateCertName(tenant)
	secret,err=client.csClient.CoreV1().Secrets(CertNamespace).Get(certname,metav1.GetOptions{})
	return
}
func (client *ClusterClient)FindTenantCert(tenant string)(cert v1alpha2.Certificate,err error){
	cert = v1alpha2.Certificate{}
	certname := generateCertName(tenant)
	if err =client.restClient.Get().Namespace(CertNamespace).Resource(ThirdpartyResourcesName).Name(certname).Do().Into(&cert); err!= nil{
		fmt.Println(err)
		return cert,err
	}
	return cert,nil

}

func (client *ClusterClient)CreateTenantCert(tenant string)error{
	certname := generateCertName(tenant)
	cert := v1alpha2.Certificate{}
	cert.APIVersion = "cert-manager.io/v1alpha2"
	cert.Kind = "Certificate"
	cert.Name = certname
	cert.Spec.DNSNames = []string{EdgeNatsServerDns}
	cert.Spec.SecretName = certname
	//default 1 year
	cert.Spec.Duration = &metav1.Duration{8736*time.Hour}
	//default 10 days
	cert.Spec.RenewBefore = &metav1.Duration{240*time.Hour}
	cert.Spec.IssuerRef.Kind = EdgeNatsServerCaKind
	cert.Spec.IssuerRef.Name = EdgeNatsServerCaName
	cert.Spec.CommonName = tenant
	cert.Spec.Organization = []string{tenant}
	cert.Spec.Usages = []v1alpha2.KeyUsage{"client auth"}
	fmt.Println(cert)
	if err :=client.restClient.Post().Namespace(CertNamespace).Resource(ThirdpartyResourcesName).Name(certname).Body(&cert).Do().Into(&v1alpha2.Certificate{}); err!= nil{
		klog.Error(err)
		return err
	}
	return nil

}

func (client *ClusterClient)CreateMecClusterSecret(mecname,mecip string)error{
	secret := &corev1.Secret{}
	secret.Name = "mecid-"+mecname
	secret.Namespace = CertNamespace
	secret.Data = map[string][]byte{"mec":[]byte(mecname),"ip":[]byte(mecip)}
	_,err:=client.csClient.CoreV1().Secrets(CertNamespace).Create(secret)
	return err

}

func (client *ClusterClient)GetAllMecClusterSecrets()(*corev1.SecretList,error){
	tmpsecrets := &corev1.SecretList{}
	secrets,err:=client.csClient.CoreV1().Secrets(CertNamespace).List(metav1.ListOptions{})
	for _,v := range secrets.Items{
		if  strings.HasPrefix(v.Name,"mecid-"){
			tmpsecrets.Items = append(tmpsecrets.Items,v)
		}
	}
	return tmpsecrets,err

}
func (client *ClusterClient)GetMecClusterSecret(mecid string)(*corev1.Secret,error){
	secret,err:=client.csClient.CoreV1().Secrets(CertNamespace).Get("mecid-"+mecid,metav1.GetOptions{})
	return secret,err

}

func (client *ClusterClient)GetCloudNatsLeafAndLocalCert()(map[string]interface{},error){
	//get cloud nats leaf certificates
	certset := map[string]interface{}{}
	edgesvcsecret,err:=client.csClient.CoreV1().Secrets(CertNamespace).Get(EdgeNatsServerTlsCertSecretName,metav1.GetOptions{})
	if err != nil{
		klog.Error(err)
		return certset,err
	}
	certset["edgeServerCert"] = edgesvcsecret.Data
	//edge nats server certificate
	edgeclientsecret,err := client.csClient.CoreV1().Secrets(CertNamespace).Get(EdgeNatsStreamingTlsCertSecretName,metav1.GetOptions{})
	if err != nil{
		klog.Error(err)
		return certset,err
	}
	certset["edgeClientCert"] = edgeclientsecret.Data
	//nats-streaming certificates
	natsleafsecret,err:= client.csClient.CoreV1().Secrets(CertNamespace).Get(NatsServerLeafTlsCertSecretName,metav1.GetOptions{})
	if err != nil{
		klog.Error(err)
		return certset,err
	}
	certset["natsLeafCert"] = natsleafsecret.Data
	return certset,nil
}
func NewClusterClient()(cluster *ClusterClient,err error){
	client := ClusterClient{}
	kubeconfig, err := clientcmd.BuildConfigFromFlags("", ConfigurationPath)
	if err != nil{
		return &client,err
	}
	csClient := kubernetes.NewForConfigOrDie(kubeconfig)
	client.csClient = csClient
	restclient,err := CreateKubeRestclient(kubeconfig,"cert-manager.io","v1alpha2","/apis")
	if err != nil{
		return &client,err
	}
	client.restClient =  restclient
	return &client,nil


}
func CreateKubeRestclient(kubeConfig *rest.Config,group,version,path string)(res *rest.RESTClient,err error){
	groupversion := schema.GroupVersion{
		Group:   group,
		Version: version,
	}
	kubeConfig.GroupVersion = &groupversion
	kubeConfig.APIPath = path
	kubeConfig.ContentType = runtime.ContentTypeJSON
	scheme1 := runtime.NewScheme()
	kubeConfig.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(scheme1)}
	restclient, err := rest.RESTClientFor(kubeConfig)
	if err != nil {
		fmt.Println(err)
	}
	return restclient,err

}
