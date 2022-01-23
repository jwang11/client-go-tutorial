# client-go-tutorial
k8s client-go tutorial

## usage
```diff
$ go mod init mod_name
+ v.0.xx.x depends on your k8s version
$ go get k8s.io/apimachinery@v0.22.3
$ go get k8s.io/client-go@v0.22.3
$ go mod tidy

+ run sample
$ go run xxx.go
