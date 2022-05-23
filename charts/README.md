#  Kubean Helm Charts

## kubean charts 调试
``` shell
helm install kubean ./charts -n kubean-system --create-namespace --debug --dry-run
```

## kubean charts 安装
``` shell
helm upgrade --install --create-namespace --cleanup-on-fail  kubean kubean_release/kubean -n kubean-system
```
