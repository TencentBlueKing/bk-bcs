/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package patchmount

import corev1 "k8s.io/api/core/v1"

const (
	pluginName          = "patchmount"
	pluginAnnotationKey = pluginName + ".webhook.bkbcs.tencent.com"

	patchMountLxcfs    = "lxcfs"
	patchMountCgroupfs = "cgroupfs"

	// 丢与/sys/devices 不隔离
	disableMountSysDevicesAnnotationKey = pluginAnnotationKey + "/disable-sys-devices"
)

var mountPropagation = corev1.MountPropagationHostToContainer
var file = corev1.HostPathFile
var dir = corev1.HostPathDirectory

var lxcfsVolumeMountsTemplate = []corev1.VolumeMount{
	{
		Name:             "lxcfs-upper-dir",
		MountPath:        "/var/lib/lxc",
		MountPropagation: &mountPropagation,
	},
	{
		Name:      "lxcfs-proc-cpuinfo",
		MountPath: "/proc/cpuinfo",
	},
	{
		Name:      "lxcfs-proc-meminfo",
		MountPath: "/proc/meminfo",
	},
	{
		Name:      "lxcfs-proc-diskstats",
		MountPath: "/proc/diskstats",
	},
	{
		Name:      "lxcfs-proc-stat",
		MountPath: "/proc/stat",
	},
	{
		Name:      "lxcfs-proc-swaps",
		MountPath: "/proc/swaps",
		ReadOnly:  true,
	},
	{
		Name:      "lxcfs-proc-uptime",
		MountPath: "/proc/uptime",
	},
	{
		Name:      "lxcfs-proc-loadavg",
		MountPath: "/proc/loadavg",
	},
}

var lxcfsVolumeMountsTemplateSysDevices = corev1.VolumeMount{
	Name:      "lxcfs-sys-devices-system-cpu-online",
	MountPath: "/sys/devices/system/cpu/online",
}

var lxcfsVolumesTemplate = []corev1.Volume{
	{
		Name: "lxcfs-upper-dir",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/var/lib/lxc",
				Type: &dir,
			},
		},
	},
	{
		Name: "lxcfs-proc-cpuinfo",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/var/lib/lxc/lxcfs/proc/cpuinfo",
				Type: &file,
			},
		},
	},
	{
		Name: "lxcfs-proc-diskstats",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/var/lib/lxc/lxcfs/proc/diskstats",
				Type: &file,
			},
		},
	},
	{
		Name: "lxcfs-proc-meminfo",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/var/lib/lxc/lxcfs/proc/meminfo",
				Type: &file,
			},
		},
	},
	{
		Name: "lxcfs-proc-stat",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/var/lib/lxc/lxcfs/proc/stat",
				Type: &file,
			},
		},
	},
	{
		Name: "lxcfs-proc-swaps",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/var/lib/lxc/lxcfs/proc/swaps",
				Type: &file,
			},
		},
	},
	{
		Name: "lxcfs-proc-uptime",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/var/lib/lxc/lxcfs/proc/uptime",
				Type: &file,
			},
		},
	},
	{
		Name: "lxcfs-proc-loadavg",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/var/lib/lxc/lxcfs/proc/loadavg",
				Type: &file,
			},
		},
	},
}

var lxcfsVolumesTemplateSysDevices = corev1.Volume{
	Name: "lxcfs-sys-devices-system-cpu-online",
	VolumeSource: corev1.VolumeSource{
		HostPath: &corev1.HostPathVolumeSource{
			Path: "/var/lib/lxc/lxcfs/sys/devices/system/cpu/online",
			Type: &file,
		},
	},
}

var cgroupfsVolumeMountsTemplate = []corev1.VolumeMount{
	{
		Name:      "cgroupfs-proc-cpuinfo",
		MountPath: "/proc/cpuinfo",
	},
	{
		Name:      "cgroupfs-proc-meminfo",
		MountPath: "/proc/meminfo",
	},
	{
		Name:      "cgroupfs-proc-diskstats",
		MountPath: "/proc/diskstats",
	},
	{
		Name:      "cgroupfs-proc-stat",
		MountPath: "/proc/stat",
	},
	{
		Name:      "cgroupfs-proc-uptime",
		MountPath: "/proc/uptime",
	},
	{
		Name:      "cgroupfs-proc-loadavg",
		MountPath: "/proc/loadavg",
	},
}

var cgroupfsVolumeMountsTemplateSysDevices = corev1.VolumeMount{
	Name:      "cgroupfs-sys-devices-system-cpu",
	MountPath: "/sys/devices/system/cpu",
}

var cgroupfsVolumesTemplate = []corev1.Volume{
	{
		Name: "cgroupfs-proc-cpuinfo",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/cgroupfs/proc/cpuinfo",
				Type: &file,
			},
		},
	},
	{
		Name: "cgroupfs-proc-diskstats",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/cgroupfs/proc/diskstats",
				Type: &file,
			},
		},
	},
	{
		Name: "cgroupfs-proc-meminfo",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/cgroupfs/proc/meminfo",
				Type: &file,
			},
		},
	},
	{
		Name: "cgroupfs-proc-stat",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/cgroupfs/proc/stat",
				Type: &file,
			},
		},
	},
	{
		Name: "cgroupfs-proc-uptime",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/cgroupfs/proc/uptime",
				Type: &file,
			},
		},
	},
	{
		Name: "cgroupfs-proc-loadavg",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/cgroupfs/proc/loadavg",
				Type: &file,
			},
		},
	},
}

var cgroupfsVolumesTemplateSysDevices = corev1.Volume{
	Name: "cgroupfs-sys-devices-system-cpu",
	VolumeSource: corev1.VolumeSource{
		HostPath: &corev1.HostPathVolumeSource{
			Path: "/cgroupfs/sys/devices/system/cpu",
			Type: &dir,
		},
	},
}
