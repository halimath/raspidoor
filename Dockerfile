FROM golang:1.18-bullseye

ARG version
ARG debrevision

ENV DEB_REVISION=${debrevision}
ENV VERSION=${version}

RUN apt-get update
RUN apt-get install -y m4

WORKDIR /src

COPY .git/ ./.git/
COPY systemd/ ./systemd/
COPY controller/ ./controller/
COPY cli/ ./cli/
COPY daemon/ ./daemon/
COPY webapp/ ./webapp/
COPY DEBIAN/ ./DEBIAN/
COPY go.work .
COPY Makefile .

CMD [\
    "sh", \
    "-c", \
    "\
        make DEB_ARCH=armhf GOARCH=arm VERSION=${VERSION} DEB_REVISION=${DEB_REVISION} raspidoor_${VERSION}-${DEB_REVISION}_armhf.deb &&\
        cp *.deb /out &&\
        make DEB_ARCH=arm64 GOARCH=arm64 VERSION=${VERSION} DEB_REVISION=${DEB_REVISION} raspidoor_${VERSION}-${DEB_REVISION}_arm64.deb &&\
        cp *.deb /out\
    "\
]

VOLUME /out