ibmcloud login -a cloud.ibm.com -r us-south -g Default --apikey <cloud-key>
ibmcloud ks cluster config --cluster bseq90dd0hbo035bbs5g
kubectl config current-context
ibmcloud cr region-set us-south
ibmcloud cr namespaces
ibmcloud cr login
docker tag flooopy us.icr.io/ishmeetreg/myrepos:latest
ibmcloud cr image-list
docker push us.icr.io/ishmeetreg/myrepos:latest
ibmcloud cr image-list
kubectl run flooopy --image=us.icr.io/ishmeetreg/myrepos:latest
kubectl get pods
kubectl expose deployment/flooopy --type=NodePort --port=8080 --name=flooopy-service --target-port=8080
