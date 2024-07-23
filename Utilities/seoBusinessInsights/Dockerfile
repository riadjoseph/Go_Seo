FROM golang:1.22

LABEL authors="jason"

WORKDIR /app

COPY go.mod .

COPY go.sum .

RUN go mod download

COPY seoBusinessInsights/* ./

RUN go build seoBusinessInsights.go

EXPOSE 8080

ENV envInsightsFolder="./seoInsightsCache"
ENV envInsightsLogFolder="."

ENTRYPOINT ["./seoBusinessInsights"]

