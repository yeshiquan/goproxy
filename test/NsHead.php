<?php

/**
 * @author baonh(baonenghui@baidu.com)
* Changelog:
* 20130305 修改Socket连接方式为原生socket操作集 hushiwei@baidu.com
*
* 对nshead的操作的PHP版本
*
* nshead C结构体:
* typedef struct _nshead_t
* {
*   unsigned short id;
*   unsigned short version;
*   unsigned int   log_id;
*   char           provider[16];
*   unsigned int   magic_num;
*   unsigned int   reserved;
*   unsigned int   body_len;
* } nshead_t;
*
**/
class NsHead {

    const NSHEAD_HEADER_LEN = 36;

    /**
     * 创建nshead包
     *
     * 传入的是一个表示nshead数据的array
     *
     *  array(
     *       'id'        => 0,
     *       'version'   => 0,
     *       'log_id'    => 0,
     *       'provider'  => ""
     *       'magic_num' => 0xfb709394, #魔鬼数字, 这个外部不需要填写程序自动填充
     *       'reserved'  => 0,
     *       'body_len'  => 0
     *   );
     *
     *
     *
     * 组装一个可以发送的nshead头数据包,不包括数据，数据在外部拼装
     *
     * @param $vars_arr 需要发送nshead头数据包,不包括数据
     * @return 返回一个可以发送的nshead头数据包,不包括数据，数据在外部拼装
     */
    public function build_nshead($vars_arr)
    {
        $nshead_arr = array(
                'id'        => 0,
                'version'   => 0,
                'log_id'    => 0,
                'provider'  => str_pad("", 16, "\0", STR_PAD_BOTH),
                'magic_num' => 0xfb709394, #魔鬼数字
                'reserved'  => 0,
                'body_len'  => 0
        );
        foreach ($vars_arr as $key => $value)
        {
            if (isset($nshead_arr[$key]))
            {
                $nshead_arr[$key] = $value;
            }
        }



        $nshead  = "";
        $nshead  = pack("L*", (($nshead_arr['version'] << 16) + ($nshead_arr['id'])), $nshead_arr['log_id']);
        //最多取15个字节的provider
        $nshead .= str_pad(substr($nshead_arr['provider'], 0, 15), 16, "\0");
        $nshead .= pack("L*", $nshead_arr['magic_num'], $nshead_arr['reserved']);
        $nshead .= pack("L", $nshead_arr['body_len']);

        return $nshead;
    }

    /**
     * 解析nshead包，并将buf放入返回数组的buf字段中
     * 一般的nshead数据包都是 nshead + buf
     *
     * 返回一个nshead的array,
     *
     *  array(
     *       'id'        =>
     *       'version'   =>
     *       'log_id'    =>
     *       'provider'  =>
     *       'magic_num' =>
     *       'reserved'  =>
     *       'body_len'  =>
     *       'buf' =>
     *   );

     *
     * 其中'buf' 是表示实际的数据
     *
     * @param $head 接收到的nshead 数据包
     * @param $get_buf 需要解析出后续的数据，如果get_buf == false, 'buf'不存在
     * @param 返回一个nshead结果的array
     */
    public function split_nshead($head, $get_buf = true)
    {
        $ret_arr = array(
                'id'        => 0,
                'version'   => 0,
                'log_id'    => 0,
                'provider'  => "",
                'magic_num' => 0,
                'reserved'  => 0,
                'body_len'  => 0,
                'buf'       => ""
        );

        $ret = unpack("v1id/v1version/I1log_id", substr($head, 0, 8));
        if ($ret == false) {
            return false;
        }
        foreach ($ret as $key => $value)
        {
            $ret_arr[$key] = $value;
        }
        $ret_arr['provider'] = substr($head, 8, 16);
        $ret = unpack("I1magic_num/I1reserved/I1body_len", substr($head, 24, 12));
        if ($ret == false) {
            return false;
        }
        foreach ($ret as $key => $value)
        {
            $ret_arr[$key] = $value;
        }
        //36是nshead_t结构体大小
        if ($get_buf) {
            $ret_arr['buf'] = substr($head, 36, $ret_arr['body_len']);
        }
        return $ret_arr;
    }

    /**
     *  将 nshead 通过socket $msgsocket 发送出去
     *
     *  @param $msgsocket 需要写的socket
     *  @param $vars_arr 需要发送的nshead头
     *  @param $buf 需要发送的实际数据
     *  @return 发送的实际数据长度
     */
    public function nshead_write($msgsocket, $vars_arr, $buf)
    {
        $nshead = $this->build_nshead($vars_arr);
        if ($nshead == false) {
            return false;
        }
        $nshead .= $buf;
        file_put_contents("/home/yeshiquan/mset.dat", $nshead);
        $left = strlen($nshead);
        while (true) {
            $len = socket_write($msgsocket, $nshead, $left);
            if ($len === false) {
                return false;
            }
            if ($len < $left) {
                $msg = substr($nshead, $len);
                $left -= $len;
                error_log('nshead truncated, resending ...');
            } else {
                return true;
            }
        }
        return false;
    }

    /**
     *  由 socket $msgsocket 获取nshead数据包
     *
     *  @param $msgsocket 需要接收的socket
     *  @param $nshead_check_magicnum 是否检查MAGICNUM, 默认检查
     */
    public function nshead_read($msgsocket, $nshead_check_magicnum = true)
    {
        $peername = "";
        socket_getpeername($msgsocket, $peername);
        //$token = dechex(Env::getToken());
        $token = dechex(123);
        $nshead_header = "";

        //先读nshead头部
        $left_head_length = NsHead::NSHEAD_HEADER_LEN;
        $header_read_times = 0;
        while ($left_head_length > 0)
        {
            $header_read_times++;
            $header_buf = socket_read($msgsocket, $left_head_length);

            //读到空串说明已经结束
            if ($header_buf === '') {
                break;
            }

            //如果返回FALSE说明读取失败
            if ($header_buf === FALSE) {
                $socket_error = socket_last_error($msgsocket);
                error_log($token . " read nshead error" . $socket_error);
                return false;
            }

            $nshead_header .= $header_buf;
            $left_head_length -= strlen($header_buf);
            if ($left_head_length < 0) {
                $socket_error = socket_last_error($msgsocket);
                error_log($token . " read nshead > " . NsHead::NSHEAD_HEADER_LEN . " bytes [$socket_error]");
                return false;
            }

             
        }

        //看是否还有剩余字节没有读完
        if ($left_head_length > 0) {
            error_log($token .  " recv header: peer reset before EOF $peername");
            return false;
        }


        $nshead = $this->split_nshead($nshead_header, false);

        //检查 magic num
        if ($nshead_check_magicnum == true
                && $nshead['magic_num'] != 0xfb709394
                && $nshead['magic_num'] != -76508268)
                //部分php版本在unpack的时候存在bug，所以这里再判断一下负数的情况
        {
            error_log($token . " magic num mismatch: ret ".$nshead['magic_num']." want 0xfb709394");
            return false;
        }

        //读nshead的数据
        $left_body_length = $nshead['body_len'];
        $body_read_times = 0;
        while ($left_body_length > 0) {
            $body_read_times++;
            $body_buf = socket_read($msgsocket, $left_body_length);

            //读到空串说明已经结束
            if ($body_buf === '') {
                break;
            }

            //返回False说明错误
            if ($body_buf === false) {
                $socket_error = socket_last_error($msgsocket);
                error_log($token . " recv buff error" . $socket_error);
                return false;
            }

            $nshead['buf'] .= $body_buf;
            $left_body_length -= strlen($body_buf);;

            if ($left_body_length < 0) {
                $socket_error = socket_last_error($msgsocket);
                error_log($token . " read nsbody > " . $nshead['body_len'] . " bytes [$socket_error]");
                return false;
            }
        }

        if ($left_body_length > 0) {
            error_log($token .  " recv body: peer reset before EOF $peername");
            return false;
        }
        return $nshead;
    }
}
