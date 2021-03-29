#!/usr/bin/env python
#-*- coding:utf8 -*-

import os
import sys

def get_cpu_core_dict():
    cpuinfos = open("/proc/cpuinfo").read()
    print cpuinfos
    core_dict = {}
    for seg in cpuinfos.split("\n\n"):
        coreid = -1
        processor_list = []
        for line in seg.split("\n"):
            if line.startswith("processor"):
                processorid = line.split(":")[1].strip()
                processor_list.append(int(processorid))
            elif line.startswith("core id"):
                tmp_core_id = line.split(":")[1].strip()
                coreid = int(tmp_core_id)
        if coreid != -1:
            if coreid in core_dict:
                core_dict[coreid] = core_dict[coreid] + processor_list
            else:
                core_dict[coreid] = processor_list
    return core_dict

if __name__ == "__main__":
    reserved_cores_str = os.environ.get("BCS_CPUSET_RESERVED_LAST_CORE_NUM", "")
    if reserved_cores_str == "":
        print "BCS_CPUSET_RESERVED_LAST_CORE_NUM is empty, do not reserve cpu cores"
        sys.exit(0)
    reserved_cores = int(reserved_cores_str)
    core_dict = get_cpu_core_dict()
    key_list = core_dict.keys()
    key_list.sort(reverse=True)
    counter = 0
    reserved_logical_core_list = []
    for key in key_list:
        if counter < reserved_cores:
            reserved_logical_core_list = reserved_logical_core_list + core_dict[key]
            counter = counter + 1
        else:
            break
    reserved_logical_core_list.sort()
    print "reserved logical core list: ", reserved_logical_core_list
    print "export bcsCpuSetReservedCpuSetList=%s" % ",".join(str(x) for x in reserved_logical_core_list)
