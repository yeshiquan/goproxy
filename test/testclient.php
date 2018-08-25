<?php
include('./NsHead.php');
include('./color.php');

if($argc < 2){
	yellow("useage: \n");	
	red("\t php testclient.php ");
	green("[get | mget | set | mset]\n");
	endcolor();
	die();
}

if ($argv[1] == 'shorturl') {
    $params = array(
        'bucket'=>'short_addr',
        'method'=>'set',
        'params' => array(
            'ee6e51bb5ef1d193e14d975117bb1fd4' => array(
                "type" => 2,
                "url" =>
                "http://jiaoyu.baidu.com/m/orgDetail?key=%E4%BC%9A%E8%AE%A1%E7%BD%91%E6%A0%A1&originQuery=%E4%BC%9A%E8%AE%A1%E7%BD%91%E6%A0%A1&ap=YVdROU1URTRPVEV5Sm1KeVp6MHdKblZ6WlhKcFpEMHhPREF6TkRBd0puQnNZVzVwWkQweE16WXlPVFF5TXlaMWJtbDBhV1E5TXpneE1UY3hOVEkxSm5CeWFXTmxQVGN3Sm1SbGMyTnBaRDB4TmpNMU16STJPVE01Sm5kcGJtWnZhV1E5TUNaaWFXUTlNekF3Sm5KaGJtczlNU1ozYjNKa2FXUTlNQ1oxYVY5eGRXVnllVDBsUlRRbFFrTWxPVUVsUlRnbFFVVWxRVEVsUlRjbFFrUWxPVEVsUlRZbFFUQWxRVEVtZDIxaGRHTm9QVEFtWkhKaGRHVTlPREF5T1RJbVkyOXpkRjl5WVc1clBURW1ZMjFoZEdOb1BUSXdORGttY0dOZmQzVnliRDBtYlc5aWFXeGxYM1Z5YkQwPQ%253D%253D&apm=144ea333bea078f207d3594671592e83&qid=1436113394340151242&pagefrom=orgmp&zt=pswise&city=%E5%A4%A9%E6%B4%A5&orgId=9247#/",
            ),
        ),
    );
}

if ($argv[1] == 'get') {
    $params = array(
        'bucket'=>'shangmao_description',
        'bucket_type'=>'default',
        'method'=>'get',
        'params'=>'414507478',
    );
}

if ($argv[1] == 'mget') {
    $params = array(
        'bucket'=>'default',
        'method'=>'mget',
        'params'=>array('key1', 'key3', 'key5', 'key6', 'keynotfound'),
    );
}

if ($argv[1] == 'set') {
    $params = array(
        'bucket'=>'default',
        'method'=>'set',
        'params' => array(
            'key1' => '<p><br/>供应印台毡|印台毡价格|印台毡厂家销售供应印台毡|印台毡价格|印台毡厂家销售</p>magicHcReturnmagicHcLine<p>公司主要生产、制作各种规格的毛毡及毛毡制品，如：细白工业毛毡、中粗毛毡、乐器专用毡、抛光毡、吸油毡、彩色毛毡、化纤毡、羊毛抛光轮、毛毡条、毛毡环、毛毡垫、毛毡筒、油封圈、羊毛球、兔毛球、羊毛盘、海绵轮、军用鞋垫、羊剪绒汽车坐垫、床毯、沙发坐垫、春羔皮行等5000多个品种。并可根据用户要求制作各种异形毡零件。产品主要用于：航天、航空、军工、机械、机电、化工、矿山、水泥、造纸、磁性材料、皮革、文具、漆包线、车辆船舶、纺织器械、各种电器、玩具、电子产品。也可用于大理石、不锈钢、玻璃、陶瓷、精密家具抛光，汽车、轮船、机械、机床等设备防尘、密封、隔音、保温、绝缘、绝热等。&nbsp;<imgsrc="http://img22.hc360.cn/22/product/393/079/b/22-39307912.jpg" alt="" width="300" height="225" /><brclass="img-brk" /><br class="img-brk" /><imgsrc="http://img03.hc360.cn/03/product/393/079/b/03-39307953.jpg" alt="" width="300" height="211" /><brclass="img-brk" /><br class="img-brk" /></p>',
        )
    );
}

if ($argv[1] == 'mset') {
    $params = array(
        'bucket'=>'default',
        'method'=>'mset',
        'params'=>array(),
    );
    for ($i = 1; $i <= 15; $i++) {
        $params['params']["key$i"] = array(
            'name' => "edu$i",
            'title' => "edutitle$i",
        );
    }
}

//$pack = mc_pack_array2pack ( $params, PHP_MC_PACK_V2 );
$pack = json_encode($params);
$head = array (
    'id' => 1,
    'provider' => 'ecom-edu-ui',
    'version' => 1,
    'log_id' => 12323423,
    'reserved' => 0,
    'magic_num' => 0xfb709394,
    'body_len' => strlen ( $pack ) 
);
$socket = socket_create ( AF_INET, SOCK_STREAM, SOL_TCP );
socket_set_option ( $socket, SOL_SOCKET, SO_SNDTIMEO, array (
        'sec' => 0,
        'usec' => 30000 
) );
socket_connect ( $socket, '10.67.25.11', 8877 );
$start = microtime(true);
$nh = new \NsHead ();
$nh->nshead_write ( $socket, $head, $pack );
$rtn = $nh->nshead_read ( $socket, false );
//print_r($rtn);
$end = microtime(true);
$cost = ($end - $start)*1000;
//print_r ( mc_pack_pack2array ( $rtn ['buf'] ));
//print_r ( ($rtn['buf']) . PHP_EOL);
file_put_contents("testresult.dat", ($rtn['buf']) . PHP_EOL);
print_r ( json_decode($rtn['buf'], true));
echo "elapsed $cost ms\n";
