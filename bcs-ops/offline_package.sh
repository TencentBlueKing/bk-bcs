#!/bin/bash

set -euo pipefail

CACHE_DIR=${CACHE_DIR:-"/tmp/bcs-ops-offline"}
VERSION=
CACHE_DIR_BIN="${CACHE_DIR}/bin-tools"
CACHE_DIR_IMG="${CACHE_DIR}/images"
CACHE_DIR_CHART="${CACHE_DIR}/charts"
CACHE_DIR_RPM="${CACHE_DIR}/rpm"

USER=$(base64 -d <<<"$USER")
TOKEN=$(base64 -d <<<"$TOKEN")

upload_mirrors() {
  local path filename url
  path=$1
  filename=$2
  url="https://mirrors.tencent.com/repository/generic\
/bcs-ops/bcs-ops-offline/${path}/"
  local curl_cmd=(curl --request PUT -u "${USER}:${TOKEN}"
    --url "${url}" --upload-file "${filename}")
  echo "${curl_cmd[@]}"
  if ! "${curl_cmd[@]}"; then
    echo "[FATAL]: fail upload ${filename} to ${url}, Please check permission"
    return 1
  fi
  return 0
}

safe_curl() {
  local url save_file
  url=$1
  save_file=$2

  if [[ -f $save_file ]]; then
    echo "[INFO]: $save_file exist"
  else
    echo "[INFO]: downloading ${url}"
    if ! curl -sSfL "${url}" -o "${save_file}" -m "360"; then
      echo "[FATAL]: Fail to download ${url}"
      rm -f "${save_file}"
      return 1
    fi
  fi
  return 0
}

download_k8s() {
  local version name url tar_name
  name=k8s
  version="$1"
  [[ -n ${version} ]] || echo "$name missing version"
  tar_name="${CACHE_DIR_BIN}/${name}-${version}.tgz"
  if [[ -f "$tar_name" ]]; then
    echo "[INFO]: $tar_name exists, skip download"
    cp -v "$tar_name" "${CACHE_DIR}/version-${VERSION}/bin-tools/"
    return 0
  fi

  url="https://dl.k8s.io/v${version}/bin/linux/${arch}"

  cache_dir="${CACHE_DIR_BIN}/${name}-${version}"
  mkdir -pv "${cache_dir}/bin" "${cache_dir}/systemd"

  safe_curl "${url}/kubeadm" "${cache_dir}/bin/kubeadm" || exit 1
  safe_curl "${url}/kubectl" "${cache_dir}/bin/kubectl" || exit 1
  safe_curl "${url}/kubelet" "${cache_dir}/bin/kubelet" || exit 1

  url="https://raw.githubusercontent.com/kubernetes/release/master/\
cmd/krel/templates/latest/kubelet/kubelet.service"
  safe_curl "${url}" "${cache_dir}/systemd/kubelet.service" || exit 1

  url="https://raw.githubusercontent.com/kubernetes/release/master/\
cmd/krel/templates/latest/kubeadm/10-kubeadm.conf"
  safe_curl "${url}" "${cache_dir}/systemd/10-kubeadm.conf" || exit 1

  chmod 111 -R "${cache_dir}/bin"
  chmod 666 -R "${cache_dir}/systemd"
  tar cvzf "$tar_name" -C "${cache_dir}" bin/ systemd/
  cp -v "$tar_name" "${CACHE_DIR}/version-${VERSION}/bin-tools/"
}

download_cni-plugins() {
  local version name url tar_name
  name="cni-plugins"
  version="$1"
  [[ -n ${version} ]] || echo "$name missing version"
  tar_name="${CACHE_DIR_BIN}/${name}-${version}.tgz"
  if [[ -f "$tar_name" ]]; then
    echo "[INFO]: $tar_name exists, skip download"
    cp -v "$tar_name" "${CACHE_DIR}/version-${VERSION}/bin-tools/"
    return 0
  fi

  url="https://github.com/containernetworking/plugins/releases/download\
/v${version}/cni-plugins-linux-${arch}-v${version}.tgz"

  cache_dir="${CACHE_DIR_BIN}/${name}-${version}"
  mkdir -pv "${cache_dir}/bin"

  safe_curl "${url}" "${cache_dir}/cni-plugins.tgz" || exit 1

  if ! tar xfvz "${cache_dir}/cni-plugins.tgz" -C "${cache_dir}/bin"; then
    echo "[FATAL]: ${cache_dir}/cni-plugins.tgz 解压失败，清理相关的缓存文件"
    rm -rf "$cache_dir"
    exit 1
  fi

  chmod 111 -R "${cache_dir}/bin"
  tar cvzf "$tar_name" -C "${cache_dir}" bin/
  cp -v "$tar_name" "${CACHE_DIR}/version-${VERSION}/bin-tools/"
}

download_crictl() {
  local version name url tar_name
  name="crictl"
  version="$1"
  [[ -n ${version} ]] || echo "$name missing version"
  tar_name="${CACHE_DIR_BIN}/${name}-${version}.tgz"
  if [[ -f "$tar_name" ]]; then
    echo "[INFO]: $tar_name exists, skip download"
    cp -v "$tar_name" "${CACHE_DIR}/version-${VERSION}/bin-tools/"
    return 0
  fi

  url="https://github.com/kubernetes-sigs/cri-tools/releases/download\
/v${version}/crictl-v${version}-linux-${arch}.tar.gz"

  cache_dir="${CACHE_DIR_BIN}/${name}-${version}"
  mkdir -pv "${cache_dir}/bin"

  safe_curl "${url}" "${cache_dir}/crictl.tar.gz"
  if ! tar xfvz "${cache_dir}/crictl.tar.gz" -C "${cache_dir}/bin"; then
    echo "[FATAL]: ${cache_dir}/crictl.tar.gz 解压失败，清理相关的缓存文件"
    rm -rf "$cache_dir"
    exit 1
  fi

  chmod 111 -R "${cache_dir}/bin"
  tar cvzf "$tar_name" -C "${cache_dir}" bin/
  cp -v "$tar_name" "${CACHE_DIR}/version-${VERSION}/bin-tools/"
}

download_docker() {
  local version name url tar_name
  name="docker"
  version="$1"
  [[ -n ${version} ]] || echo "$name missing version"
  tar_name="${CACHE_DIR_BIN}/${name}-${version}.tgz"
  if [[ -f "$tar_name" ]]; then
    echo "[INFO]: $tar_name exists, skip download"
    cp -v "$tar_name" "${CACHE_DIR}/version-${VERSION}/bin-tools/"
    return 0
  fi

  if [[ ${arch} == "amd64" ]];then
    url="https://download.docker.com/linux/static/stable/x86_64/docker-${version}.tgz"
  elif [[ ${arch} == "arm64" ]];then
    url="https://download.docker.com/linux/static/stable/aarch64/docker-${version}.tgz"
  else
    echo "[FATAL]: unknown arch ${arch}"
    exit 1
  fi

  cache_dir="${CACHE_DIR_BIN}/${name}-${version}"
  mkdir -pv "${cache_dir}/bin"

  safe_curl "${url}" "${cache_dir}/docker.tgz" || exit 1
  if ! tar xfvz "${cache_dir}/docker.tgz" -C "${cache_dir}/bin" --strip-components=1 docker/; then
    echo "[FATAL]: ${cache_dir}/docker.tgz 解压失败，清理相关的缓存文件"
    rm -rf "$cache_dir"
    exit 1
  fi

  mkdir -pv "${cache_dir}/systemd"
  systemd_ver="${version%.*}"
  url="https://raw.githubusercontent.com/moby/moby/\
${systemd_ver}/contrib/init/systemd/docker.socket"
  safe_curl "${url}" "${cache_dir}/systemd/docker.socket"
  url="https://raw.githubusercontent.com/moby/moby/\
${systemd_ver}/contrib/init/systemd/docker.service"
  safe_curl "${url}" "${cache_dir}/systemd/docker.service"

  chmod 111 -R "${cache_dir}/bin"
  chmod 666 -R "${cache_dir}/systemd"
  tar cvzf "$tar_name" -C "${cache_dir}" bin/ systemd/
  cp -v "$tar_name" "${CACHE_DIR}/version-${VERSION}/bin-tools/"
}

download_containerd() {
  local version name url tar_name
  name="containerd"
  version="$1"
  [[ -n "${version}" ]] || echo "$name missing version"
  tar_name="${CACHE_DIR_BIN}/${name}-${version}.tgz"
  if [[ -f "$tar_name" ]]; then
    echo "[INFO]: $tar_name exists, skip download"
    cp -v "$tar_name" "${CACHE_DIR}/version-${VERSION}/bin-tools/"
    return 0
  fi

  url="https://github.com/containerd/containerd/releases/download/\
v${version}/containerd-${version}-linux-${arch}.tar.gz"

  cache_dir="${CACHE_DIR_BIN}/${name}-${version}"

  mkdir -pv "${cache_dir}/bin"
  safe_curl "${url}" "${cache_dir}/containerd.tar.gz"
  if ! tar xfvz "${cache_dir}/containerd.tar.gz" -C "${cache_dir}/bin" --strip-components=1 bin/; then
    echo "[FATAL]: ${cache_dir}/containerd.tar.gz 解压失败，清理相关的缓存文件"
    rm -rf "$cache_dir"
    exit 1
  fi

  mkdir -pv "${cache_dir}/systemd"
  url="https://raw.githubusercontent.com/containerd/containerd/v1.6.20/containerd.service"
  url="https://raw.githubusercontent.com/containerd/containerd\
/v${version}/containerd.service"
  safe_curl "${url}" "${cache_dir}/systemd/containerd.service"

  chmod 111 -R "${cache_dir}/bin"
  chmod 666 -R "${cache_dir}/systemd"
  tar cvzf "${CACHE_DIR_BIN}/${name}-${version}.tgz" \
    -C "${cache_dir}" bin/ systemd/
  cp -v "$tar_name" "${CACHE_DIR}/version-${VERSION}/bin-tools/"
}

download_runc() {
  local version name url tar_name
  name="runc"
  version="$1"
  [[ -n ${version} ]] || echo "$name missing version"
  tar_name="${CACHE_DIR_BIN}/${name}-${version}.tgz"
  if [[ -f "$tar_name" ]]; then
    echo "[INFO]: $tar_name exists, skip download"
    cp -v "$tar_name" "${CACHE_DIR}/version-${VERSION}/bin-tools/"
    return 0
  fi

  url="https://github.com/opencontainers/runc/releases/\
download/v${version}/runc.${arch}"

  cache_dir="${CACHE_DIR_BIN}/${name}-${version}"
  mkdir -pv "${cache_dir}/bin"

  safe_curl "$url" "${cache_dir}/bin/runc"

  chmod 111 -R "${cache_dir}/bin"
  tar cvzf "${tar_name}" -C "${cache_dir}" bin/
  cp -v "$tar_name" "${CACHE_DIR}/version-${VERSION}/bin-tools/"
}

download_yq() {
  local version name url tar_name
  version="$1"
  name="yq"
  [[ -n ${version} ]] || echo "$name missing version"
  tar_name="${CACHE_DIR_BIN}/${name}-${version}.xz"

  url="https://github.com/mikefarah/yq/releases/download/v${version}/yq_linux_${arch}.tar.gz"
  safe_curl "$url" "yq_linux_${arch}tar.gz"
  tar -xf yq_linux_${arch}tar.gz -O | xz > ${tar_name}
  cp -v "$tar_name" "${CACHE_DIR}/version-${VERSION}/bin-tools/"
}

download_jq() {
  local version name url tar_name
  version="$1"
  name="jq"
  [[ -n ${version} ]] || echo "$name missing version"
  tar_name="${CACHE_DIR_BIN}/${name}-${version}.xz"

  url="https://github.com/jqlang/jq/releases/download/jq-${version}/jq-linux-${arch}"
  safe_curl "$url" "jq-linux-${arch}"
  tar -cJf "${tar_name}" "jq-linux-${arch}"
  cp -v "$tar_name" "${CACHE_DIR}/version-${VERSION}/bin-tools/"
}

download_rpm() {
  local rpm_name url rpm_file rpm
  IFS=' ' read -ra rpm <<<"$@"
  mkdir -pv "${CACHE_DIR_RPM}"
  for rpm_name in "${rpm[@]}"; do
    rpm_file="${CACHE_DIR_RPM}/${rpm_name}"
    if [[ -f $rpm_file ]]; then
      echo "[INFO]:${rpm_file} exist, skip download"
      cp -v "$rpm_file" "${CACHE_DIR}/version-${VERSION}/rpm/"
      break
    fi

    url="https://mirrors.tencent.com/repository/generic/\
bcs-ops/bcs-ops-offline/rpm/${rpm_name}"
    safe_curl "$url" "$rpm_file" || exit 1
  done <<<"$1"
}

download_charts() {
  local charts chart_name url chart_file
  IFS=' ' read -ra charts <<<"$@"
  mkdir -pv "${CACHE_DIR_CHART}"
  for chart_name in "${charts[@]}"; do
    chart_file="${CACHE_DIR_CHART}/${chart_name}"
    if [[ -f $chart_file ]]; then
      echo "[INFO]:${chart_file} exist, skip download"
      cp -v "$chart_file" "${CACHE_DIR}/version-${VERSION}/charts/"
      continue
    fi

    url="https://mirrors.tencent.com/repository/generic\
/bcs-ops/bcs-ops-offline/charts/${chart_name}"
    safe_curl "$url" "$chart_file" || exit 1
  done
}

download_img() {
  local imgs img img_name img_tag img_tar

  repo=$1
  shift
  IFS=' ' read -ra imgs <<<"$@"
  mkdir -pv "${CACHE_DIR_IMG}"
  for img in "${imgs[@]}"; do
    if [[ "${repo}" != "null" ]];then
      rel_img=${img//hub.bktencent.com/$repo}
    else
      rel_img=${img}
    fi
    img_name=${img##*/}
    img_tag=${img_name##*:}
    img_name=${img_name%%:*}
    img_tar="${CACHE_DIR_IMG}/${img_name}-${img_tag}.tar"
#    if [[ -f "${img_tar}" ]]; then
#      echo "[INFO]:${img} exist, skip pull"
#      cp -v "$img_tar" "${CACHE_DIR}/version-${VERSION}/images/"
#      continue
#    fi
    echo "[INFO]: trying to docker pull --platform ${arch} ${rel_img}"
    if docker manifest inspect "${rel_img}" >/dev/null; then
#    if skopeo inspect --raw "docker://${img}" >/dev/null; then
      if docker pull --platform linux/${arch} ${rel_img} >/dev/null;then
        echo docker tag ${rel_img} ${img}
        docker tag ${rel_img} ${img} >/dev/null
        echo docker save ${img} -o ${img_tar}
        docker save ${img} -o ${img_tar} >/dev/null
#      if skopeo copy --arch ${arch} "docker://${rel_img}" "docker-archive:${img_tar}:${img}" >/dev/null; then
        mv -v "$img_tar" "${CACHE_DIR}/version-${VERSION}/images/"
      else
        echo "[FATAL]: fail to pull ${img}"
        rm -rf "$img_tar"
        exit 1
      fi
    else
      echo "[FATAL]: can't find ${img} in registry!"
      exit 1
    fi
  done
}

unMarshall_mainfest() {
  local manifest_file ver_num version projects images charts rpms
  manifest_file=$1
  repo=$2

  ver_num=$(yq e '.bcs-ops | length' "$manifest_file")
  local i=0
  while ((i < ver_num)); do
    IFS=',' read -ra projects <<<"$(yq -o csv e '.bcs-ops[0] | keys' "$manifest_file")"
    VERSION=$(yq e ".bcs-ops[$i].version" "$manifest_file")
    for project in "${projects[@]}"; do
      case $project in
        "version")
          echo "version: $VERSION"
          ;;
        "bin-tools")
          mkdir -pv "${CACHE_DIR}/version-${VERSION}/bin-tools"
          yq e ".bcs-ops[$i].bin-tools" "$manifest_file" | while IFS=': ' read -r p v; do "download_$p" "${v//\"/}"; done
          ;;
        "images")
          mkdir -pv "${CACHE_DIR}/version-${VERSION}/images"
          IFS=',' read -ra images <<<"$(yq -o csv e ".bcs-ops[$i].images" "$manifest_file")"

          download_img "${repo}" "${images[@]}"
          ;;
        "charts")
          mkdir -pv "${CACHE_DIR}/version-${VERSION}/charts"
          IFS=',' read -ra charts <<<"$(yq -o csv e ".bcs-ops[$i].charts" "$manifest_file")"
          download_charts "${charts[@]}"
          ;;
        "rpm")
          mkdir -pv "${CACHE_DIR}/version-${VERSION}/rpm"
          IFS=',' read -ra rpms <<<"$(yq -o csv e ".bcs-ops[$i].rpm" "$manifest_file")"
          download_rpm "${rpms[@]}"
          ;;
        *)
          echo "unknow key $project"
          ;;
      esac
    done
    tar cvzf "${CACHE_DIR}/bcs-ops-offline-${VERSION}-${arch}.tgz" -C "${CACHE_DIR}" "version-${VERSION}/"
    upload_mirrors "version" "${CACHE_DIR}/bcs-ops-offline-${VERSION}-${arch}.tgz"
    ((i += 1))
  done
}

if [[ $# -eq 1 ]] || [[ -z "$2" ]]; then
  repo="null"
else
  repo=$2
fi

export arch=arm64
unMarshall_mainfest "$1" "$repo"
export arch=amd64
unMarshall_mainfest "$1" "$repo"