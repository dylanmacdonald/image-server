FROM centurylink/ca-certs
WORKDIR /app

COPY image-service /app/
COPY images /app/images

ENTRYPOINT [ "./image-service" ]
