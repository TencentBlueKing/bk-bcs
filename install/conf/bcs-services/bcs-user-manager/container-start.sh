#!/bin/bash

module="bcs-user-manager"

cd /data/bcs/${module}
chmod +x ${module}

#check configuration render
if [ "x$BCS_CONFIG_TYPE" == "xrender" ]; then
  cat ${module}.json.template | \
# set default value
bcsTokenNotifyDryRun=${bcsTokenNotifyRtxTitle:-false} \
bcsTokenNotifyCron="${bcsTokenNotifyCron:-0 10 * * *}" \
bcsTokenNotifyTitle="${bcsTokenNotifyTitle:-TKEx(蓝鲸容器平台) API 密钥续期提醒}" \
bcsTokenNotifyContent="${bcsTokenNotifyContent:-你好，{{ .Username \}\}:<br>您的 API 密钥过期时间为: {{ .ExpiredAt \}\}，如有需要请前往 API 密钥页面及时续期。}" \
bcsTokenNotifyESBEmailPath="${bcsTokenNotifyESBEmailPath:-/component/compapi/tof/send_mail/}" \
bcsTokenNotifyESBRtxPath="${bcsTokenNotifyESBRtxPath:-/component/compapi/tof/send_rtx/}" \
envsubst | tee ${module}.json
fi

#ready to start
exec /data/bcs/${module}/${module} $@
