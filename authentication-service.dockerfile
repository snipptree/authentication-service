FROM alpine:latest
WORKDIR /app
COPY authApp /app/authApp
EXPOSE 80
CMD ["./authApp"]
