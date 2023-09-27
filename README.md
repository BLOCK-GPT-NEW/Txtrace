# Txtrace
## 修改的文件
### txpool.go
### state_processor.go
### interpreter.go
### instructions.go
### blockchai.go
### geth启动命令
`` sudo nohup geth --datadir data/ --networkid 5555 --http --http.addr 0.0.0.0 --http.port 8545 --http.corsdomain "*" --http.api "eth,web3,personal,net" --ws --ws.addr 0.0.0.0 --ws.port 8546 --ws.origins "*" --port 30306 --rpc.enabledeprecatedpersonal > geth.log 2>&1 ``
