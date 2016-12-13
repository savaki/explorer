FROM ubuntu:16.04
MAINTAINER Matt Ho <matt.ho@gmail.com>

ADD explorer /bin/explorer

ENV PORT 80
EXPOSE 80
CMD [ "/bin/explorer" ]

