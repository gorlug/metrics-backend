# https://gist.github.com/avishayp/33fcee06ee440524d21600e2e817b6b7
# non root user example for alpine
#
# usage:
# $ docker build --build-arg "USER=someuser" --tag test .
# $ docker run --rm test

FROM alpine

ARG USER=default
ENV HOME /home/$USER

RUN apk add --update libc6-compat

# add new user
RUN adduser -D $USER

USER $USER
WORKDIR $HOME

# files in /home/$USER to be owned by $USER
# docker has --chown flag for COPY, but it does not expand ENV so we fallback to:
# COPY src src
# RUN sudo chown -R $USER:$USER $HOME

COPY --chown=default:default metrics-backend .
COPY --chown=default:default .env .
COPY --chown=default:default views/ ./views/
CMD ./metrics-backend

