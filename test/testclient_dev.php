<?php
include('./NsHead.php');

$socket = socket_create ( AF_INET, SOCK_STREAM, SOL_TCP );
socket_set_option ( $socket, SOL_SOCKET, SO_SNDTIMEO, array (
        'sec' => 0,
        'usec' => 30000 
) );

socket_connect ( $socket, '10.95.31.38', 8878 );

if ($argv[1] == 'get') {
    $params = array(
        'bucket'=>'shorturl',
        'method'=>'get',
        'params'=>'ui:edu:shorturl:425bea31311e90cb69f52170ee29d4c9',
    );
}

if ($argv[1] == 'mget') {
    $params = array(
        'bucket'=>'default',
        'method'=>'mget',
        'params'=>array('key1', 'key3', 'key5', 'key6'),
    );
}

if ($argv[1] == 'set') {
    $params = array(
        'bucket'=>'default',
        'method'=>'set',
        'params' => array(
            'key1' => 23.77
        )
    );
}

if ($argv[1] == 'mset') {
    $params = array(
        'bucket'=>'default',
        'method'=>'mset',
        'params'=>array(
            'key1' => array(
                'name' => 'edu1',
                'title' => 'edutitle1',
            ),
            'key2' => array(
                'name' => 'edu2',
                'title' => 'edutitle2',
            ),
        )
    );
    for ($i = 3; $i <= 15; $i++) {
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

$start = microtime(true);
$nh = new \NsHead ();
$nh->nshead_write ( $socket, $head, $pack );
$rtn = $nh->nshead_read ( $socket, false );
$end = microtime(true);
$cost = ($end - $start)*1000;
echo "elapsed $cost ms\n";
//print_r ( mc_pack_pack2array ( $rtn ['buf'] ));
file_put_contents("/home/yeshiquan/ddd.dat", ($rtn['buf']) . PHP_EOL);
print_r ( json_decode($rtn['buf'], true));
