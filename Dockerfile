FROM mischief/docker-golang
ENV HOME /root
ENV GOPATH /root
RUN apt-get update && apt-get install -y wget ca-certificates
RUN apt-mark hold initscripts udev plymouth mountall

# Add postgres repo and update
RUN echo "deb http://apt.postgresql.org/pub/repos/apt/ precise-pgdg main" > /etc/apt/sources.list.d/pg.list
RUN wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add -
RUN apt-get update
RUN apt-get upgrade -y

# Store postgres directories as environment variables
ENV DATA_DIR /var/lib/postgresql/9.3/main
ENV BIN_DIR /usr/lib/postgresql/9.3/bin
ENV CONF_DIR /etc/postgresql/9.3/main

RUN apt-get install -y git-core bzr daemontools
RUN apt-get install -y postgresql-9.3 libpq-dev
RUN apt-get install -y liblua5.1-dev

# Configure postgres to accept connections from everywhere
# and let it listen to all addresses of the container.
RUN echo "host all all 0.0.0.0/0 password" | tee -a $CONF_DIR/pg_hba.conf
RUN echo "listen_addresses='*'" | tee -a $CONF_DIR/postgresql.conf

# Setup monty db
RUN service postgresql restart &&\
	echo CREATE USER test WITH PASSWORD \'test\' | su -c psql postgres &&\
	echo CREATE DATABASE monty | su -c psql postgres &&\
	echo GRANT ALL PRIVILEGES ON DATABASE monty to test | su -c psql postgres

# Setup go libraries and app
ENV PATH $GOPATH/bin:$PATH
RUN go get launchpad.net/godeps
RUN mkdir -p $GOPATH/src/github.com/hackerlist/monty
ADD . $GOPATH/src/github.com/hackerlist/monty/
#RUN cd $GOPATH/src/github.com/hackerlist; git clone https://github.com/hackerlist/monty

WORKDIR /root/src/github.com/hackerlist/monty
RUN go get github.com/revel/cmd/revel
RUN go get -u ./...
RUN godeps -u dependencies.tsv
RUN godeps -u $GOPATH/src/github.com/hackerlist/monty/dependencies.tsv

# Setup daemontools to run postgres and monty
RUN mkdir -p /service
RUN mkdir /service/postgres
ADD docker/postgres.run /service/postgres/run
RUN chmod +x /service/postgres/run
RUN mkdir /service/monty
ADD docker/monty.run /service/monty/run
RUN chmod +x /service/monty/run

# Run daemontools by default
CMD /usr/bin/svscan /service

