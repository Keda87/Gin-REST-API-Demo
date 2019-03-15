FROM golang:1.11.5-alpine3.9

RUN mkdir /sourcecode
WORKDIR /sourcecode
COPY . /sourcecode/

RUN apk update && apk upgrade &&\
    apk add --no-cache build-base\
                       bash\
                       git\
                       openssh

RUN go get github.com/gin-gonic/gin &&\
    go get github.com/jmoiron/sqlx &&\
    go get github.com/mattn/go-sqlite3 &&\
    go build -o main .