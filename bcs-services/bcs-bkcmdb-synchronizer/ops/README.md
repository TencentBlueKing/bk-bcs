1. helm install bcs-bkcmdb-synchronizer-dev -f values-dev.yaml --namespace bcs-system ./helm-chart/bcs-bkcmdb-synchronizer
2. helm upgrade bcs-bkcmdb-synchronizer-dev -f values-dev.yaml --namespace bcs-system ./helm-chart/bcs-bkcmdb-synchronizer
3. helm install bcs-bkcmdb-synchronizer-debug -f values-debug.yaml --namespace bcs-system ./helm-chart/bcs-bkcmdb-synchronizer
4. helm upgrade bcs-bkcmdb-synchronizer-debug -f values-debug.yaml --namespace bcs-system ./helm-chart/bcs-bkcmdb-synchronizer
5. helm install bcs-bkcmdb-synchronizer-prod -f values-prod.yaml --namespace bcs-system ./helm-chart/bcs-bkcmdb-synchronizer
6. helm upgrade bcs-bkcmdb-synchronizer-prod -f values-prod.yaml --namespace bcs-system ./helm-chart/bcs-bkcmdb-synchronizer
7. kubectl scale --replicas=0 
8. helm install bcs-bkcmdb-synchronizer -f values-bkdev.yaml --namespace bcs-system ./bcs-bkcmdb-synchronizer
9. helm upgrade bcs-bkcmdb-synchronizer -f values-bkdev.yaml --namespace bcs-system ./bcs-bkcmdb-synchronizer --install
10. helm install bcs-bkcmdb-synchronizer-stag -f values-stag.yaml --namespace bcs-system ./helm-chart-0918/bcs-bkcmdb-synchronizer
11. helm upgrade bcs-bkcmdb-synchronizer-stag -f values-stag.yaml --namespace bcs-system ./helm-chart-0918/bcs-bkcmdb-synchronizer --install
12. helm upgrade bcs-bkcmdb-synchronizer-stag -f values-stag.yaml --namespace bcs-system ./helm-chart-1111/bcs-bkcmdb-synchronizer --install
12. helm install bcs-bkcmdb-synchronizer -f values-bkop.yaml --namespace bcs-system ./bcs-bkcmdb-synchronizer
13. helm upgrade bcs-bkcmdb-synchronizer -f values-bkop.yaml --namespace bcs-system ./bcs-bkcmdb-synchronizer --install
14. helm install bcs-bkcmdb-synchronizer -f values-prod.yaml --namespace blueking-prod ./bcs-bkcmdb-synchronizer
15. helm upgrade bcs-bkcmdb-synchronizer -f values-prod.yaml --namespace blueking-prod ./bcs-bkcmdb-synchronizer --install
16. helm install bcs-bkcmdb-synchronizer -f values-ipv6.yaml --namespace bcs-system ./bcs-bkcmdb-synchronizer
17. helm upgrade bcs-bkcmdb-synchronizer -f values-ipv6.yaml --namespace bcs-system ./bcs-bkcmdb-synchronizer --install