#!/usr/bin/env python
#encoding=gb18030
import sys, json, re, traceback 

query_type_dict = {
    '-1':u'注册',
    '0':u'语音',
    '1':u'键盘',
    '2':u'编辑',
    '3':u'引导',
    '4':u'重发',
    '6':u'回访',
    '8':u'导航栏',
}

system_dict = {
    'androidapp':'android',
    'android':'android',
    'iosapp':'ios',
    'ios':'ios'
}

client_dict = {
    'androidapp':'app',
    'android':'shoubai',
    'iosapp':'app',
    'ios':'shoubai',
}

def get_line_data(line):
    p2 = line.find(']', 2)
    if p2 < 0: return None
    miao = line[ :p2].rfind(':')
    if miao < 0: return None
    dt = line[1:miao]
    p1 = line.find('[', p2)
    if p1 < 0: return None
    logid = line[p2+1:p1].strip()
    p2 = line.find(']', p1)
    if p2 < 0: return None
    state = line[p1+1:p2]
    #if state != 'INFO': return None
    p1 = line.find('{', p2)
    if p1 < 0: return None
    p2 = line.rfind('}')
    if p2 < 0: return None
    data = line[p1:p2+1]
    try:
        jdata = json.loads(data.decode('gb18030','ignore'))#.encode('utf8', 'ignore'))
    except Exception, ex:
        print >> sys.stderr, "json decode error: " + str(ex)
        return None
        #try: 
        #    jdata = json.loads(data.replace(", }", "}"), encoding='gb18030') 
        #except Exception, ex:
        #    print >> sys.stderr, "json decode error: " + str(ex)
        #    return None
    return (dt, logid, jdata)

def print_extr_res(dt, logid, jdata):
    uid = baidu_uid = dumi_id = service_id = query_type = isNewUser = generator = kefu_id = query = result_list_json = hints_json = 'null'
    client_info = client = os = is_login = location_json = operation_system = app_ver = im_type = 'null'
    debug = '0'
    time_use_dict = {}
    time_use_json = 'null'
    dt = 'null'
    msg_logid = ''
    try:
        request_type = jdata.get('request_type', 'null')
        #if request_type == 'direct':return 1
        baidu_uid = jdata.get('baidu_uid', 'null')
        uid = jdata.get('user_id', 'null')

        time_use_dict['fetch_user_time'] = jdata.get('fetch_user_time', 'null')
        #time_use_dict['timeuse_us_fetch'] = jdata.get('timeuse_us_fetch', 'null')
        #time_use_dict['timeuse_us_return'] = jdata.get('timeuse_us_return', 'null')
        #time_use_dict['timeuse_intervene'] = jdata.get('timeuse_intervene', 'null')
        time_use_dict['timeuse_ui_return'] = jdata.get('timeuse_ui_return', 'null')
        time_use_json = json.dumps(time_use_dict, ensure_ascii=False).replace('\\n', '[n]')

        if 'client_info' in jdata:
            client_info = jdata['client_info']
        elif 're_client_info' in jdata:
            client_info = jdata['re_client_info']
        else:
            pass

        req = jdata['ori_post_data']
        msg_logid = req.get('logid', msg_logid)
        is_login = req.get('is_login', 'null')
        dumi_id = req.get('dumi_id', 'null')
        if uid == 'null':
            uid = req.get('user_id', 'null')
        if 'query' in req:
            query = req['query'].replace('\n', '[n]').replace('\t', '[t]')

        if 'location' in req:
            location_json = json.dumps(req['location'], ensure_ascii=False).replace('\\n', '[n]')
        # about msg
        if 'msg' not in req:
            print >> sys.stderr, 'logid: %s do not have msg data' %logid
        else:
            msg = req['msg']
            if msg_logid == '':
                msg_logid = msg.get('logid', msg_logid)
            if msg_logid != '' and msg_logid != logid:
                print >> sys.stdout, logid + '\t' + msg_logid

            service_id = msg.get('service_id', 'null')
            query_type = msg.get('query_type', 'null')
            query_type = query_type_dict.get(query_type, query_type)
            request_type = msg.get('request_type', 'null') #request_type in different place means different
            if request_type == '12':
                query_type = u'点击'
            app_ver = msg.get('app_ver', 'null')
            operation_system = msg.get('operation_system', 'null')
            im_type = msg.get('im_type', 'null')
            isNewUser = msg.get('isNewUser', 'null')
            if uid == 'null' or uid == '' or uid is None:
                uid = msg.get('user_id', 'null')
            if uid == 'null' or uid == '' or uid is None:
                uid = msg.get('request_uid', 'null')
            if uid == 'null':
                uid = msg.get('CUID', 'null')
            if 'location' in msg:
                location_json = json.dumps(msg['location'], ensure_ascii=False).replace('\\n', '[n]')
            if 'client_info' in msg:
                client_info = msg['client_info']

        if client_info != 'null' and client_info != '':
            client = client_info.get('client_from', 'null_null')
            clpos = client.rfind('_')
            if clpos < 0:
                os = client
            else:
                os = client[clpos + 1:]
                client = client[:clpos]
            if app_ver == 'null':
                app_ver = client_info.get('app_ver', 'null')
        elif operation_system != 'null':
            client = client_dict.get(operation_system)
            os = system_dict.get(operation_system)
        else:
            pass

        out_line = [uid, dt, logid, time_use_json, debug, is_login, isNewUser, baidu_uid, dumi_id, client, app_ver, os, location_json, service_id, query_type, im_type, generator, kefu_id, query, result_list_json, hints_json]
        for i, s in enumerate(out_line):
            if s is None or s is 'null':
                out_line[i] = ''
            elif isinstance(s, unicode):
                out_line[i] = str(s.encode('gb18030'))
            elif not isinstance(s, str):
                out_line[i] = str(s)
            else:
                pass
        print >> sys.stdout, '\t'.join(out_line)

    except Exception as ex:
        print >> sys.stderr, '[Error] failed to get query log: ' + str(ex)
        traceback.print_exc()
        print >> sys.stderr, 'result: ' + json.dumps(jdata, ensure_ascii=False).encode('gb18030', 'ignore')


#trace_log = open('ui_log/raw_ui_log.trace', 'aw')
for line in sys.stdin:
    PC_FLAG = False
    line = line.strip().decode('utf8', 'ignore').encode('gb18030', 'ignore')
    if ("[WARN]" in line) : continue
    if '"request_type":"pc"' in line : PC_FLAG = True
    if '_monitor"' in line: continue
    res = get_line_data(line)
    if not res : 
        print >> sys.stderr, '[WARN] do not get line data'
        continue
    elif PC_FLAG : 
        print >> sys.stderr, '[WARN] request_type == pc, logid: %s ' % res[1]
        #print >> trace_log, line
        continue
    #print >> trace_log, line
    else:
        dt, logid, jdata = res
        if logid.isdigit(): 
            print >> sys.stderr, '[WARN] digit logid: %s' % logid
            continue
        print_extr_res(dt, logid, jdata)


