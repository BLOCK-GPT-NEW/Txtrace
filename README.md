# Txtrace
## 修改的文件
- txpool.go
- state_processor.go
- interpreter.go
- instructions.go
- blockchai.go

## geth启动命令
- sudo nohup geth --datadir data/ --networkid 5555 --http --http.addr 0.0.0.0 --http.port 8545 --http.corsdomain "*" --http.api "eth,web3,personal,net" --ws --ws.addr 0.0.0.0 --ws.port 8546 --ws.origins "*" --port 30306 --rpc.enabledeprecatedpersonal > geth.log 2>&1 & 

- geth --datadir ./execution  --networkid 12345 --http --http.api eth,web3,net,debug --http.addr "0.0.0.0" --http.port 8545 --http.corsdomain "*"  --http.vhosts "*" --miner.etherbase "0xd815AE3BAfA1A7e3B1b1Ac096152c56b1a9BA97d" --mine --allow-insecure-unlock  --unlock "0xd815AE3BAfA1A7e3B1b1Ac096152c56b1a9BA97d" --txpool.pricelimit 0 --txpool.pricebump 0 --nodiscover --password "./tmp/password.txt"

## 数据表
- mongodb->geth->transaction 记录交易基础信息与trace信息

## 要获取信息如下 
- call trace: from\to\function hash\gas\value\input\output 在trace中，需要再分析
- state trace:read/write\key\value 在trace中
- log trace:contract hash\event hash\log data 已获取

## 注意事项
- 当前mongodb端口为27018

