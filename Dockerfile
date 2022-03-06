# TODO revisit this to reduce the size of docker image, We can run an remove installing golang
FROM golang:latest

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download
COPY . .
COPY ./entrypoint.sh /
RUN go build -o main .

EXPOSE 8080
CMD ["/entrypoint.sh"]