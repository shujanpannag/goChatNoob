FROM golang:1.16-alpine3.14 as gobuilder

RUN apk update && apk upgrade && apk add --no-cache make

WORKDIR /app
COPY . .

RUN make vendor
RUN make build

FROM alpine
RUN apk update && apk upgrade

COPY --from=gobuilder /app/goChat  /goChat

EXPOSE 6565

CMD [ "./goChat" ]