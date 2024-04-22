'''rm Dockerfiles/servidor/srvraft
rm Dockerfiles/cliente/cltraft

echo -e "\nCompilando los ficheros Golang"
cd raft/cmd/srvraft
CGO_ENABLED=0 go build -o ///Dockerfiles/servidor/srvraft main.go
cd //pkg/cltraft
CGO_ENABLED=0 go build -o ///Dockerfiles/cliente/cltraft cltraft.go

echo -e "\nCreando las imágenes Docker"
cd ///Dockerfiles/servidor
docker build . -t localhost:5001/servidor:latest
docker push localhost:5001/servidor:latest
cd /cliente
docker build . -t localhost:5001/cliente:latest
docker push localhost:5001/cliente:latest

cd /

echo -e "\nLanzando Kubernetes"
kubectl delete statefulset raft &>/dev/null
kubectl delete pod client &>/dev/null
kubectl delete service raft-service &>/dev/null
kubectl create -f statefulset_go.yaml'''