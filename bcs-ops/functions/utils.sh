#!/bin/bash

#######################################
# Tencent is pleased to support the open source community by making Blueking Container Service available.
# Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
# Licensed under the MIT License (the "License"); you may not use this file except
# in compliance with the License. You may obtain a copy of the License at
# http://opensource.org/licenses/MIT
# Unless required by applicable law or agreed to in writing, software distributed under
# the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
# either express or implied. See the License for the specific language governing permissions and
# limitations under the License.
#######################################

# generic function, independent of business

#######################################
# colorful echo
# Arguments:
# $1 | color type, red/green/yellow/blue/purple
# Returns:
#   0
#######################################
utils::color_echo() {
  local color_code color
  color="${1,,}"
  shift
  if [[ "${HASTTY:-}" == 0 && $(tput colors) -gt 0 ]]; then
    echo "$*"
  else
    case "${color}" in
      "red")
        color_code="\033[031;1m"
        ;;
      "green")
        color_code="\033[032;1m"
        ;;
      "yellow")
        color_code="\033[033;1m"
        ;;
      "blue")
        color_code="\033[034;1m"
        ;;
      "purple")
        color_code="\033[035;1m"
        ;;
      "bwhite")
        color_code="\033[7m"
        ;;
      *)
        echo "missing color: ${color}"
        color_code=""
        ;;
    esac
    echo -e "${color_code}$*\033[0m"
  fi
  return 0
}

#######################################
# log function
# INFO|OK stream msg to /dev/stdout, OK means success info
# DEBUG|WARN|ERROR|FATAL stream msg to /dev/stderr
# Arguments:
# $1: log level, DEBUG|INFO|OK|WARN|ERROR|FATAL
# Globals:
#	$DEBUG | enable $DEBUG level
#	$LOG_FILE | write log to file
#   $ERR_CODE | fatal exit code, default is 1
# PrintColor:
# DEBUG-purple;INFO-blue;OK-green;WARN-yellow;ERROR-red;FATAL-red
# Returns:
# DEBUG-0;INFO-0;OK-0;WARN-0;ERROR-1;FATAL-$ERR_CODE and exit;
#######################################
utils::log() {
  # stream
  local level msg stream timestamp color format func_name
  level=${1^^}
  shift
  msg=$*
  timestamp="$(date +%Y/%m/%d-%H:%M:%S)"
  # latest call func
  if ((${#FUNCNAME[@]} <= 2)); then
    func_name="main"
  else
    if [[ -n ${TRAP_FLAG:-} ]]; then
      func_name="${FUNCNAME[2]}"
    else
      func_name="${FUNCNAME[1]}"
    fi
  fi
  if [[ -n ${TRAP_FLAG:-} ]]; then
    format="$timestamp [$level] ${BASH_SOURCE[2]}|${BASH_LINENO[1]}|${func_name}:"
  else
    format="$timestamp [$level] ${BASH_SOURCE[1]}|${BASH_LINENO[0]}|${func_name}:"
  fi
  case "${level}" in
    "DEBUG")
      if [[ -z ${DEBUG:-} ]]; then
        return 0
      fi
      stream="/dev/stderr"
      color="purple"
      ;;
    "INFO")
      stream="/dev/stdout"
      color="blue"
      ;;
    "OK")
      stream="/dev/stdout"
      color="green"
      ;;
    "WARN")
      stream="/dev/stderr"
      color="yellow"
      ;;
    "ERROR")
      stream="/dev/stderr"
      color="red"
      ;;
    "FATAL")
      local err_code
      err_code=${ERR_CODE:-1}
      format="${format}E$err_code"
      stream="/dev/stderr"
      color="red"
      ;;
    *)
      # unkown level, direct echo to /dev/stderr, and return 1
      utils::color_echo "purple" "unkown log level: $level"
      echo "$format $msg" >&2
      return 1
      ;;
  esac

  echo "$(utils::color_echo "$color" "$format")" "$msg" >"$stream"

  if [[ -n ${LOG_FILE:-} ]]; then
    echo "$format $msg" >>"$LOG_FILE"
  fi

  case $level in
    "DEBUG|INFO|OK|WARN")
      return 0
      ;;
    "ERROR")
      return 1
      ;;
    "FATAL")
      exit "$err_code"
      ;;
  esac
}

# ######################################
# trap ERR. When meet unpredictable impact error, force fatal and exit immediately.
# usage:
# add `trap "utils::on_ERR;" ERR` in your shell script or function inside.
# ######################################
utils::on_ERR() {
  ERR_CODE=$?
  export TRAP_FLAG=1
  utils::log FATAL "$(utils::color_echo bwhite "$BASH_COMMAND")"
}

#######################################
# check if operate_project function exists
# Arguments:
# $1: Operate action
# $2: Project name
# Return:
# "$1_$2" function exists return 0, else exit 1
# function
#######################################

utils::check_op() {
  local op_type="$1"
  local project="$2"
  if [[ -n ${op_type} ]] && [[ -n ${project} ]]; then
    type "${op_type}_${project}" &>/dev/null || utils::log FATAL "${op_type} [$project] NOT SUPPORT"
  else
    utils::log "FATAL" "missing op_type: ${op_type} or project: [$project] "
  fi
  return 0
}

#######################################
# check if operate_project function exists
# Arguments:
# $1: Operate action
# $2: Project name
# Return:
# "$1_$2" function exists return 0, else exit 1
# function
#######################################

utils::get_arch() {
  ARCH=$(uname -m)
  if [[ "${ARCH}" == "x86_64" ]];then
    ARCH="amd64"
  elif [[ "${ARCH}" == "aarch64" ]];then
    ARCH="arm64"
  else
    utils::log "FATAL" "unknown arch ${ARCH}"
  fi
}
