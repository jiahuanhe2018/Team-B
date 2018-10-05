## 操作说明

### 编译可执行文件
项目cmd 目录执行go build main.go

###前提
go run main.go -c account lzhx_ listaddresses
-t 参数用于区分是pow还是pos
### pow
go run main.go -c chain -t pow -s lzhx_ -l 8080 -a XXXXXXXXXXXXXXXXXX(address)
然后根据提示在另一个终端继续go run main.go  XXXXXXXXXXX
### pos
go run main.go -c chain -t pos -s lzhx_ -l 8080 -a XXXXXXXXXXXXXXXXXX(address)
然后根据提示在另一个终端继续go run main.go  XXXXXXXXXXX