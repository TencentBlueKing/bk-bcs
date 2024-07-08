FROM centos:7 AS build
RUN sed -i s/mirror.centos.org/vault.centos.org/g /etc/yum.repos.d/*.repo && sed -i s/^#.*baseurl=http/baseurl=http/g /etc/yum.repos.d/*.repo && sed -i s/^mirrorlist=http/#mirrorlist=http/g /etc/yum.repos.d/*.repo

# install compilation tools
RUN yum install -y make gcc perl pcre-devel zlib-devel wget readline-devel openssl-devel unzip net-tools less

# install nginx with http_proxy_connect
RUN yum groupinstall -y 'Development Tools' && yum install -y libxslt-devel gd-devel gperftools-devel
RUN cd /home && rm -rf nginx-1.16.* ngx_http_proxy_connect_module \
&& wget http://nginx.org/download/nginx-1.16.1.tar.gz \
&& git clone https://github.com/chobits/ngx_http_proxy_connect_module \
&& tar -zxvf nginx-1.16.1.tar.gz
RUN  cd /home/nginx-1.16.1 \
&& patch -p1 < /home/ngx_http_proxy_connect_module/patch/proxy_connect_rewrite_101504.patch
RUN cd /home/nginx-1.16.1 && ./configure \
--prefix=/usr/local/nginx \
--add-module=/home/ngx_http_proxy_connect_module \
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
--with-file-aio \
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
--with-ld-opt='-Wl,-z,relro -specs=/usr/lib/rpm/redhat/redhat-hardened-ld -Wl,-E' 
RUN cd /home/nginx-1.16.1 && make && make install

# STEP 2: build controller image
FROM centos:7
RUN sed -i s/mirror.centos.org/vault.centos.org/g /etc/yum.repos.d/*.repo && sed -i s/^#.*baseurl=http/baseurl=http/g /etc/yum.repos.d/*.repo && sed -i s/^mirrorlist=http/#mirrorlist=http/g /etc/yum.repos.d/*.repo

# copy nginx file
COPY --from=build /usr/local/nginx /usr/local/nginx
RUN mkdir -p /usr/local/nginx/tmp && mkdir -p /var/log/nginx

#for command envsubst
RUN yum install -y gettext

# setup required ENV
ENV BCS_EGRESS_PORT=12345

# prepare egress-controller requirements
RUN mkdir -p /data/bcs/bcs-egress-controller/template
COPY config/nginx-template.conf /data/bcs/bcs-egress-controller/template
COPY config/nginx.conf /usr/local/nginx/conf/
COPY container-start.sh /data/bcs/bcs-egress-controller/
RUN chmod +x /data/bcs/bcs-egress-controller/container-start.sh 
WORKDIR /data/bcs/bcs-egress-controller
CMD [ "/data/bcs/bcs-egress-controller/container-start.sh " ]
