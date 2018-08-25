<?php
/***************************************************************************
 * 
 * Copyright (c) 2015 Baidu.com, Inc. All Rights Reserved
 * 
 **************************************************************************/
 
 
 
/**
 * @file color.php
 * @author wangfei19(com@baidu.com)
 * @date 2015/05/26 15:42:01
 * @brief 
 *  
 **/


$red = "\x1b[1;31;40m";
$yellow = "\x1b[1;33;40m";
$green = "\x1b[1;32;40m";
$normal = "\x1b[0m";

function red($str){
    global $red;
    global $normal;
    echo "{$red}{$str}{$normal}";
}
function yellow($str){
    global $yellow;
    global $normal;
    echo "{$yellow}{$str}{$normal}";
}


function green($str){
    global $green;
    global $normal;
    echo "{$green}{$str}{$normal}";
    echo "\r\n\r\n";
}

function greenstart($str){
    global $green;
    global $normal;
    echo "{$green}{$str}";
    echo "\r\n\r\n";
}

function endcolor(){
    global $normal;
    echo $normal;
}




/* vim: set expandtab ts=4 sw=4 sts=4 tw=100: */
?>
