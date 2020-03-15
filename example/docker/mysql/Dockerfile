FROM mysql:5.7.29

RUN apt-get update \
    && apt-get install --no-install-recommends --no-install-suggests -y git \
    && rm -rf /var/lib/apt/lists/*

COPY mysql/my.cnf /etc/

RUN git config --global http.sslVerify false \
    && cd /tmp \
    && git clone https://github.com/datacharmer/test_db.git

