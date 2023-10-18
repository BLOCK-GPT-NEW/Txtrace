import json
import re
import time
import pymongo
import multiprocessing


# 逐条处理记录
def process_record(record):

# 遍历每个记录
# for record in data:
    output_data = []
    start_time = time.time()

    # 读取每一条记录的数据
    tx_hash = record["tx_hash"]
    tx_fromaddr = record["tx_fromaddr"]
    tx_toaddr = record["tx_toaddr"]
    tx_gas = record["tx_gas"]
    tx_value = record["tx_value"]
    tx_input = record["tx_input"]
    tx_trace = record["tx_trace"]
    tx_status = record["re_status"]
    re_contract_address = record["re_log_address"]
    re_log_topics = record["re_log_topics"]
    log_trace = record["log_trace"]

    # 计算每个trace中含有“CALL"的个数count,每个交易hash应该对应count+1个call记录
    # count = len(re.findall("CALL:", tx_trace))

    # 分割tx_trace中的"CALL"
    tx_trace_slice = re.split(';CALL;', tx_trace)

    # 记录call切片的位置
    slice_count = 0



    call_records = []

    contract_address_count = 0
    event_hash_count = 0
    log_data_count = 0
    # 输出划分后的部分
    for slice in tx_trace_slice:

        if slice_count == 0 :
            # 处理call----------------------------------------------------------------------------------------
            call_from = tx_fromaddr
            call_to = tx_toaddr
            call_function_name = tx_input[:8]
            call_value = tx_value
            call_gas = tx_gas
            call_input = []
            call_output = []

            # 处理call_input,根据input的前8bit去标识函数，可能找到，也可能找不到-----------
            call_input_func_dex = tx_input[:8]
            call_input_args_count = int(len(tx_input[8:])/64)
            func_name = ""
            # 打开4byte文件，用input前面8位去找相应的函数
            with open('4byte.json', 'r') as file:
                data = json.load(file)
            try :
                func_name = data[call_input_func_dex]
            except KeyError as e:
                for i in range(call_input_args_count):
                    call_input.append({"call_input_type":"","call_input_value":"0x"+tx_input[8+64*i:8+64*(i+1)]})
            # 如果找到了，使用正则表达式来找函数里面的参数类型并写入call_input
            matches = re.findall(r'\((.*?)\)', func_name)
            if matches:
                parameters = matches[0]
                parameter_types = parameters.split(',')
                i = 0
                for parameter_type in parameter_types:
                    call_input.append({"call_input_type":parameter_type,"call_input_value":"0x"+tx_input[8+64*i:8+64*(i+1)]})
                    i = i+1

            # 处理call_output---------------
            match_call_output = re.search(r'RETURN;(.*?)[|]', slice)
            if match_call_output:
                if match_call_output.group(1) != "": 
                    call_output_args_count = int(len(match_call_output.group(1))/64)
                    for i in range(call_output_args_count):
                        call_output.append({"call_output_type":"","call_output_value":match_call_output.group(1)[64*i:64+64*i]})

            # 处理state----------------------------------------------------------------------
            state = []

            pattern_write = r'SLOAD;(.*?)[|]'
            matches_write = re.findall(pattern_write, slice)         
            for match_write in matches_write:
                pattern_write_key = r'key:(.*?);'
                pattern_write_value = r'val:(.*?)$'
                match_write_key = re.search(pattern_write_key, match_write)
                match_write_value = re.search(pattern_write_value, match_write)
                state.append({"tag":"write","key":match_write_key.group(1),"value":match_write_value.group(1)})
            
            pattern_read = r'SSTORE;(.*?)[|]'
            matches_read = re.findall(pattern_read, slice)         
            for match_read in matches_read:
                pattern_read_key = r'key:(.*?);'
                pattern_read_value = r'val:(.*?)$'
                match_read_key = re.search(pattern_read_key, match_read)
                match_read_value = re.search(pattern_read_value, match_read)
                state.append({"tag":"read","key":match_read_key.group(1),"value":match_read_value.group(1)})

            # 处理log----------------------------------------------------------
            # log中的字符串变成数组方便后续处理
            contract_address = re_contract_address.split(",")
            event_hash = re_log_topics.split(",")
            # 查找LOG后面的数字是几
            log = []
            pattern_log = 'LOG(.*?);'
            matches_log = re.findall(pattern_log, slice)

            # 处理log_trace，将他们转换为数组，数组中的一个元素就是一次log的结果，还需要分开为32B
            log_trace_list = re.split("data:",log_trace)
            log_trace_list = log_trace_list[1:]

            for i in range(len(log_trace_list)):
                log_trace_list[i] = log_trace_list[i][0:-1]

            for match_log in matches_log:

                if match_log == "0":
                    try:
                        log.append[{"contract_address":contract_address[contract_address_count],"event_hash":"","data":""}]
                    except IndexError :
                        continue
                    log.append[{"contract_address":contract_address[contract_address_count],"event_hash":"","data":""}]

                    contract_address_count = contract_address_count + 1

                elif match_log == "1" :
                    log_data = []
                    log_data_insert = int(len(log_trace_list[log_data_count])/64)
                    for i in range(log_data_insert):
                        log_data.append({"type":"","value":"0x"+log_trace_list[log_data_count][64*i:i*64+64]})
                    try:
                        log.append({"contract_address":contract_address[contract_address_count],"event_hash":event_hash[event_hash_count],"data":log_data})
                    except IndexError :
                        continue
                    log.append({"contract_address":contract_address[contract_address_count],"event_hash":event_hash[event_hash_count],"data":log_data})
                    contract_address_count = contract_address_count + 1
                    event_hash_count = event_hash_count + 1
                    log_data_count = log_data_count + 1

                elif match_log == "2":    
                    log_data = []
                    log_data_insert = int(len(log_trace_list[log_data_count])/64)
                    for i in range(log_data_insert):
                        log_data.append({"type":"","value":"0x"+log_trace_list[log_data_count][64*i:i*64+64]})
                    try:
                        log.append({"contract_address":contract_address[contract_address_count],"event_hash":event_hash[event_hash_count],"data":log_data})
                    except IndexError :
                        continue
                    log.append({"contract_address":contract_address[contract_address_count],"event_hash":event_hash[event_hash_count],"data":log_data})
                    contract_address_count = contract_address_count + 1
                    event_hash_count = event_hash_count + 2
                    log_data_count = log_data_count + 1

                elif match_log == "3":    
                    log_data = []
                    log_data_insert = int(len(log_trace_list[log_data_count])/64)
                    for i in range(log_data_insert):
                        log_data.append({"type":"","value":"0x"+log_trace_list[log_data_count][64*i:i*64+64]})
                    try:
                        log.append({"contract_address":contract_address[contract_address_count],"event_hash":event_hash[event_hash_count],"data":log_data})
                    except IndexError:
                        continue
                    log.append({"contract_address":contract_address[contract_address_count],"event_hash":event_hash[event_hash_count],"data":log_data})
                    contract_address_count = contract_address_count + 1
                    event_hash_count = event_hash_count + 3
                    log_data_count = log_data_count + 1

                elif match_log == "4":    
                    log_data = []
                    log_data_insert = int(len(log_trace_list[log_data_count])/64)
                    for i in range(log_data_insert):
                        log_data.append({"type":"","value":"0x"+log_trace_list[log_data_count][64*i:i*64+64]})
                    try :
                        log.append({"contract_address":contract_address[contract_address_count],"event_hash":event_hash[event_hash_count],"data":log_data})
                    except IndexError:
                        continue
                    log.append({"contract_address":contract_address[contract_address_count],"event_hash":event_hash[event_hash_count],"data":log_data})
                    contract_address_count = contract_address_count + 1
                    event_hash_count = event_hash_count + 4
                    log_data_count = log_data_count + 1

            slice_count = slice_count + 1

            # 汇总到一个call_record里面---------------
            call_records.append({"call_from":call_from,"call_to":call_to,"call_function_name":call_input_func_dex,"call_gas":call_gas,"call_value":call_value,"call_input":call_input,"call_output":call_output,"stata":state,"log":log})
        else :
            # 处理call----------------------------------------------
            
            match_call_from = re.search("from addr:(.*?);",slice)
            call_from = match_call_from.group(1)
            

            match_call_to = re.search("to addr:(.*?);",slice)
            call_to = match_call_to.group(1)

            match_call_gas = re.search("gas:(.*?);",slice)
            call_gas = match_call_gas.group(1)

            match_call_value = re.search("value:(.*?);",slice)
            call_value = match_call_value.group(1)

            match_call_args = re.search('args:(.*?)\|',slice)
            call_args = match_call_args.group(1)

            call_func_name = ""
            call_input = []
            func_name = ""
            if len(call_args) != 0:
                call_func_name = call_args[:8]
                call_args_count = int(len(call_args[8:])/64)
                with open('4byte.json', 'r') as file:
                    data = json.load(file)
                try :
                    func_name = data[call_func_name]
                except KeyError as e:
                   for i in range(call_args_count):
                        call_input.append({"call_input_type":"","call_input_value":"0x"+tx_input[8+64*i:8+64*(i+1)]})
                matches = re.findall(r'\((.*?)\)', func_name)
                if matches:
                    parameters = matches[0]
                    parameter_types = parameters.split(',')
                    i = 0
                    for parameter_type in parameter_types:
                        call_input.append({"call_input_type":parameter_type,"call_input_value":"0x"+tx_input[8+64*i:8+64*(i+1)]})
                        i = i+1
            # 处理output
            match_call_output = re.search(r"Result:(.*?);", slice)
            if match_call_output:
                if match_call_output.group(1) != "": 
                    call_output_args_count = int(len(match_call_output.group(1))/64)
                    for i in range(call_output_args_count):
                        call_output.append({"call_output_type":"","call_output_value":match_call_output.group(1)[64*i:64+64*i]})
            # 处理state---------------------------------------------
            state = []

            pattern_write = r'SLOAD;(.*?)[|]'
            matches_write = re.findall(pattern_write, slice)         
            for match_write in matches_write:
                pattern_write_key = r'key:(.*?);'
                pattern_write_value = r'val:(.*?)$'
                match_write_key = re.search(pattern_write_key, match_write)
                match_write_value = re.search(pattern_write_value, match_write)
                state.append({"tag":"write","key":match_write_key.group(1),"value":match_write_value.group(1)})
            
            pattern_read = r'SSTORE;(.*?)[|]'
            matches_read = re.findall(pattern_read, slice)         
            for match_read in matches_read:
                pattern_read_key = r'key:(.*?);'
                pattern_read_value = r'val:(.*?)$'
                match_read_key = re.search(pattern_read_key, match_read)
                match_read_value = re.search(pattern_read_value, match_read)
                state.append({"tag":"read","key":match_read_key.group(1),"value":match_read_value.group(1)})

            # 处理log-------------------------------------------------
             # log中的字符串变成数组方便后续处理
            contract_address = re_contract_address.split(",")
            event_hash = re_log_topics.split(",")
            # 查找LOG后面的数字是几
            log = []
            pattern_log = 'LOG(.*?);'
            matches_log = re.findall(pattern_log, slice)

            # 处理log_trace，将他们转换为数组，数组中的一个元素就是一次log的结果，还需要分开为32B
            log_trace_list = re.split("data:",log_trace)
            log_trace_list = log_trace_list[1:]

            for i in range(len(log_trace_list)):
                log_trace_list[i] = log_trace_list[i][0:-1]

            for match_log in matches_log:

                if match_log == "0":
                    try:
                        log.append[{"contract_address":contract_address[contract_address_count],"event_hash":"","data":""}]
                    except IndexError:
                        continue
                    log.append[{"contract_address":contract_address[contract_address_count],"event_hash":"","data":""}]
                    contract_address_count = contract_address_count + 1

                elif match_log == "1" :
                    log_data = []
                    log_data_insert = int(len(log_trace_list[log_data_count])/64)
                    for i in range(log_data_insert):
                        log_data.append({"type":"","value":"0x"+log_trace_list[log_data_count][64*i:i*64+64]})

                    try:
                        log.append({"contract_address":contract_address[contract_address_count],"event_hash":event_hash[event_hash_count],"data":log_data})
                    except IndexError:
                        continue
                    log.append({"contract_address":contract_address[contract_address_count],"event_hash":event_hash[event_hash_count],"data":log_data})
                    contract_address_count = contract_address_count + 1
                    event_hash_count = event_hash_count + 1
                    log_data_count = log_data_count + 1

                elif match_log == "2":    
                    log_data = []
                    log_data_insert = int(len(log_trace_list[log_data_count])/64)
                    for i in range(log_data_insert):
                        log_data.append({"type":"","value":"0x"+log_trace_list[log_data_count][64*i:i*64+64]})
                    try :
                        log.append({"contract_address":contract_address[contract_address_count],"event_hash":event_hash[event_hash_count],"data":log_data})
                    except IndexError :
                        continue
                    log.append({"contract_address":contract_address[contract_address_count],"event_hash":event_hash[event_hash_count],"data":log_data})
                    contract_address_count = contract_address_count + 1
                    event_hash_count = event_hash_count + 2
                    log_data_count = log_data_count + 1

                elif match_log == "3":    
                    log_data = []
                    log_data_insert = int(len(log_trace_list[log_data_count])/64)
                    for i in range(log_data_insert):
                        log_data.append({"type":"","value":"0x"+log_trace_list[log_data_count][64*i:i*64+64]})
                    try :
                        log.append({"contract_address":contract_address[contract_address_count],"event_hash":event_hash[event_hash_count],"data":log_data})
                    except IndexError:
                        continue
                    log.append({"contract_address":contract_address[contract_address_count],"event_hash":event_hash[event_hash_count],"data":log_data})
                    contract_address_count = contract_address_count + 1
                    event_hash_count = event_hash_count + 3
                    log_data_count = log_data_count + 1

                elif match_log == "4":    
                    log_data = []
                    log_data_insert = int(len(log_trace_list[log_data_count])/64)
                    for i in range(log_data_insert):
                        log_data.append({"type":"","value":"0x"+log_trace_list[log_data_count][64*i:i*64+64]})
                    try:
                        log.append({"contract_address":contract_address[contract_address_count],"event_hash":event_hash[event_hash_count],"data":log_data})
                    except IndexError:
                        continue
                    log.append({"contract_address":contract_address[contract_address_count],"event_hash":event_hash[event_hash_count],"data":log_data})
                    contract_address_count = contract_address_count + 1
                    event_hash_count = event_hash_count + 4
                    log_data_count = log_data_count + 1

            # 汇总到一个call_record里面---------------
            call_records.append({"call_from":call_from,"call_to":call_to,"call_function_name":call_input_func_dex,"call_gas":call_gas,"call_value":call_value,"call_input":call_input,"call_output":call_output,"stata":state,"log":log})
        slice_count = slice_count + 1

    end_time = time.time()
    processing_time = end_time - start_time
    print(processing_time)

    output_data.append({"tx_hash":tx_hash,"call":call_records})
    # 写入新数据
    with open("trace_processed.json", "w") as output_file:
        json.dump(output_data, output_file, indent=2)

def main():
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
    cursor = collection.find(query)

    num_cores = multiprocessing.cpu_count()
    pool = multiprocessing.Pool(processes=num_cores)

    for i in range(num_cores):
        pool.apply_async(process_record, args=(i,))
    for record in cursor:
        for i in range(num_cores):
            pool.apply_async(process_record(record), args=(i,))

    pool.close()
    pool.join()



if __name__ == "__main__":
    main()