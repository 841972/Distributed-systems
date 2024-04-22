#jp2a --colors gopher.jpg

echo "💣 Deleting cluster"
kind delete cluster &>/dev/null

echo "✨ Creating cluster"
./kind-with-registry.sh

docker run -d -p 5000:5000 --restart=always --name registry registry:2

echo "🖼️ Building docker image"
docker build -t localhost:5000/server:latest .

docker push localhost:5000/server:latest

echo "🚀 Launching Kubernetes"
#kubectl delete statefulset raft #&>/dev/null
kubectl create -f deployment.yaml