#!/bin/bash -e

HELM_LINK=http://bkopen-1252002024.file.myqcloud.com/bcs/kubectl-helm.tar.gz

# 打包APP
function build_bcs_app() {
    APP_CODE="bk_bcs_app"
    BUILD_DIR=build/$APP_CODE
    SRC_DIR=build/$APP_CODE/src
    PKGS_DIR=build/$APP_CODE/pkgs
    BIN_DIR=$SRC_DIR/bin

    # 清理环境
    rm -rf $BUILD_DIR
    mkdir -p $SRC_DIR $PKGS_DIR $BIN_DIR

    # 下载静态二进制文件
    curl $HELM_LINK | tar xzf - -C $BIN_DIR && chmod a+x $BIN_DIR/*

    # 编译前端静态资源
    docker run --rm -it -v `pwd`/frontend:/data node:10.15.3-stretch bash -c "cd /data && npm install . && npm run build"

    # 收集后端，前端资源
    docker run --rm -it -e "DJANGO_SETTINGS_MODULE"="backend.settings.ce.dev" -v `pwd`:/data python:3.6.8-stretch bash -c "cd /data && pip install -r requirements.txt && python manage.py collectstatic --noinput && pip download -d $PKGS_DIR -r requirements.txt"

    # 同步源码和配置文件
    rsync -avz requirements.txt $SRC_DIR
    rsync -avz runtime.txt $SRC_DIR
    rsync -avz bk_bcs_app.png $SRC_DIR
    rsync -avz manage.py $SRC_DIR
    rsync -avz wsgi.py $SRC_DIR
    rsync -avz backend --exclude="*.pyc" $SRC_DIR
    rsync -avz staticfiles $SRC_DIR

    # 修改app.yml文件
    rsync -avz app.yml $BUILD_DIR
    echo "libraries:" >> $BUILD_DIR/app.yml
    grep -e "^[^#].*$" requirements.txt | awk '{split($1,b,"==");printf "- name: "b[1]"\n  version: \""b[2]"\"\n"}' >> $BUILD_DIR/app.yml

    # 打包, 从环境变量或者当前时间做版本号
    VERSION=${VERSION:-`date "+%Y%m%d%H%M%S"`}
    PKG_DATETIME=`date "+%Y-%m-%d %H:%M:%S"`

    # mac下sed命令参数不一样
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i "" "s/__VERSION__/${VERSION}/g" $BUILD_DIR/app.yml
        sed -i "" "s/__DATETIME__/${PKG_DATETIME}/g" $BUILD_DIR/app.yml
    else
        sed -i "s/__VERSION__/${VERSION}/g" $BUILD_DIR/app.yml
        sed -i "s/__DATETIME__/${PKG_DATETIME}/g" $BUILD_DIR/app.yml
    fi

    cd build
    tar -zcvf "$APP_CODE-V$VERSION.tar.gz" $APP_CODE
}

# 打包WebConsole
function build_web_console() {
    PROJECT="bcs_web_console"
    EDITION="ce"
    BUILD_DIR=build/bcs/web_console

    # 清理环境
    rm -rf build/bcs
    mkdir -p $BUILD_DIR

    # 同步源码和配置文件
    rsync -avz requirements.txt $BUILD_DIR
    rsync -avz backend --exclude="*.pyc" $BUILD_DIR

    # 打包
    VERSION=${VERSION:-`date "+%Y%m%d%H%M%S"`}
    cd build
    tar -zcvf "$PROJECT-$EDITION-$VERSION.tar.gz" bcs
}


#############
# Main Loop #
#############
case $1 in
    build_bcs_app)
        build_bcs_app
        ;;
    build_web_console)
        build_web_console
        ;;
    *)
        echo "usage $0 {build_bcs_app|build_web_console}"
        exit 1
        ;;
esac