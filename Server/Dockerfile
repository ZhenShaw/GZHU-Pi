FROM golang:alpine

WORKDIR /home
ADD . /home
RUN go build --mod=vendor main.go

FROM zhenshaw/frp
#FROM alpine

RUN apk update \
    && apk add tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

COPY --from=0 /home/main /home
COPY --from=0 /home/config.toml /home
COPY --from=0 /home/queries /home/queries
WORKDIR /home

CMD ["/home/main"]