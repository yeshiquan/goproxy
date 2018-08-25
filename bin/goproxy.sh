#!/bin/bash
BASE_DIR=$(cd $(dirname $BASH_SOURCE)/../; pwd)

LOG_DIR=$HOME/var/vs-edu/log/goproxy

usage(){
    /usr/bin/clear
    echo -ne "\33[1mNAME\33[0m\n"
    echo -ne "\tgoproxy.sh - manager goproxy\n\n"
    echo -ne "\33[1mDESCRIPTION\33[0m\n\n"
    echo -ne "\tgoproxy.sh start\n"
    echo -en "\t\t- start goproxy\n\n"
    echo -ne "\tgoproxy.sh stop\n"
    echo -en "\t\t- stop goproxy\n\n"
    echo -ne "\tgoproxy.sh restart\n"
    echo -en "\t\t- restart goproxy\n\n"
    echo -ne "\tgoproxy.sh status\n"
    echo -ne "\t\t- show status of goproxy\n\n" 
    echo -ne "\tgoproxy.sh h|-h|--help \n"
    echo -en "\t\t- show usage\n\n"
    exit 1
}

running_count() {
    number=$(ps aux | grep -v "grep" | grep -v "goproxy.sh" | grep "goproxy" | wc -l)
    echo $number
}

start() {
    mkdir -p $LOG_DIR
    number=$(running_count)
    if [ $number -gt 0 ];
    then
        echo "goproxy is running, start failed"
    else
        export DATA_PLATFORM_ENV=product
        nohup supervise -p /home/work/var/vs-edu/status/goproxy -f $BASE_DIR/bin/goproxy>> $LOG_DIR/goproxy_nohup.log 2>&1 &
        RET=$?
        if [ $RET -eq 0 ];then
            echo "goproxy started, log path $LOG_DIR/goproxy_nohup.log"
        else
            echo "goproxy start failed"
        fi
    fi
}
 
stop() {
    number=$(running_count)
    if [ $number -gt 0 ]
    then
        ps gaux | grep "goproxy" | grep "supervise" | grep -v grep | grep -v "goproxy.sh" | awk '{print $2}' | xargs kill
        ps gaux | grep "goproxy" | grep -v grep | grep -v "goproxy.sh" | awk '{print $2}' | xargs kill
        sleep 1
    fi
    echo "goproxy stoped"
}

restart() {
    number=$(running_count)
    if [ $number -gt 0 ]
    then
        ps gaux | grep "goproxy" | grep -v grep | grep -v "goproxy.sh" | awk '{print $2}' | xargs kill -s SIGHUP
        sleep 1
    fi
    echo "goproxy restart done"
}

case "$1" in
    start)
        if [ $# -ne 1 ]
        then
            usage
        fi
        start
        ;;
     
    stop)
        if [ $# -ne 1 ]
        then
            usage
        fi
        stop
        ;;
    restart)
        restart
        ;;
     
    status)
        number=$(running_count)
        if [ $number -gt 0 ]
        then
            echo "goproxy is running"
        else
            echo "goproxy is not running"
        fi
        ;;

    h|-h|--help)
        usage
        ;;
    *)
        usage
        ;;
esac
