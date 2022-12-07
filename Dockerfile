FROM golang:1.19 as builder
WORKDIR /
COPY . ./lancktack
WORKDIR /lancktack
RUN CGO_ENABLED=0 GOOS=linux go build -a -o testapp
#CMD [ "./testapp" ]

FROM alpine:3.16
WORKDIR /lancktack
COPY --from=builder /lancktack/testapp .
COPY --from=builder /lancktack/examples.sh .
COPY --from=builder /lancktack/data ./data
RUN chmod +x examples.sh
CMD [ "/lancktack/examples.sh" ]