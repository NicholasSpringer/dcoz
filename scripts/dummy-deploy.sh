docker push springern/dcoz-eval-dummy:latest
kubectl delete deployment dummy-1 dummy-2
kubectl apply -f ./dummy.yaml
kubectl get pods --selector=name=dummy-1 --field-selector=status.phase!=Terminating
kubectl get pods --selector=name=dummy-2 --field-selector=status.phase!=Terminating