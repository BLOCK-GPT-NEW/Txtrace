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
- call trace: from\to\function hash\gas\value\input\output   output没有，input需要配合合约abi解析，其余已获取
- state trace:read/write\key\value 已获取
- log trace:contract hash\event hash\log data 已获取

## 注意事项
- 当前mongodb端口为27018

## 10.12 
- 端口自定义问题，~~两张表合并问题~~，不影响使用
- 提取的数据需要再拿出来进行解析，才能做bedding
- bug：
    - tracegloabl缓冲区可能有问题，私链测试下异常，主网同步区块到51380出现错误（目前已将代码修改至和Txspector一样，但是待测试）
    - opcode中call相关指令存在递归调用，我按照txspctor中的代码修改了指令返回信息，但是可能不是我们所需要的，还需要再分析
