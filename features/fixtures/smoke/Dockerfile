FROM busybox:latest
MAINTAINER docker-rebase FTW!

# Remove base-supplied files
RUN rm -rf /etc/init.d/*

# Add some files and directories
RUN mkdir /foo
RUN touch /foobar /baz /home.is.not.here
RUN cp -a /etc/network /

# Remove directory that is also prefix of a file
RUN rm -rf /etc/init.d/* /network/if-* /home /foo

USER nobody

