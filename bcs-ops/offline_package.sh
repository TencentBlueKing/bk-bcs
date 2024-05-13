#!/bin/bash

set -euo pipefail

CACHE_DIR=${CACHE_DIR:-"./bcs-ops-offline"}
VERSION=
CACHE_DIR_BIN="${CACHE_DIR}/bin-tools"
CACHE_DIR_IMG="${CACHE_DIR}/images"

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

  url="https://dl.k8s.io/v${version}/bin/linux/amd64"

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
/v${version}/cni-plugins-linux-amd64-v${version}.tgz"

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
/v${version}/crictl-v${version}-linux-amd64.tar.gz"

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

  url="https://download.docker.com/linux/static/stable/\
x86_64/docker-${version}.tgz"

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
v${version}/containerd-${version}-linux-amd64.tar.gz"

  cache_dir="${CACHE_DIR_BIN}/${name}-${version}"

  mkdir -pv "${cache_dir}/bin"
  safe_curl "${url}" "${cache_dir}/containerd.tar.gz"
  if ! tar xfvz "${cache_dir}/containerd.tar.gz" -C "${cache_dir}/bin" --strip-components=1 bin/; then
    echo "[FATAL]: ${cache_dir}/containerd.tar.gz 解压失败，清理相关的缓存文件"
    rm -rf "$cache_dir"
    exit 1
  fi

  mkdir -pv "${cache_dir}/systemd"
  url="https://raw.githubusercontent.com/containerd/containerd/v${version}/containerd.service"
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
download/v${version}/runc.amd64"

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

  url="https://bkopen-1252002024.file.myqcloud.com/ce7/tools/yq-${version}.xz"
  safe_curl "$url" "${tar_name}"
  cp -v "$tar_name" "${CACHE_DIR}/version-${VERSION}/bin-tools/"
}

download_jq() {
  local version name url tar_name
  version="$1"
  name="jq"
  [[ -n ${version} ]] || echo "$name missing version"
  tar_name="${CACHE_DIR_BIN}/${name}-${version}.xz"

  url="https://bkopen-1252002024.file.myqcloud.com/ce7/tools/jq-${version}.xz"
  safe_curl "$url" "${tar_name}"
  cp -v "$tar_name" "${CACHE_DIR}/version-${VERSION}/bin-tools/"
}

download_img() {
  local imgs img img_name img_tag img_tar
  IFS=' ' read -ra imgs <<<"$@"
  mkdir -pv "${CACHE_DIR_IMG}"
  for img in "${imgs[@]}"; do
    img_name=${img##*/}
    img_tag=${img_name##*:}
    img_name=${img_name%%:*}
    img_tar="${CACHE_DIR_IMG}/${img_name}-${img_tag}.tar"
    if [[ -f "${img_tar}" ]]; then
      echo "[INFO]:${img} exist, skip pull"
      cp -v "$img_tar" "${CACHE_DIR}/version-${VERSION}/images/"
      continue
    fi
    echo "[INFO]: trying to pull ${img}"
    if skopeo inspect --raw "docker://${img}" >/dev/null; then
      if skopeo copy "docker://${img}" "docker-archive:${img_tar}:${img}" >/dev/null; then
        cp -v "$img_tar" "${CACHE_DIR}/version-${VERSION}/images/"
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
  local manifest_file ver ver_num version projects images
  manifest_file=$1
  ver=$2

  ver_num=$(yq e '.bcs-ops | length' "$manifest_file")
  local i=0
  while ((i < ver_num)); do
    IFS=',' read -ra projects <<<"$(yq -o csv e '.bcs-ops[0] | keys' "$manifest_file")"
    VERSION=$(yq e ".bcs-ops[$i].version" "$manifest_file")
    if [[ "$VERSION" == "$ver" ]] || [[ -z $ver ]]; then
      for project in "${projects[@]}"; do
        case $project in
          "version")
            echo "version: $VERSION"
            ;;
          "bin-tools")
            mkdir -pv "${CACHE_DIR}/version-${VERSION}/bin-tools"
            yq e ".bcs-ops[$i].bin-tools" "$manifest_file" \
              | while IFS=': ' read -r p v; do "download_$p" "${v//\"/}"; done
            ;;
          "images")
            mkdir -pv "${CACHE_DIR}/version-${VERSION}/images"
            IFS=',' read -ra images <<<"$(yq -o csv e ".bcs-ops[$i].images" "$manifest_file")"
            download_img "${images[@]}"
            ;;
          *)
            echo "unknow key $project"
            ;;
        esac
      done
      tar cvzf "${CACHE_DIR}/bcs-ops-offline-${VERSION}.tgz" -C "${CACHE_DIR}" "version-${VERSION}/"
    else
      echo "skip pacakge $VERSION"
    fi
    ((i += 1))
  done
}

unMarshall_mainfest "$1" "${2:-}"
