FROM alpine
LABEL maintainer="ishmeetsinghis@gmail.com"
LABEL description="This is the Flooopy Burd"

#RUN apk update

RUN mkdir -p /home

#setup and run flooopy server
COPY . /home

#/home/bbsadmin/bin/
WORKDIR /home

EXPOSE 8080

ENTRYPOINT [ "/home/flooopy" ]
