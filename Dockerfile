FROM golang:1.18-rc-bullseye

ARG version
ARG debrevision
ARG goarch
ARG debarch

ENV DEB_ARCH=${debarch}
ENV DEB_REVISION=${debrevision}
ENV GOARCH=${goarch}
ENV VERSION=${version}

RUN apt-get update
RUN apt-get install -y m4

WORKDIR /src

COPY .git/ ./.git/
COPY controller/ ./controller/
COPY cli/ ./cli/
COPY daemon/ ./daemon/
COPY DEBIAN/ ./DEBIAN/
COPY go.work .
COPY Makefile .

CMD ["sh", "-c", "make DEB_ARCH=${DEB_ARCH} GOARCH=${GOARCH} VERSION=${VERSION} DEB_REVISION=${DEB_REVISION} raspidoor_${VERSION}-${DEB_REVISION}_${DEB_ARCH}.deb && cp *.deb /out"]

VOLUME /out