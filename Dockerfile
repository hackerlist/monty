FROM mischief/docker-golang
ENV HOME /root
RUN apt-get update && apt-get install -y wget ca-certificates
RUN apt-mark hold initscripts udev plymouth mountall

RUN echo "deb http://apt.postgresql.org/pub/repos/apt/ precise-pgdg main" > /etc/apt/sources.list.d/pg.list
RUN wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | sudo apt-key add -
RUN apt-get update
RUN apt-get upgrade -y

RUN apt-get install -y git-core bzr
RUN apt-get install -y postgresql-9.3 libpq-dev
RUN apt-get install -y liblua5.1-dev

RUN go get launchpad.net/godeps
RUN mkdir -p $GOPATH/src/github.com/hackerlist
#RUN cd $GOPATH/src/github.com/hackerlist; git clone https://github.com/hackerlist/monty
ADD . $GOPATH/src/github.com/hackerlist/monty/
RUN godeps -u $GOPATH/src/github.com/hackerlist/monty/dependencies.tsv
CMD revel run github.com/hackerlist/monty prod

