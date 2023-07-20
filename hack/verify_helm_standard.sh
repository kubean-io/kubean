#!/usr/bin/env bash

# Copyright 2023 Authors of kubean-io
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail
#############################
#
#
# 根据 helm规范 https://dwiki.daocloud.io/x/R0WOBw
# 进行CI 自动化检查
#
# Usaage:
# bash verify_helm_standard.sh ghippo-0.9.28.tgz
#
############################


##########################
# Case列表
#
##########################
function helmCheck::cases(){
   local array=("noLatest" "noInternalImage" "noLargeRequest" "withReadme")
   echo ${array[*]}
}

#########################
# Case
########################
function helmCheck::noLatest(){
  local helmContext=$1
  local match=$(echo "${helmContext}" |grep "image:" |grep "latest")
  if [ "$match" == "" ];then
     echo true
  fi
  echo false
}

#########################
# Case
########################
function helmCheck::noInternalImage(){
  local helmContext=$1
  local match=$(echo ${helmContext} |grep "image:" |grep "release-ci.daocloud.io")
  if [ "$match" == "" ];then
     echo true
  fi
  echo false
}


#########################
# Case
########################
function helmCheck::noLargeRequest(){
  local helmContext=$1
  local cpuThreshold="500" #m
  local memThreshold="500" #$Mi
  local pass=true
  local cpus=$(echo "$helmContext" | yq '.. | select(has("requests"))| .requests.cpu'       | grep -v "\-\-\-" |grep -v null |sed "s/m//g")
  local memories=$(echo "$helmContext" | yq '.. | select(has("requests"))| .requests.memory'| grep -v "\-\-\-"| grep -v null |sed "s/Mi//g")
  for cpu in ${cpus[*]};do
     if [ "$cpu" -gt $cpuThreshold ] ;then
        #echo "[Info] CPU request $cpu exceed threshold $cpuThreshold ! "
        pass=1
        break
     fi
  done
  for mem in ${memories[*]};do
     if [ "$mem" -gt $memThreshold ] ;then
        #echo "[Info] Memory request $mem exceed threshold $memThreshold ! "
        pass=1
        break
     fi
  done

echo $pass
}

#########################
# Case
########################
function helmCheck::withReadme(){
  local helmContext=$1
  local tgzFile=$2
  tmpFolder=$(mktemp -d)
  tar -zxf $tgzFile -C $tmpFolder
  readme=$(ls $tmpFolder/*/README.md)
  rm -rf $tmpFolder || ls;
  if [ "$readme" == "" ];then
    #no README found
    echo false
  fi
  echo true
}
######################################################
# Framework Below ---------------
####################################################

######################
# expand the helm to string
#######################
function helmCheck::render(){
  local tgzFile=$1
  local extra_param=""
  if  [[ "$tgzFile"  =~ .*"insight-agent".* ]];then
    extra_param="  --set kube-prometheus-stack.prometheus.prometheusSpec.externalLabels.cluster=\"\"
    --set opentelemetry-collector.enabled=true  \
    --set opentelemetry-operator.enabled=true"
  fi
  output=$(helm template  $extra_param $tgzFile)
  if [ "$?" -ne 0 ];then
    echo "[Error] Failed to helm template the $tgzFile! "
    exit 1
  fi
  echo "${output}"
}


##############################
#
# Arguments:
# 1. tgz file
#
############################
function helmCheck::runValidate(){
  local tgzFile=$1
  local context=$(helmCheck::render $tgzFile)
  local allpass="true"
  local pass=0
  local array=$(helmCheck::cases)
  for checkfun in ${array[*]}; do
     pass=$( helmCheck::$checkfun "$context" "$tgzFile" )
     if $pass; then
       echo "[Info] ✅  Case Successed!      ($checkfun).."
     else
       echo "[Error] ❌ Case Failed!  ($checkfun) .."
       allpass="false"
     fi
  done
  if [ "$allpass" == "false" ];then
    exit 0
  fi
}


#################
# Main
################
helmCheck::runValidate $1
