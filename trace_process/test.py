import pymongo
import time
# 远程MongoDB服务器的连接信息

# 远程连接服务器中mongodb，选中transaction集合
database_name = 'geth'
client = pymongo.MongoClient(host="10.12.46.33", port=27018,username="b515",password="sqwUiJGHYQTikv6z")
db = client[database_name]
collection = db['transaction']


# 查询规则
query = {
    "tx_blocknum": {"$gt": 4000000, "$lt": 4100000},
    "tx_trace": {"$ne": ""}
}

# 查询记录，返回游标
cursor = collection.find(query)

# 逐条处理记录
for document in cursor:
    print(document)
    
