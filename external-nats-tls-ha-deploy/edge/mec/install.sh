if [ $# != 2 ]; then
    echo "arguments must have cloudIp and datacenter_id"
    exit 1
fi
natsMasterIP="10.121.115.21"
natsMasterIP=$1
natsMasterLeafPort="32005"
certManagerPort="32009"
natsMasterDNS="nats.nats.svc"
internalNatsDNS="nats-leaf.nats.svc"

echo "MasterIP: "$natsMasterIP
#get localip
netcard=$(route |grep default|awk '{print $8}' |awk -F "/" '{print $1}')
localip=$(ip addr |grep inet |grep -v inet6 |grep $netcard|awk '{print $2}' |awk -F "/" '{print $1}')
echo "localip:$localip"


rootPath="/etc/swarm/"
edgeServerTlsPath="/etc/swarm/edge-nats-server-tls-certs"
edgeClientTlsPath="/etc/swarm/edge-nats-client-tls-certs"
natsLeafTlsPath="/etc/swarm/nats-client-tls-certs"

natsDockerImage="nats:2.1.7-alpine3.11"
streamingDockerImage="nats-streaming:alpine"

###init
#pull docker images
docker pull $natsDockerImage
docker pull $streamingDockerImage

#clean /etc/hosts's dns setting
sed -i "/$natsMasterDNS/d" /etc/hosts
sed -i "/$internalNatsDNS/d" /etc/hosts

#clean docker container
docker stop nats streaming
docker rm nats streaming

###depend installation
#install wget
if rpm -q wget &>/dev/null; then
    echo "wget is already installed."
else
    echo "install wget!"
    yum install wget -y
fi



#install jq
if rpm -q jq &>/dev/null; then
    echo "jq is already installed."
else
    echo "install jq!"
    yum install https://dl.fedoraproject.org/pub/epel/epel-release-latest-7.noarch.rpm
    wget http://dl.fedoraproject.org/pub/epel/epel-release-latest-7.noarch.rpm
    rpm -ivh epel-release-latest-7.noarch.rpm
    yum repolist
    yum -y install jq
fi



#map localip to internalNatsDNS in /etc/hosts
echo "$localip  $internalNatsDNS" >> /etc/hosts

#map cloud nats to natsMasterDNS in /etc/hosts
echo "$natsMasterIP  $natsMasterDNS">> /etc/hosts


#save nats-server and nats-streaming configurations and certificate
mkdir $rootPath

#save cloud nats certifcate
mkdir $natsLeafTlsPath
#save edge nats server certificate
mkdir $edgeServerTlsPath
#save edge nats client certificate
mkdir $edgeClientTlsPath


echo "register mec to cloud"

#register mecid
#mecID=$(curl http://$getMecIdAddress)
#if [ $register = "" ]; then
#echo "get mecid failed"
#exit 5
#fi
mecID=$2
echo "datacenter_id: "$mecID
#Registered edge system
register=$(curl -o /dev/null -s -w %{http_code} -H "Content-Type:application/json" -H "Data_Type:msg" -X POST --data '{"mec": "'$mecID'" ,"ip": "'$localip'"}' http://$natsMasterIP:$certManagerPort/api/v1/mec/register)
#register=$(curl -H "Content-Type:application/json" -H "Data_Type:msg" -X POST --data '{"mec": "'$mecID'" ,"ip": "'$localip'"}' http://$natsMasterIP:$certManagerPort/api/v1/mec/register)
echo "register edge: $register"
if [ $register != "200" ]; then
echo "register edge failed"
#exit 5
fi

###install certificate
echo "prepare download certificate"
#requests required certificates
edgecertset=$(curl http://$natsMasterIP:$certManagerPort/api/v1/cert/edgecert)

#edge server certificate
edgesvccert=$(echo $edgecertset | jq '.edgeServerCert')
ca=$(echo $edgesvccert|jq '."ca.crt"')
echo ${ca//\"}|base64 -d > $edgeServerTlsPath"/ca.crt"
tlscrt=$(echo $edgesvccert|jq '."tls.crt"')
echo ${tlscrt//\"}|base64 -d > $edgeServerTlsPath"/tls.crt"
tlskey=$(echo $edgesvccert|jq '."tls.key"')
echo ${tlskey//\"}|base64 -d > $edgeServerTlsPath"/tls.key"

#edge client certificate
edgeclientcert=$(echo $edgecertset | jq '.edgeClientCert')
ca=$(echo $edgeclientcert|jq '."ca.crt"')
echo ${ca//\"}|base64 -d > $edgeClientTlsPath"/ca.crt"
tlscrt=$(echo $edgeclientcert|jq '."tls.crt"')
echo ${tlscrt//\"}|base64 -d > $edgeClientTlsPath"/tls.crt"
tlskey=$(echo $edgeclientcert|jq '."tls.key"')
echo ${tlskey//\"}|base64 -d > $edgeClientTlsPath"/tls.key"

#nats leaf certificate
natsleafcert=$(echo $edgecertset | jq '.natsLeafCert')
ca=$(echo $natsleafcert|jq '."ca.crt"')
echo ${ca//\"}|base64 -d > $natsLeafTlsPath"/ca.crt"
tlscrt=$(echo $natsleafcert|jq '."tls.crt"')
echo ${tlscrt//\"}|base64 -d > $natsLeafTlsPath"/tls.crt"
tlskey=$(echo $natsleafcert|jq '."tls.key"')
echo ${tlskey//\"}|base64 -d > $natsLeafTlsPath"/tls.key"

#create nats  configurations
cat>$rootPath"nats.conf"<<EOF
cluster {
      port: 6222
      connect_retries: 30
}
tls: {
    verify:   true
    ca_file: $edgeServerTlsPath/ca.crt
    cert_file: $edgeServerTlsPath/tls.crt
    key_file: $edgeServerTlsPath/tls.key
 },
leafnodes {
  remotes = [
        {
          url: "nats-leaf://$natsMasterDNS:$natsMasterLeafPort"
          tls: {
            cert_file: "$natsLeafTlsPath/tls.crt"
            key_file: "$natsLeafTlsPath/tls.key"
            ca_file: "$natsLeafTlsPath/ca.crt"
          }
        },
  ]

}
EOF

#create nats-streaming configuration
cat>$rootPath"stan.conf"<<EOF
streaming {
      id: test-cluster
      store: file
      dir: /data/stan/store
      ft_group_name: "$mecID-cluster"
      file_options {
          buffer_size: 32mb
          sync_on_flush: false
          slice_max_bytes: 512mb
          read_buffer_size:  64mb
          parallel_recovery: 64
      }
      nats_server_url: nats://$internalNatsDNS:4222
      partitioning: true
      store_limits {
          channels: {
           $mecID.>: {}
          }
          max_channels: 0
          max_msgs: 0
          max_bytes: 256gb
          max_subs: 0
      }
      tls: {
      verify:   true
      client_cert: "$edgeClientTlsPath/tls.crt"
      client_key: "$edgeClientTlsPath/tls.key"
      client_ca: "$edgeClientTlsPath/ca.crt"
     }
}
EOF

#start nats-server
docker run --name nats -idt -p 4222:4222  -p 7777:7777 -v /etc/hosts:/etc/hosts  -v $rootPath:$rootPath  --mount type=tmpfs,destination=/var/run/nats/   --entrypoint nats-server  $natsDockerImage --config $rootPath"nats.conf"

#start nats-streaming
docker run --name streaming -idt -p 8222:8222 -v $rootPath:$rootPath -v /data/:/data/ -v /etc/hosts:/etc/hosts $streamingDockerImage -sc $rootPath"stan.conf"





