
Start sysdig -
```
sudo docker run -it --name sysdig --privileged --net=host -v /var/run/docker.sock:/host/var/run/docker.sock -v /dev:/host/dev -v /proc:/host/proc:ro -v /boot:/host/boot:ro -v /lib/modules:/host/lib/modules:ro -v /usr:/host/usr:ro -v /etc:/host/etc:ro -v /tmp/tracefiles/:/tmp/tracefiles -v /tmp/app:/tmp/app -e SYSDIG_BPF_PROBE="" sysdig/sysdig
```

```
sudo docker start sysdig
```