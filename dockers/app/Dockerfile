FROM cr.shanhu.io/lib/base
MAINTAINER Shanhu Tech Inc.

RUN adduser -D -u 3000 app
RUN mkdir -p /opt/app
RUN cd /opt/app/ && mkdir bin var etc lib tmp
RUN chown -R app /opt/app
WORKDIR /opt/app
