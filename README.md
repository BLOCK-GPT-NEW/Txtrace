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
- mongodb->geth->transaction 记录交易基础信息
- mongodb->geth->trace 记录需要的指令trace，如sstore中key value

## 要获取信息如下 
- call trace: from\to\function hash\gas\value\input\output   output没有，input需要配合合约abi解析，其余已获取
- state trace:read/write\key\value 已获取
- log trace:contract hash\event hash\log data 已获取

## 注意事项
- 当前mongodb端口为27018

## 10.11 
- 端口自定义问题，两张表合并问题 
- bug：
    - 指令中四个call有问题（已修改，未测试） 
    - tracegloabl缓冲区可能有问题，私链测试下异常，主网同步区块到51380出现错误（切片访问错误）
    - 目前将一个交易hash绑定一个完整的trace，但是一个交易hash可能涉及到多个调用，所以数据表里trace中tx_hash字段可能会和tx_trace对应不上
