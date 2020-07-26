FROM alpine
LABEL maintainer="ishmeetsinghis@gmail.com"
LABEL description="This is the Flooopy Burd"

RUN apk update && \
apk add --no-cache curl && \
apk add --no-cache ca-certificates && \
apk add --no-cache -u libcrypto1.1 && \
apk add --no-cache -u libssl1.1 && \
rm -rf /var/cache/apk/*


RUN mkdir -p /home

#setup and run dl rest server
COPY ./data/ /home
COPY ./bin/dl-rest /home

#/home/bbsadmin/bin/
WORKDIR /home

ENTRYPOINT [ "/home/dl-rest" ]
