FROM alpine:latest

MAINTAINER James Allison <james.allison@monzo.com>

WORKDIR "/opt"

ADD .docker_build/monzo-heroku /opt/bin/monzo-heroku
ADD ./templates /opt/templates
ADD ./static /opt/static

CMD ["/opt/bin/monzo-heroku"]

