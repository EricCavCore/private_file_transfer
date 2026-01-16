FROM golang:1.25.5

LABEL org.opencontainers.image.authors="ecaverly@corenetwork.ca"

WORKDIR /app

COPY ./app ./

RUN go mod download

RUN go build -o sft

ENTRYPOINT [ "./sft" ]