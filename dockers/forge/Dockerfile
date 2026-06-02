FROM cr.shanhu.io/lib/golang
MAINTAINER Shanhu Tech Inc.

RUN mkdir /usr/local/idle
COPY idle.go /usr/local/idle/idle.go
RUN cd /usr/local/idle && /usr/local/go/bin/go build idle.go
RUN rm /usr/local/idle/idle.go

CMD ["/usr/local/idle/idle"]
