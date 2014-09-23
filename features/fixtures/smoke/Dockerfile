FROM busybox:latest
MAINTAINER docker-rebase FTW!

ADD Dockerfile /df
ADD hello.sh /hello

RUN cp -a /etc/network /

RUN rm -rf /etc/init.d/* /network/if-* /home /df

USER nobody
ENTRYPOINT [ "/hello" ]
