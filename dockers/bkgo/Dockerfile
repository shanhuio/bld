FROM cr.shanhu.io/lib/golang
LABEL author="Shanhu Tech Inc."

RUN apk add --update bash
RUN adduser -D -u 1000 forge
RUN chown -R forge /go

USER 1000
