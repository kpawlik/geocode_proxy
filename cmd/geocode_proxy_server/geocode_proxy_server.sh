#!/bin/sh

here=$(dirname $0)

action=$1
logdir=$2
process_name=geocode_proxy_server

function is_runnung(){
    RESULT=`pgrep -f "${process_name} -config"`
    if [ "${RESULT:-null}" != null ]
    then
        echo ${RESULT}
        return
    fi
    echo "0"
    return
}

function start(){
    pid=$(is_runnung)
    if [ "${pid}" != "0" ]
    then
        echo "Process $process_name already running"
        status
        return
    fi
    if [ -z "${logdir}" ]
    then
        logdir=$here
    fi
    echo "Starting process..."
    nohup $here/geocode_proxy_server -config $here/config.json >> $logdir/geocode_proxy_server.log 2>&1 &
    sleep 1
    status
}

function status(){
    pid=$(is_runnung)
    if [ "${pid}" == "0" ]
    then
        echo "Process $process_name is not running"
        return
    fi
    echo "Process is running"
    ps -fp $pid
}

function stop(){
    pid=$(is_runnung)
    if [ "${pid}" == "0" ]
    then
        echo "Process $process_name is not running. Nothing to stop"
        return
    fi
    kill -9 $pid
}



if [ "${action}" == "start" ]
then
    start
    exit 0
fi
if [ "${action}" == "status" ]
then
    status
    exit 0
fi
if [ "${action}" == "stop" ]
then
    stop
    exit 0
fi

echo "One of the option required [start, status, stop]"