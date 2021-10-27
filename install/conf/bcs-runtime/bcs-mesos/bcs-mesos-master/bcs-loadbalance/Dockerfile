# STEP 1: build haproxy lua5.3 luarocks and lua libs
FROM centos:centos7 AS build

# install haproxy
RUN yum install -y make gcc perl pcre-devel zlib-devel wget readline-devel openssl-devel unzip net-tools less
RUN cd /home \
&& wget http://www.lua.org/ftp/lua-5.3.5.tar.gz \
&& tar -zxvf lua-5.3.5.tar.gz && rm lua-5.3.5.tar.gz \
&& cd lua-5.3.5 && make linux && make install
RUN cd /home \
&& wget https://luarocks.org/releases/luarocks-3.0.4.tar.gz \
&& tar zxpf luarocks-3.0.4.tar.gz && rm luarocks-3.0.4.tar.gz \
&& cd luarocks-3.0.4 \
&& ./configure && make bootstrap
RUN cd /home \
&& luarocks install penlight \
&& luarocks install luajson \
&& luarocks install luaidl \
&& luarocks install luadoc \
&& luarocks install vstruct \
&& luarocks install luasocket \
&& luarocks install loop \
&& luarocks install router \
&& luarocks install lua-cjson-ol
RUN cd /home \
&& wget http://www.haproxy.org/download/1.8/src/haproxy-1.8.19.tar.gz \
&& tar -xvzf haproxy-1.8.19.tar.gz \
&& rm haproxy-1.8.19.tar.gz \
&& cd haproxy-1.8.19/ \
&& make TARGET=linux2628 USE_LINUX_TPROXY=1 USE_ZLIB=1 USE_REGPARM=1 USE_OPENSSL=1 USE_LUA=1 USE_PCRE=1 USE_PCRE_JIT=1 \
&& make install

# install nginx
RUN yum groupinstall -y 'Development Tools' && yum install -y libxslt-devel gd gd-devel gperftools-devel
RUN cd /home \
&& wget http://nginx.org/download/nginx-1.15.9.tar.gz \
&& tar -zxvf nginx-1.15.9.tar.gz && rm nginx-1.15.9.tar.gz && cd nginx-1.15.9 \
&& ./configure \
--prefix=/usr/local/nginx \
--http-client-body-temp-path=/usr/local/nginx/tmp/client_body \
--http-proxy-temp-path=/usr/local/nginx/tmp/proxy \
--http-fastcgi-temp-path=/usr/local/nginx/tmp/fastcgi \
--http-uwsgi-temp-path=/usr/local/nginx/tmp/uwsgi \
--http-scgi-temp-path=/usr/local/nginx/tmp/scgi \
--pid-path=/run/nginx.pid \
--lock-path=/run/lock/subsys/nginx \
--with-perl_modules_path=/usr/local/nginx/perl5 \
--user=nginx \
--group=nginx \
--with-stream \
--with-stream_ssl_module \
--with-file-aio --with-ipv6 \
--with-http_ssl_module \
--with-http_v2_module \
--with-http_realip_module \
--with-http_addition_module \
--with-http_xslt_module \
--with-http_image_filter_module \
--with-http_sub_module \
--with-http_dav_module \
--with-http_flv_module \
--with-http_mp4_module \
--with-http_gunzip_module \
--with-http_gzip_static_module \
--with-http_random_index_module \
--with-http_secure_link_module \
--with-http_degradation_module \
--with-http_stub_status_module \
--with-pcre \
--with-pcre-jit \
--with-google_perftools_module \
--with-debug \
--with-cc-opt='-O2 -g -pipe -Wall -Wno-error -Wp,-D_FORTIFY_SOURCE=2 -fexceptions -fstack-protector-strong --param=ssp-buffer-size=4 -grecord-gcc-switches -specs=/usr/lib/rpm/redhat/redhat-hardened-cc1 -m64 -mtune=generic' \
--with-ld-opt='-Wl,-z,relro -specs=/usr/lib/rpm/redhat/redhat-hardened-ld -Wl,-E' \
&& make && make install

# copy bcs-statistic
COPY ./bcs-statistic /data/bcs/bcs-lb/bcs-statistic

RUN mkdir -p /data/bcs/bcs-lb/backup /data/bcs/bcs-lb/generate \
/data/bcs/bcs-lb/logs /data/bcs/bcs-lb/template /data/bcs/bcs-lb/cert
COPY bcs-loadbalance /data/bcs/bcs-lb
COPY start.sh /data/bcs/bcs-lb
COPY config/haproxy.cfg.template /data/bcs/bcs-lb/template
COPY config/nginx.conf.template /data/bcs/bcs-lb/template
COPY config/haproxy.cfg /etc/haproxy/
COPY config/nginx.conf /usr/local/nginx/conf/

# STEP 2: build loadbalance image
FROM centos:centos7
COPY --from=build /data/bcs/bcs-lb /data/bcs/bcs-lb
# copy haproxy file
COPY --from=build /etc/haproxy /etc/haproxy
COPY --from=build /usr/local/lib /usr/local/lib
COPY --from=build /usr/local/include /usr/local/include
COPY --from=build /usr/local/lib64 /usr/local/lib64
COPY --from=build /usr/bin /usr/bin
COPY --from=build /usr/lib64 /usr/lib64
COPY --from=build /usr/local/share /usr/local/share
COPY --from=build /usr/local/bin/lua /usr/local/bin/lua
COPY --from=build /usr/local/bin/luac /usr/local/bin/luac
COPY --from=build /usr/local/sbin/haproxy /usr/local/sbin/haproxy

# copy nginx file
COPY --from=build /usr/local/nginx /usr/local/nginx
RUN mkdir /usr/local/nginx/tmp && mkdir /var/log/nginx

# create user bcs;  install net tool
RUN useradd -u 1025 bcs
# set sesssion timeout time
ENV LB_SESSION_TIMEOUT 90
ENV BCS_PROXY_MODULE=haproxy

# chmod start.sh
RUN chmod +x /data/bcs/bcs-lb/start.sh /data/bcs/bcs-lb/bcs-loadbalance
WORKDIR /data/bcs/bcs-lb