FROM golang:1.16-alpine as build
COPY . /app
WORKDIR /app
RUN CGO_ENABLED=0 go build -o phonehome .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=build /app/phonehome ./
COPY docs/swagger.json /app/docs/swagger.json
CMD ["./phonehome"]