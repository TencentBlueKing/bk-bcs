#!/bin/bash

# Open ipvs
modprobe -r ip_vs
modprobe -r ip_vs_rr
modprobe -r ip_vs_wrr
modprobe -r ip_vs_sh