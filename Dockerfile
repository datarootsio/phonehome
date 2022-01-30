FROM golang:1.16-alpine
COPY . /app
WORKDIR /app
RUN CGO_ENABLED=0 go build -o phonehome .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=0 /app/phonehome ./
COPY --from=0 /app/settings.yaml ./
CMD ["./phonehome"]  