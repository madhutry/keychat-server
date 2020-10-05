FROM golang:1.14.3-alpine AS build
WORKDIR /src
ENV CGO_ENABLED=0
COPY go.* ./
RUN go mod download
COPY . ./
#RUN go get -d -v ./...
RUN go build -o /out/chatserver .

FROM scratch AS bin
COPY --from=build /out/chatserver /

CMD ["./chatserver"]

