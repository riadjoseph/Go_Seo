FROM golang:1.22

LABEL authors="jason"

WORKDIR /app

COPY go.mod .

COPY go.sum .

RUN go mod download

COPY segmentifyLite/* ./

RUN go build segmentifyLite.go

EXPOSE 8081

ENV envSegmentifyLiteFolder="./segmentifyLiteCache"
ENV envSegmentifyLiteLogFolder="."

ENTRYPOINT ["./segmentifyLite"]

