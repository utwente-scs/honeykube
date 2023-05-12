# Setup Audit Logs Collection

1. Move audit-policy.yaml to Master Node location:

```
docker cp audit-policy.yaml <container-id>:/etc/kubernetes/audit-policy.yaml
```

2. Move kube-apliserver.yaml to Master Node location:

```
docker cp kube-apiserver.yaml <container-id>:/etc/kubernetes/manifests/
```


This setup will store the audit logs in the directory in the Master Node:

```
/var/log/audit.log
```
