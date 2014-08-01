FROM mischief/docker-golang
ENV HOME /root
ENV GOPATH /root
RUN apt-get update && apt-get install -y wget ca-certificates
RUN apt-mark hold initscripts udev plymouth mountall

RUN apt-get install -y git-core mercurial bzr daemontools
RUN apt-get install -y liblua5.1-dev postgresql-client

# Setup go libraries and app
ENV PATH $GOPATH/bin:$PATH
RUN go get launchpad.net/godeps
RUN go get github.com/revel/cmd/revel

#RUN cd $GOPATH/src/github.com/hackerlist; git clone https://github.com/hackerlist/monty

RUN mkdir -p $GOPATH/src/github.com/hackerlist/monty
ADD . $GOPATH/src/github.com/hackerlist/monty/
WORKDIR /root/src/github.com/hackerlist/monty
RUN go get -u ./...
RUN godeps -u dependencies.tsv

# Setup daemontools to run postgres and monty
RUN mkdir -p /service
RUN mkdir /service/monty
ADD docker/monty.run /service/monty/run
RUN chmod +x /service/monty/run

# Run daemontools by default
CMD /usr/bin/svscan /service

