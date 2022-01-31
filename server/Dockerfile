FROM golang:1.16-alpine as build
COPY /server /app
WORKDIR /app
RUN CGO_ENABLED=0 go build -o phonehome .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=build /app/phonehome ./
CMD ["./phonehome"]  