
installFilePath=./install.sh
installContainerFunc(){
    nats=$(docker ps -f name=nats)
    streaming=$(docker ps -f name=streaming)
    if [ -z "$nats" ] || [ -z "$streaming" ]
    then
     echo "installing swarm containers "
     sh $installFilePath $1 $2
    fi
}

uninstallContainerFunc(){
    docker stop nats streaming
    docker rm nats streaming
}

while [ 1 ]
do
      sleep 10
      data=$(curl -s 127.0.0.1:8088/api/v1/registering/register?watch=0)
      cloudMasterIp=""
      datacenter_id=$(echo $data|jq '.datacenter_id')
      register_status=$(echo $data|jq '.register_status')
      if [ $register_status == 1 ]; then
         installContainerFunc $cloudMasterIp $datacenter_id
      else
         uninstallContainerFunc
      fi
done
