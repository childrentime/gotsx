# Build any of the demo apps (or your own gotsx app laid out the same way) into a ~15 MB image.
#   docker build --build-arg APP=shop -t gotsx-shop .
#   docker run -p 3000:3000 gotsx-shop
# The binary embeds the client runtime and public/ assets (go:embed), so the runtime image needs nothing else.
ARG APP=shop

FROM golang:1.26 AS build
ARG APP
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN go run ./cmd/gotsx tailwind --dir /src/.tools \
 && go run ./cmd/gotsx build ./$APP \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/app ./$APP

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/app /app
ENV PORT=3000
EXPOSE 3000
ENTRYPOINT ["/app", "-addr", ":3000"]
