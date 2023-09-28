FROM docker.io/centos:7

RUN yum -y install rsync openssh-clients

WORKDIR /imp/

ENV TZ=Asia/Shanghai

ADD bin/ /imp/
ADD conf/ /imp/

CMD ["/imp/bin/imp-server","-c","/imp/conf/imp-server.conf"]