FROM golang:1.14 as builder

ENV APP_USER app
ENV APP_HOME /go/src/deployer

RUN groupadd $APP_USER && useradd -m -g $APP_USER -l $APP_USER
RUN mkdir -p $APP_HOME && chown -R $APP_USER:$APP_USER $APP_HOME

WORKDIR $APP_HOME
USER $APP_USER

RUN go mod download
RUN go mod verify
RUN go build -o server

FROM debian:buster

ENV APP_USER app
ENV APP_HOME /go/src/deployer
ENV APP_PORT 5501

RUN groupadd $APP_USER && useradd -m -g $APP_USER -l $APP_USER
RUN mkdir -p $APP_HOME
WORKDIR $APP_HOME

COPY --chown=0:0 --from=builder $APP_HOME/deployer $APP_HOME

EXPOSE ${APP_PORT}
USER $APP_USER
CMD ["./server"]