#!/bin/bash

module="bcs-cpuset-device"

cd /data/bcs/${module}
chmod +x ${module}

#set reserved cores via BCS_CPUSET_RESERVED_LAST_CORE_NUM
res_cores_env=`python reserve_cores.py | grep bcsCpuSetReservedCpuSetList`
if [[ -n $res_cores_env ]]; then
    $res_cores_env
    echo "get reserved ${bcsCpuSetReservedCpuSetList}"
fi

#check configuration render
if [ "x$BCS_CONFIG_TYPE" == "xrender" ]; then
  cat ${module}.json.template | envsubst | tee ${module}.json
fi

#ready to start
/data/bcs/${module}/${module} $@
