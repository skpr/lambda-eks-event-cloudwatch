FROM golang:1.22 as build
ADD . /go/src/github.com/skpr/lambda-eks-event-cloudwatch
WORKDIR /go/src/github.com/skpr/lambda-eks-event-cloudwatch
RUN go fmt ./...
RUN go test ./...
RUN go build -tags lambda.norpc -o main main.go

FROM public.ecr.aws/lambda/provided:al2023
COPY --from=build /go/src/github.com/skpr/lambda-eks-event-cloudwatch/main ./main
ENTRYPOINT [ "./main" ]
