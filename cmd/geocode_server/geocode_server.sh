#!/bin/sh

here=$(dirname $0)

action=$1
logdir=$2
process_name=geocode_server


function is_runnung(){
    RESULT=`pgrep -f ${process_name}$`
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
        return
    fi
    if [ -z "${logdir}" ]
    then
        logdir=$here
    fi
    echo "Starting process ${process_name}"
    echo "$here/geocode_server > $logdir/geocode_server.log 2>&1 &"

    nohup $here/geocode_server > $logdir/geocode_server.log 2>&1 &
}


function status(){
    pid=$(is_runnung)
    if [ "${pid}" == "0" ]
    then
        echo "Process $process_name is not running"
        return
    fi
    echo "Process $process_name is running"
    ps -fp $pid
}

function stop(){
    pid=$(is_runnung)
    if [ "${pid}" == "0" ]
    then
        echo "Process $process_name is not running"
        return
    fi
    kill -9 $pid
}



if [ "${action}" == "start" ]
then
    start
fi
if [ "${action}" == "status" ]
then
    status
fi
if [ "${action}" == "stop" ]
then
    stop
fi
